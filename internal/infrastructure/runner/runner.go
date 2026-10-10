// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"runtime/debug"
	"sync"
	"time"

	"github.com/telepresenceio/watchable"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
)

var tracer = otel.Tracer("envoy-gateway/infrastructure")

type Config struct {
	config.Server
	InfraIR      *message.InfraIR
	RunnerErrors *message.RunnerErrors
}

type Runner struct {
	Config
	mgr infrastructure.Manager

	// done tracks goroutines started by Start so that Close can wait for
	// them to exit before closing mgr, which they call into.
	done sync.WaitGroup

	// sem is held around the proxy infra calls of the subscription handler and
	// of retryProxyInfra, so a key is never applied by both at once. The Go
	// runtime hands it to blocked senders in the order they blocked, so neither
	// waits for more than one call of the other. Unlike a sync.Mutex, a blocked
	// send is durably blocking under testing/synctest, which the tests rely on.
	sem chan struct{}
	// failed holds, for each key whose last CreateOrUpdateProxyInfra failed
	// with a retryable error, what to retry and when. retries hands the keys
	// out to retryProxyInfra, backing off per key with backoff.
	failed  map[string]*failedInfra
	retries workqueue.TypedRateLimitingInterface[string]
	backoff workqueue.TypedRateLimiter[string]
}

// failedInfra is an Infra IR to retry once due, with a link to the span of the
// update that failed to apply it.
type failedInfra struct {
	infra *ir.Infra
	link  trace.Link
	due   time.Time
}

// Close implements Runner interface.
func (r *Runner) Close() error {
	r.done.Wait()
	return r.mgr.Close()
}

// Name implements Runner interface.
func (r *Runner) Name() string {
	return string(egv1a1.LogComponentInfrastructureRunner)
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

// Start starts the infrastructure runner
func (r *Runner) Start(ctx context.Context) (err error) {
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name())
	if r.EnvoyGateway.Provider.Type == egv1a1.ProviderTypeCustom &&
		r.EnvoyGateway.Provider.Custom.Infrastructure == nil {
		r.Logger.Info("provider is not specified, no infrastructure is available")
		return nil
	}
	errNotifier := message.RunnerErrorNotifier{RunnerName: r.Name(), RunnerErrors: r.RunnerErrors}
	r.mgr, err = infrastructure.NewManager(ctx, &r.Server, r.Logger, errNotifier)
	if err != nil {
		r.Logger.Error(err, "failed to create new manager")
		return err
	}

	// This is a blocking function that subscribes to the infraIR and initializes the infrastructure.
	subscribeInitInfra := func() {
		// Subscribe to InfraIR updates.
		sub := r.InfraIR.Subscribe(ctx)
		r.done.Go(func() {
			r.updateProxyInfraFromSubscription(ctx, sub)
		})

		// Create the shared ratelimit infra during startup.
		r.done.Go(func() {
			r.initializeRateLimitInfra(ctx)
		})

		r.Logger.Info("started")
		<-ctx.Done()
		r.Logger.Info("shutting down")
	}

	// When leader election is active, infrastructure initialization occurs only upon acquiring leadership
	// to avoid multiple EG instances processing envoy proxy infra resources.
	if r.EnvoyGateway.Provider.IsRunningOnKubernetes() &&
		!ptr.Deref(r.EnvoyGateway.Provider.GetKubernetesConfiguration().LeaderElection.Disable, false) {
		r.done.Go(func() {
			select {
			case <-ctx.Done():
				return
			case <-r.Elected:
				// As a leader EG instance subscribe to infraIR to initialize the infrastructure.
				subscribeInitInfra()
			}
		})
	} else {
		// Since leader election is disabled subscribe to infraIR to initialize the infrastructure.
		r.done.Go(subscribeInitInfra)
	}
	return err
}

func (r *Runner) updateProxyInfraFromSubscription(ctx context.Context, sub <-chan watchable.Snapshot[string, *message.InfraIRWithContext]) {
	// Retry a key after 1s, 2s, 4s, ... and then once a minute. The queue has
	// no name, so it registers no workqueue metrics.
	r.sem = make(chan struct{}, 1)
	r.failed = make(map[string]*failedInfra)
	r.backoff = workqueue.NewTypedItemExponentialFailureRateLimiter[string](time.Second, time.Minute)
	r.retries = workqueue.NewTypedRateLimitingQueue(r.backoff)
	defer r.retries.ShutDown()
	r.done.Go(func() { r.retryProxyInfra(ctx) })

	// Subscribe to resources
	message.HandleSubscription(
		r.Logger,
		message.Metadata{Runner: r.Name(), Message: message.InfraIRMessageName}, sub,
		func(update message.Update[string, *message.InfraIRWithContext], errChan chan error) {
			// Check if context is done before logging to avoid writing to test output after test completes
			select {
			case <-ctx.Done():
				return
			default:
			}

			parentCtx := update.Value.ParentContext(ctx)
			var startOpts []trace.SpanStartOption
			if !update.Delete && !update.Initial {
				parentCtx, startOpts = message.RecordQueueWait(parentCtx, tracer, r.Name(), update.Value.StoredAtTime())
			}

			traceCtx, span := tracer.Start(parentCtx, "InfrastructureRunner.updateProxyInfraFromSubscription", startOpts...)
			defer span.End()
			traceLogger := r.Logger.WithTrace(traceCtx)

			traceLogger.Info("received an update", "key", update.Key, "delete", update.Delete)
			message.PublishRunnerEventMetric(r.Name(), update.Delete)
			span.SetAttributes(
				attribute.String("infra-ir.key", update.Key),
				attribute.Bool("update.delete", update.Delete),
			)

			r.sem <- struct{}{}
			defer func() { <-r.sem }()
			// This update supersedes the Infra IR of any retry pending for its key.
			delete(r.failed, update.Key)
			r.retries.Forget(update.Key)

			var val *ir.Infra
			if update.Value != nil {
				val = update.Value.Infra
			}

			if update.Delete {
				if err := r.mgr.DeleteProxyInfra(traceCtx, val); err != nil {
					select {
					case <-ctx.Done():
						return
					default:
						traceLogger.Error(err, "failed to delete infra")
					}
					errChan <- err
				}
			} else {
				// Manage the proxy infra.
				// Skip creating or updating infra if the Infra IR without any listener.
				// e.g.https://github.com/envoyproxy/gateway/issues/3044 --- Invalid Listener
				//     https://github.com/envoyproxy/gateway/issues/7735 --- Invalid EnvoyProxy
				if len(val.Proxy.Listeners) == 0 {
					select {
					case <-ctx.Done():
						return
					default:
						traceLogger.Info("Infra IR was updated, but no listeners were found. Skipping infra creation.")
					}
					return
				}

				if err := r.mgr.CreateOrUpdateProxyInfra(traceCtx, val); err != nil {
					select {
					case <-ctx.Done():
						return
					default:
						traceLogger.Error(err, "failed to create new infra")
					}
					if isRetryable(ctx, err) {
						// val is shared with the watchable map, and the Manager
						// may modify the Infra IR it applies, so retry a copy.
						r.scheduleRetry(update.Key, &failedInfra{infra: val.DeepCopy(), link: trace.LinkFromContext(traceCtx)})
					}
					errChan <- err
				}
			}
		},
	)
	select {
	case <-ctx.Done():
		return
	default:
		r.Logger.Info("infra subscriber shutting down")
	}
}

