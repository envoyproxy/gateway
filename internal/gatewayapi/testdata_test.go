package gatewayapi

const BasicHTTPRouteAttachingToGatewayIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        allowedRoutes:
          namespaces:
            from: All
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
`

const BasicHTTPRouteAttachingToGatewayOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          allowedRoutes:
            namespaces:
              from: All
    status:
      listeners:
        - name: http
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 1
          conditions:
            - type: Ready
              status: "True"
              reason: Ready
              message: Listener is ready
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "True"
              reason: Accepted
              message: Route is accepted
ir:
  name: ""
  http:
    - name: envoy-gateway-gateway-1-http
      address: 0.0.0.0
      port: 80
      routes:
        - name: ""
          pathMatch:
            prefix: "/"
          destinations:
            - host: 7.7.7.7
              port: 8080
              weight: 1
`

const BasicHTTPRouteAttachingToListenerIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        allowedRoutes:
          namespaces:
            from: All
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
        sectionName: http
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
`

const BasicHTTPRouteAttachingToListenerOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          allowedRoutes:
            namespaces:
              from: All
    status:
      listeners:
        - name: http
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 1
          conditions:
            - type: Ready
              status: "True"
              reason: Ready
              message: Listener is ready
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
          sectionName: http
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
            sectionName: http
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "True"
              reason: Accepted
              message: Route is accepted
ir:
  name: ""
  http:
    - name: envoy-gateway-gateway-1-http
      address: 0.0.0.0
      port: 80
      routes:
        - name: ""
          pathMatch:
            prefix: "/"
          destinations:
            - host: 7.7.7.7
              port: 8080
              weight: 1
`

const GatewayAllowsSameNamespaceWithAllowedHTTPRouteIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        allowedRoutes:
          namespaces:
            from: Same
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: envoy-gateway
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
services:
- apiVersion: v1
  kind: Service
  metadata:
    namespace: envoy-gateway
    name: service-1
  spec:
    clusterIP: 7.7.7.7
    ports:
    - port: 8080
`

const GatewayAllowsSameNamespaceWithAllowedHTTPRouteOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          allowedRoutes:
            namespaces:
              from: Same
    status:
      listeners:
        - name: http
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 1
          conditions:
            - type: Ready
              status: "True"
              reason: Ready
              message: Listener is ready
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: envoy-gateway
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "True"
              reason: Accepted
              message: Route is accepted
ir:
  name: ""
  http:
    - name: envoy-gateway-gateway-1-http
      address: 0.0.0.0
      port: 80
      routes:
        - name: ""
          pathMatch:
            prefix: "/"
          destinations:
            - host: 7.7.7.7
              port: 8080
              weight: 1
`

const GatewayAllowsSameNamespaceWithDisallowedHTTPRouteIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        allowedRoutes:
          namespaces:
            from: Same
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
`

const GatewayAllowsSameNamespaceWithDisallowedHTTPRouteOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          allowedRoutes:
            namespaces:
              from: Same
    status:
      listeners:
        - name: http
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 0
          conditions:
            - type: Ready
              status: "True"
              reason: Ready
              message: Listener is ready
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "False"
              reason: NotAllowedByListeners
              message: No listeners included by this parent ref allowed this attachment.
ir:
  name: ""
  http:
    - name: envoy-gateway-gateway-1-http
      address: 0.0.0.0
      port: 80
`

const HTTPRouteAttachingToGatewayWithTwoListenersIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http-1
        protocol: HTTP
        port: 80
        hostname: foo.com
        allowedRoutes:
          namespaces:
            from: All
      - name: http-2
        protocol: HTTP
        port: 80
        hostname: bar.com
        allowedRoutes:
          namespaces:
            from: All
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
`

const HTTPRouteAttachingToGatewayWithTwoListenersOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http-1
          protocol: HTTP
          port: 80
          hostname: foo.com
          allowedRoutes:
            namespaces:
              from: All
        - name: http-2
          protocol: HTTP
          port: 80
          hostname: bar.com
          allowedRoutes:
            namespaces:
              from: All
    status:
      listeners:
        - name: http-1
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 1
          conditions:
            - type: Ready
              status: "True"
              reason: Ready
              message: Listener is ready
        - name: http-2
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 1
          conditions:
            - type: Ready
              status: "True"
              reason: Ready
              message: Listener is ready
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "True"
              reason: Accepted
              message: Route is accepted
ir:
  name: ""
  http:
    - name: envoy-gateway-gateway-1-http-1
      address: 0.0.0.0
      port: 80
      hostnames:
        - foo.com
      routes:
        - name: ""
          pathMatch:
            prefix: "/"
          destinations:
            - host: 7.7.7.7
              port: 8080
              weight: 1
    - name: envoy-gateway-gateway-1-http-2
      address: 0.0.0.0
      port: 80
      hostnames:
        - bar.com
      routes:
        - name: ""
          pathMatch:
            prefix: "/"
          destinations:
            - host: 7.7.7.7
              port: 8080
              weight: 1
`

