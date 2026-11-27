Fixed OIDC login looping when `redirectURL` has a query string: the redirect path matcher now uses the path of `redirectURL` without its query and fragment, the part Envoy matches against.
