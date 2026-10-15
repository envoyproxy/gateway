// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gateway

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-redis/redis/v7"
)

// reloadChannel is the Redis pub/sub channel used to broadcast route-config
// changes to every Envoy Gateway task sharing the routes volume.
const reloadChannel = "gateway:route-reload"

// Redis wraps a Redis client used to broadcast and receive reload signals.
type Redis struct {
	client *redis.Client
	logger logr.Logger
}

// InitRedis connects to Redis using the given URL (e.g. redis://host:6379/0).
func InitRedis(url string, logger logr.Logger) (*Redis, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}
	client := redis.NewClient(opt)
	if err := client.Ping().Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}
	return &Redis{client: client, logger: logger}, nil
}

// Publish broadcasts a reload signal carrying the cluster name as its payload.
// Failures are logged but not returned, since the local file write already
// succeeded and other tasks will recover on the next publish or manual reload.
func (r *Redis) Publish(cluster string) {
	if err := r.client.Publish(reloadChannel, cluster).Err(); err != nil {
		r.logger.Error(err, "failed to publish route reload signal", "cluster", cluster)
		return
	}
	r.logger.Info("published route reload signal", "cluster", cluster)
}

// StartSubscriber subscribes to the reload channel and invokes trigger for each
// message received, until ctx is cancelled. It runs in its own goroutine.
func (r *Redis) StartSubscriber(ctx context.Context, trigger func()) {
	pubsub := r.client.Subscribe(reloadChannel)
	ch := pubsub.Channel()
	r.logger.Info("subscribed to route reload channel", "channel", reloadChannel)

	go func() {
		defer func() { _ = pubsub.Close() }()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				r.logger.Info("received route reload signal", "cluster", msg.Payload)
				trigger()
			}
		}
	}()
}

// Close releases the Redis connection.
func (r *Redis) Close() error {
	return r.client.Close()
}