const HTTPRouteAttachingToListenerOnGatewayWithTwoListenersIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http-1
        protocol: HTTP
        port: 80
        hostname: foo.com
        allowedRoutes:
          namespaces:
            from: All
      - name: http-2
        protocol: HTTP
        port: 80
        hostname: bar.com
        allowedRoutes:
          namespaces:
            from: All
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
        sectionName: http-2
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
`

const HTTPRouteAttachingToListenerOnGatewayWithTwoListenersOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http-1
          protocol: HTTP
          port: 80
          hostname: foo.com
          allowedRoutes:
            namespaces:
              from: All
        - name: http-2
          protocol: HTTP
          port: 80
          hostname: bar.com
          allowedRoutes:
            namespaces:
              from: All
    status:
      listeners:
        - name: http-1
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 0
          conditions:
            - type: Ready
              status: "True"
              reason: Ready
              message: Listener is ready
        - name: http-2
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 1
          conditions:
            - type: Ready
              status: "True"
              reason: Ready
              message: Listener is ready
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
          sectionName: http-2
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
            sectionName: http-2
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "True"
              reason: Accepted
              message: Route is accepted
ir:
  name: ""
  http:
    - name: envoy-gateway-gateway-1-http-1
      address: 0.0.0.0
      port: 80
      hostnames:
        - foo.com
    - name: envoy-gateway-gateway-1-http-2
      address: 0.0.0.0
      port: 80
      hostnames:
        - bar.com
      routes:
        - name: ""
          pathMatch:
            prefix: "/"
          destinations:
            - host: 7.7.7.7
              port: 8080
              weight: 1
`

const HTTPRouteWithSpecificHostnameAttachingToGatewayWithWildcardHostnameIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        hostname: "*.envoyproxy.io"
        allowedRoutes:
          namespaces:
            from: All
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    hostnames:
      - gateway.envoyproxy.io
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
`

const HTTPRouteWithSpecificHostnameAttachingToGatewayWithWildcardHostnameOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          hostname: "*.envoyproxy.io"
          allowedRoutes:
            namespaces:
              from: All
    status:
      listeners:
        - name: http
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 1
          conditions:
            - type: Ready
              status: "True"
              reason: Ready
              message: Listener is ready
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      hostnames:
        - gateway.envoyproxy.io
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "True"
              reason: Accepted
              message: Route is accepted
ir:
  name: ""
  http:
    - name: envoy-gateway-gateway-1-http
      address: 0.0.0.0
      port: 80
      hostnames:
        - "*.envoyproxy.io"
      routes:
        - name: ""
          pathMatch:
            prefix: "/"
          headerMatches:
            - name: ":authority"
              exact: gateway.envoyproxy.io
          destinations:
            - host: 7.7.7.7
              port: 8080
              weight: 1
`

const HTTPRouteWithTwoSpecificHostnamesAttachingToGatewayWithWildcardHostnameIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        hostname: "*.envoyproxy.io"
        allowedRoutes:
          namespaces:
            from: All
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    hostnames:
      - gateway.envoyproxy.io
      - whales.envoyproxy.io
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
`

const HTTPRouteWithTwoSpecificHostnamesAttachingToGatewayWithWildcardHostnameOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          hostname: "*.envoyproxy.io"
          allowedRoutes:
            namespaces:
              from: All
    status:
      listeners:
        - name: http
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 1
          conditions:
            - type: Ready
              status: "True"
              reason: Ready
              message: Listener is ready
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      hostnames:
        - gateway.envoyproxy.io
        - whales.envoyproxy.io
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "True"
              reason: Accepted
              message: Route is accepted
ir:
  name: ""
  http:
    - name: envoy-gateway-gateway-1-http
      address: 0.0.0.0
      port: 80
      hostnames:
        - "*.envoyproxy.io"
      routes:
        - name: ""
          pathMatch:
            prefix: "/"
          headerMatches:
            - name: ":authority"
              exact: gateway.envoyproxy.io
          destinations:
            - host: 7.7.7.7
              port: 8080
              weight: 1
        - name: ""
          pathMatch:
            prefix: "/"
          headerMatches:
            - name: ":authority"
              exact: whales.envoyproxy.io
          destinations:
            - host: 7.7.7.7
              port: 8080
              weight: 1
`

