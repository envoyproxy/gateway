---
title: "Watching Components Design"
weight: 3
---

Envoy Gateway is made up of several components that communicate in-process.  Some of them (namely Providers) watch
external resources, and "publish" what they see for other components to consume; others watch what another publishes and
act on it (such as the resource translator watches what the providers publish, and then publishes its own results that
are watched by another component).  Some of these internally published results are consumed by multiple components.

To facilitate this communication use the [watchable][] library.  The `watchable.Map` type is very similar to the
standard library's `sync.Map` type, but supports a `.Subscribe` (and `.SubscribeSubset`) method that promotes a pub/sub
pattern.

## Pub

Many of the things we communicate around are naturally named, either by a bare "name" string or by a "name"/"namespace"
tuple.  And because `watchable.Map` is typed, it makes sense to have one map for each type of thing (very similar to if
we were using native Go `map`s).  For example, a struct that might be written to by the Kubernetes provider, and read by
the IR translator:

   ```go
   type ResourceTable struct {
       // gateway classes are cluster-scoped; no namespace
       GatewayClasses watchable.Map[string, *gwapiv1.GatewayClass]

       // gateways are namespace-scoped, so use a k8s.io/apimachinery/pkg/types.NamespacedName as the map key.
       Gateways watchable.Map[types.NamespacedName, *gwapiv1.Gateway]

       HTTPRoutes watchable.Map[types.NamespacedName, *gwapiv1.HTTPRoute]
   }
   ```

The Kubernetes provider updates the table by calling `table.Thing.Store(name, val)` and `table.Thing.Delete(name)`;
updating a map key with a value that is deep-equal (usually `reflect.DeepEqual`, but you can implement your own `.Equal`
method) the current value is a no-op; it won't trigger an event for subscribers.  This is handy so that the publisher
doesn't have as much state to keep track of; it doesn't need to know "did I already publish this thing", it can just
`.Store` its data and `watchable` will do the right thing.

## Sub

Meanwhile, the translator and other interested components subscribe to it with `table.Thing.Subscribe` (or
`table.Thing.SubscribeSubset` if they only care about a few "Thing"s).  So the translator goroutine might look like:

   ```go
   func(ctx context.Context) error {
       for snapshot := range k8sTable.HTTPRoutes.Subscribe(ctx) {
           fullState := irInput{
              GatewayClasses: k8sTable.GatewayClasses.LoadAll(),
              Gateways:       k8sTable.Gateways.LoadAll(),
              HTTPRoutes:     snapshot.State,
           }
           translate(irInput)
       }
   }
   ```

Or, to watch multiple maps in the same loop:

   ```go
   func worker(ctx context.Context) error {
       classCh := k8sTable.GatewayClasses.Subscribe(ctx)
       gwCh := k8sTable.Gateways.Subscribe(ctx)
       routeCh := k8sTable.HTTPRoutes.Subscribe(ctx)
       for ctx.Err() == nil {
           var arg irInput
           select {
           case snapshot := <-classCh:
               arg.GatewayClasses = snapshot.State
           case snapshot := <-gwCh:
               arg.Gateways = snapshot.State
           case snapshot := <-routeCh:
               arg.Routes = snapshot.State
           }
           if arg.GateWayClasses == nil {
               arg.GatewayClasses = k8sTable.GateWayClasses.LoadAll()
           }
           if arg.GateWays == nil {
               arg.Gateways = k8sTable.GateWays.LoadAll()
           }
           if arg.HTTPRoutes == nil {
               arg.HTTPRoutes = k8sTable.HTTPRoutes.LoadAll()
           }
           translate(irInput)
       }
   }
   ```

From the updates it gets from `.Subscribe`, it can get a full view of the map being subscribed to via `snapshot.State`;
but it must read the other maps explicitly.  Like `sync.Map`, `watchable.Map`s are thread-safe; while `.Subscribe` is a
handy way to know when to run, `.Load` and friends can be used without subscribing.

There can be any number of subscribers.  For that matter, there can be any number of publishers `.Store`ing things, but
it's probably wise to just have one publisher for each map.

The channel returned from `.Subscribe` **is immediately readable** with a snapshot of the map as it existed when
`.Subscribe` was called; and becomes readable again whenever `.Store` or `.Delete` mutates the map.  If multiple
mutations happen between reads (or if mutations happen between `.Subscribe` and the first read), they are coalesced in
to one snapshot to be read; the `snapshot.State` is the most-recent full state, and `snapshot.Updates` is a listing of
each of the mutations that cause this snapshot to be different than the last-read one.  This way subscribers don't need
to worry about a backlog accumulating if they can't keep up with the rate of changes from the publisher.

If the map contains anything before `.Subscribe` is called, that very first read won't include `snapshot.Updates`
entries for those pre-existing items; if you are working with `snapshot.Update` instead of `snapshot.State`, then you
must add special handling for your first read.  We have a utility function `./internal/message.HandleSubscription` to
help with this.

## Other Notes

The common pattern will likely be that the entrypoint that launches the goroutines for each component instantiates the
map, and passes them to the appropriate publishers and subscribers; same as if they were communicating via a dumb
`chan`.

A limitation of `watchable.Map` is that in order to ensure safety between goroutines, it does require that value types
be deep-copiable; either by having a `DeepCopy` method, being a `proto.Message`, or by containing no reference types and
so can be deep-copied by naive assignment.  Fortunately, we're using `controller-gen` anyway, and `controller-gen` can
generate `DeepCopy` methods for us: just stick a `// +k8s:deepcopy-gen=true` on the types that you want it to generate
methods for.

[watchable]: https://pkg.go.dev/github.com/telepresenceio/watchable