// retryProxyInfra applies the failed Infra IR of each key handed out by
// r.retries again, until it succeeds, fails with an error that is not
// retryable, or a newer update for the key supersedes it.
func (r *Runner) retryProxyInfra(ctx context.Context) {
	for {
		key, shutdown := r.retries.Get()
		if shutdown {
			return
		}
		r.retryKey(ctx, key)
		r.retries.Done(key)
	}
}

// scheduleRetry records f as the Infra IR to retry for key, due after the
// next backoff of key.
func (r *Runner) scheduleRetry(key string, f *failedInfra) {
	delay := r.backoff.When(key)
	f.due = time.Now().Add(delay)
	r.failed[key] = f
	r.retries.AddAfter(key, delay)
}

// retryKey applies the failed Infra IR of key again, unless an update has
// superseded it since.
func (r *Runner) retryKey(ctx context.Context, key string) {
	r.sem <- struct{}{}
	defer func() { <-r.sem }()
	f, ok := r.failed[key]
	if !ok || ctx.Err() != nil {
		return
	}
	// The queue hands out a key early if it still held the key for an earlier
	// failure; wait for the backoff of the latest one.
	if wait := time.Until(f.due); wait > 0 {
		r.retries.AddAfter(key, wait)
		return
	}
	traceCtx, span := tracer.Start(ctx, "InfrastructureRunner.retryProxyInfra", trace.WithLinks(f.link))
	defer span.End()
	span.SetAttributes(attribute.String("infra-ir.key", key))
	logger := r.Logger.WithTrace(traceCtx)
	defer func() {
		// Like the handler, survive a panic of the Manager, but stop retrying.
		if p := recover(); p != nil {
			logger.Error(fmt.Errorf("%+v", p), "observed a panic", "key", key, "stackTrace", string(debug.Stack()))
			delete(r.failed, key)
			r.retries.Forget(key)
		}
	}()
	err := r.mgr.CreateOrUpdateProxyInfra(traceCtx, f.infra)
	switch {
	case err == nil:
		logger.Info("created infra on retry", "key", key, "retries", r.retries.NumRequeues(key))
	case ctx.Err() == nil:
		logger.Error(err, "failed to retry creating infra", "key", key, "retries", r.retries.NumRequeues(key))
	}
	if isRetryable(ctx, err) {
		r.scheduleRetry(key, f)
	} else {
		delete(r.failed, key)
		r.retries.Forget(key)
	}
}

// isRetryable reports whether a proxy infra call that failed with err may
// succeed if it is retried while the runner is still running: the API server
// answered that it was briefly unavailable or overloaded (the reasons the
// Kubernetes provider's isTransientError also treats as transient), or no
// answer came back because the request failed in transport (a *url.Error).
func isRetryable(ctx context.Context, err error) bool {
	return err != nil && ctx.Err() == nil &&
		(kerrors.IsServerTimeout(err) || kerrors.IsTimeout(err) || kerrors.IsTooManyRequests(err) ||
			kerrors.IsServiceUnavailable(err) || kerrors.IsStoreReadError(err) || kerrors.IsInternalError(err) ||
			kerrors.IsUnexpectedServerError(err) || errors.Is(err, context.DeadlineExceeded) ||
			errors.As(err, new(*url.Error)))
}

func (r *Runner) initializeRateLimitInfra(ctx context.Context) {
	if !r.waitForProviderReady(ctx) {
		return
	}

	if r.EnvoyGateway.RateLimit != nil {
		if err := r.mgr.CreateOrUpdateRateLimitInfra(ctx); err != nil {
			r.Logger.Error(err, "failed to create ratelimit infra")
		}
		return
	}

	if err := r.mgr.DeleteRateLimitInfra(ctx); err != nil {
		r.Logger.Error(err, "failed to delete ratelimit infra")
	}
}

func (r *Runner) waitForProviderReady(ctx context.Context) bool {
	if !r.EnvoyGateway.Provider.IsRunningOnKubernetes() {
		return true
	}

	select {
	case <-ctx.Done():
		return false
	case <-r.ProviderReady:
		return true
	}
}