const HTTPRouteWithNonMatchingSpecificHostnameAttachingToGatewayWithWildcardHostnameIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        hostname: "*.envoyproxy.io"
        allowedRoutes:
          namespaces:
            from: All
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    hostnames:
      - whales.kubernetes.io
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
`

const HTTPRouteWithNonMatchingSpecificHostnameAttachingToGatewayWithWildcardHostnameOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          hostname: "*.envoyproxy.io"
          allowedRoutes:
            namespaces:
              from: All
    status:
      listeners:
        - name: http
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 0
          conditions:
            - type: Ready
              status: "True"
              reason: Ready
              message: Listener is ready
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      hostnames:
        - whales.kubernetes.io
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "False"
              reason: NoMatchingListenerHostname
              message: There were no hostname intersections between the HTTPRoute and this parent ref's Listener(s).
ir:
  name: ""
  http:
    - name: envoy-gateway-gateway-1-http
      address: 0.0.0.0
      port: 80
      hostnames:
        - "*.envoyproxy.io"
`

const GatewayWithListenerWithNonHTTPProtocolIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTPS
        port: 80
        allowedRoutes:
          namespaces:
            from: All
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
`

const GatewayWithListenerWithNonHTTPProtocolOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTPS
          port: 80
          allowedRoutes:
            namespaces:
              from: All
    status:
      listeners:
        - name: http
          attachedRoutes: 0
          conditions:
            - type: Detached
              status: "True"
              reason: UnsupportedProtocol
              message: Protocol must be HTTP
            - type: Ready
              status: "False"
              reason: Invalid
              message: Listener is invalid, see other Conditions for details.
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "False"
              reason: NoReadyListeners
              message: There are no ready listeners for this parent ref
ir:
  name: ""
`

const GatewayWithListenerWithMissingAllowedNamespacesSelectorIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        allowedRoutes:
          namespaces:
            from: Selector
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
`

const GatewayWithListenerWithMissingAllowedNamespacesSelectorOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          allowedRoutes:
            namespaces:
              from: Selector
    status:
      listeners:
        - name: http
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 0
          conditions:
            - type: Ready
              status: "False"
              reason: Invalid
              message: The allowedRoutes.namespaces.selector field must be specified when allowedRoutes.namespaces.from is set to "Selector".
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "False"
              reason: NoReadyListeners
              message: There are no ready listeners for this parent ref
ir:
  name: ""
`

const GatewayWithListenerWithInvalidAllowedNamespacesSelectorIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        allowedRoutes:
          namespaces:
            from: Selector
            selector:
              matchExpressions:
                - key: foo
                  operator: Exists
                  values:
                    - bar
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
`

const GatewayWithListenerWithInvalidAllowedNamespacesSelectorOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          allowedRoutes:
            namespaces:
              from: Selector
              selector:
                matchExpressions:
                  - key: foo
                    operator: Exists
                    values:
                      - bar
    status:
      listeners:
        - name: http
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 0
          conditions:
            - type: Ready
              status: "False"
              reason: Invalid
              message: "The allowedRoutes.namespaces.selector could not be parsed: values: Invalid value: []string{\"bar\"}: values set must be empty for exists and does not exist."
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "False"
              reason: NoReadyListeners
              message: There are no ready listeners for this parent ref
ir:
  name: ""
`

const GatewayWithListenerWithInvalidAllowedRoutesGroupIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        allowedRoutes:
          namespaces:
            from: All
          kinds:
            - group: foo.io
              kind: HTTPRoute
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
`

const GatewayWithListenerWithInvalidAllowedRoutesGroupOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          allowedRoutes:
            namespaces:
              from: All
            kinds:
              - group: foo.io
                kind: HTTPRoute
    status:
      listeners:
        - name: http
          attachedRoutes: 0
          conditions:
            - type: ResolvedRefs
              status: "False"
              reason: InvalidRouteKinds
              message: "Group is not supported, group must be gateway.networking.k8s.io"
            - type: Ready
              status: "False"
              reason: Invalid
              message: Listener is invalid, see other Conditions for details.
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "False"
              reason: NoReadyListeners
              message: There are no ready listeners for this parent ref
ir:
  name: ""
`

