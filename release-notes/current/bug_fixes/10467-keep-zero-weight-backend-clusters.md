Fixed zero-weight HTTPRoute and GRPCRoute backendRefs being dropped from xDS, so their clusters now exist (with a zero-weight route entry) before a canary rollout first shifts traffic to them.
