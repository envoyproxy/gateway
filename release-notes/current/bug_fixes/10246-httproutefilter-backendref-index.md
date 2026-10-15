Fixed an HTTPRoute not being reconciled when an `HTTPRouteFilter` that it references only from backendRef filters is created or updated, if the rule has no rule-level filters.