const GatewayWithListenerWithInvalidAllowedRoutesKindIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        allowedRoutes:
          namespaces:
            from: All
          kinds:
            - group: gateway.networking.k8s.io
              kind: FooRoute
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
`

const GatewayWithListenerWithInvalidAllowedRoutesKindOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          allowedRoutes:
            namespaces:
              from: All
            kinds:
              - group: gateway.networking.k8s.io
                kind: FooRoute
    status:
      listeners:
        - name: http
          attachedRoutes: 0
          conditions:
            - type: ResolvedRefs
              status: "False"
              reason: InvalidRouteKinds
              message: "Kind is not supported, kind must be HTTPRoute"
            - type: Ready
              status: "False"
              reason: Invalid
              message: Listener is invalid, see other Conditions for details.
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "False"
              reason: NoReadyListeners
              message: There are no ready listeners for this parent ref
ir:
  name: ""
`

const GatewayWithTwoListenersWithSamePortAndHostnameIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http-1
        protocol: HTTP
        port: 80
        hostname: foo.com
        allowedRoutes:
          namespaces:
            from: All
      - name: http-2
        protocol: HTTP
        port: 80
        hostname: foo.com
        allowedRoutes:
          namespaces:
            from: All
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
`

const GatewayWithTwoListenersWithSamePortAndHostnameOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http-1
          protocol: HTTP
          port: 80
          hostname: foo.com
          allowedRoutes:
            namespaces:
              from: All
        - name: http-2
          protocol: HTTP
          port: 80
          hostname: foo.com
          allowedRoutes:
            namespaces:
              from: All
    status:
      listeners:
        - name: http-1
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          conditions:
            - type: Conflicted
              status: "True"
              reason: HostnameConflict
              message: All listeners for a given port must use a unique hostname
            - type: Ready
              status: "False"
              reason: Invalid
              message: Listener is invalid, see other Conditions for details.
        - name: http-2
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          conditions:
            - type: Conflicted
              status: "True"
              reason: HostnameConflict
              message: All listeners for a given port must use a unique hostname
            - type: Ready
              status: "False"
              reason: Invalid
              message: Listener is invalid, see other Conditions for details.
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "False"
              reason: NoReadyListeners
              message: There are no ready listeners for this parent ref
ir:
  name: ""
  http:
`

const GatewayWithTwoListenersWithSamePortAndIncompatibleProtocolsIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http-1
        protocol: HTTP
        port: 80
        hostname: foo.com
        allowedRoutes:
          namespaces:
            from: All
      - name: http-2
        protocol: HTTPS
        port: 80
        hostname: bar.com
        allowedRoutes:
          namespaces:
            from: All
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
`

const GatewayWithTwoListenersWithSamePortAndIncompatibleProtocolsOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http-1
          protocol: HTTP
          port: 80
          hostname: foo.com
          allowedRoutes:
            namespaces:
              from: All
        - name: http-2
          protocol: HTTPS
          port: 80
          hostname: bar.com
          allowedRoutes:
            namespaces:
              from: All
    status:
      listeners:
        - name: http-1
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          conditions:
            - type: Conflicted
              status: "True"
              reason: ProtocolConflict
              message: All listeners for a given port must use a compatible protocol
            - type: Ready
              status: "False"
              reason: Invalid
              message: Listener is invalid, see other Conditions for details.
        - name: http-2
          conditions:
            - type: Conflicted
              status: "True"
              reason: ProtocolConflict
              message: All listeners for a given port must use a compatible protocol
            - type: Detached
              status: "True"
              reason: UnsupportedProtocol
              message: Protocol must be HTTP
            - type: Ready
              status: "False"
              reason: Invalid
              message: Listener is invalid, see other Conditions for details.
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      rules:
        - matches:
            - path:
                value: "/"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "False"
              reason: NoReadyListeners
              message: There are no ready listeners for this parent ref
ir:
  name: ""
  http:
`

const HTTPRouteWithSingleRuleWithPathPrefixAndExactHeaderMatchesIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        allowedRoutes:
          namespaces:
            from: All
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    rules:
      - matches:
          - path:
              value: "/pathprefix"
            headers:
              - name: Header-1
                value: Val-1
              - name: Header-2
                value: Val-2
        backendRefs:
          - name: service-1
            port: 8080
`

const HTTPRouteWithSingleRuleWithPathPrefixAndExactHeaderMatchesOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          allowedRoutes:
            namespaces:
              from: All
    status:
      listeners:
        - name: http
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 1
          conditions:
            - type: Ready
              status: "True"
              reason: Ready
              message: Listener is ready
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      rules:
        - matches:
          - path:
              value: "/pathprefix"
            headers:
              - name: Header-1
                value: Val-1
              - name: Header-2
                value: Val-2
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "True"
              reason: Accepted
              message: Route is accepted
ir:
  name: ""
  http:
    - name: envoy-gateway-gateway-1-http
      address: 0.0.0.0
      port: 80
      routes:
        - name: ""
          pathMatch:
            prefix: "/pathprefix"
          headerMatches:
            - name: Header-1
              exact: Val-1
            - name: Header-2
              exact: Val-2
          destinations:
            - host: 7.7.7.7
              port: 8080
              weight: 1
`

