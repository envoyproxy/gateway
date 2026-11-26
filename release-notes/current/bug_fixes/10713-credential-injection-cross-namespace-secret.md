Fixed HTTPRouteFilter credential injection ignoring the namespace of its Secret reference, so a Secret in another namespace that a ReferenceGrant allows is now loaded and the route is accepted.
