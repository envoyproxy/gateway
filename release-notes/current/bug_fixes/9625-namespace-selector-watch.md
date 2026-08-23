Fixed HTTPRoutes (and other routes) staying `NotAllowedByListeners` after a
namespace was labeled to match a Gateway listener's
`allowedRoutes.namespaces.selector`. The controller had no watch on `Namespace`
objects, so label-only changes never triggered a reconcile; the fix now watches
`Namespace` label changes and re-evaluates affected Gateways and routes.