const HTTPRouteWithSingleRuleWithExactPathMatchIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        allowedRoutes:
          namespaces:
            from: All
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    rules:
      - matches:
          - path:
              type: Exact
              value: "/exact"
        backendRefs:
          - name: service-1
            port: 8080
`

const HTTPRouteWithSingleRuleWithExactPathMatchOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          allowedRoutes:
            namespaces:
              from: All
    status:
      listeners:
        - name: http
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 1
          conditions:
            - type: Ready
              status: "True"
              reason: Ready
              message: Listener is ready
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      rules:
        - matches:
          - path:
              type: Exact
              value: "/exact"
          backendRefs:
            - name: service-1
              port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "True"
              reason: Accepted
              message: Route is accepted
ir:
  name: ""
  http:
    - name: envoy-gateway-gateway-1-http
      address: 0.0.0.0
      port: 80
      routes:
        - name: ""
          pathMatch:
            exact: "/exact"
          destinations:
            - host: 7.7.7.7
              port: 8080
              weight: 1
`

const HTTPRouteRuleWithMultipleBackendsAndNoWeightsIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        allowedRoutes:
          namespaces:
            from: All
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
          - name: service-2
            port: 8080
          - name: service-3
            port: 8080
`

const HTTPRouteRuleWithMultipleBackendsAndNoWeightsOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          allowedRoutes:
            namespaces:
              from: All
    status:
      listeners:
        - name: http
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 1
          conditions:
            - type: Ready
              status: "True"
              reason: Ready
              message: Listener is ready
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
          - name: service-2
            port: 8080
          - name: service-3
            port: 8080
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "True"
              reason: Accepted
              message: Route is accepted
ir:
  name: ""
  http:
    - name: envoy-gateway-gateway-1-http
      address: 0.0.0.0
      port: 80
      routes:
        - name: ""
          pathMatch:
            prefix: "/"
          destinations:
            - host: 7.7.7.7
              port: 8080
              weight: 1
            - host: 7.7.7.7
              port: 8080
              weight: 1
            - host: 7.7.7.7
              port: 8080
              weight: 1
`

const HTTPRouteRuleWithMultipleBackendsAndWeightsIn = `
gateways:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        allowedRoutes:
          namespaces:
            from: All
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: httproute-1
  spec:
    parentRefs:
      - namespace: envoy-gateway
        name: gateway-1
    rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
            weight: 1
          - name: service-2
            port: 8080
            weight: 2
          - name: service-3
            port: 8080
            weight: 3
`

const HTTPRouteRuleWithMultipleBackendsAndWeightsOut = `
gateways:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: Gateway
    metadata:
      namespace: envoy-gateway
      name: gateway-1
    spec:
      gatewayClassName: envoy-gateway-class
      listeners:
        - name: http
          protocol: HTTP
          port: 80
          allowedRoutes:
            namespaces:
              from: All
    status:
      listeners:
        - name: http
          supportedKinds:
            - group: gateway.networking.k8s.io
              kind: HTTPRoute
          attachedRoutes: 1
          conditions:
            - type: Ready
              status: "True"
              reason: Ready
              message: Listener is ready
httpRoutes:
  - apiVersion: gateway.networking.k8s.io/v1beta1
    kind: HTTPRoute
    metadata:
      namespace: default
      name: httproute-1
    spec:
      parentRefs:
        - namespace: envoy-gateway
          name: gateway-1
      rules:
      - matches:
          - path:
              value: "/"
        backendRefs:
          - name: service-1
            port: 8080
            weight: 1
          - name: service-2
            port: 8080
            weight: 2
          - name: service-3
            port: 8080
            weight: 3
    status:
      parents:
        - parentRef:
            namespace: envoy-gateway
            name: gateway-1
          # controllerName: envoyproxy.io/gateway-controller
          conditions:
            - type: Accepted
              status: "True"
              reason: Accepted
              message: Route is accepted
ir:
  name: ""
  http:
    - name: envoy-gateway-gateway-1-http
      address: 0.0.0.0
      port: 80
      routes:
        - name: ""
          pathMatch:
            prefix: "/"
          destinations:
            - host: 7.7.7.7
              port: 8080
              weight: 1
            - host: 7.7.7.7
              port: 8080
              weight: 2
            - host: 7.7.7.7
              port: 8080
              weight: 3
`
