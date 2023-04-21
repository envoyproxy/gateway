# config.gateway.envoyproxy.io/v1alpha1
<p>Packages:</p>
<ul class="simple">
<li>
<a href="#config.gateway.envoyproxy.io%2fv1alpha1">config.gateway.envoyproxy.io/v1alpha1</a>
</li>
</ul>
<h2 id="config.gateway.envoyproxy.io/v1alpha1">config.gateway.envoyproxy.io/v1alpha1</h2>
<p>Package v1alpha1 contains API schema definitions for the config.gateway.envoyproxy.io
API group.</p>
Resource Types:
<ul class="simple"></ul>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.EnvoyGateway">EnvoyGateway
</h3>
<p>EnvoyGateway is the schema for the envoygateways API.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>EnvoyGatewaySpec</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewaySpec">
EnvoyGatewaySpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>EnvoyGatewaySpec</code> are embedded into this type.)
</p>
<p>EnvoyGatewaySpec defines the desired state of EnvoyGateway.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewayFileProvider">EnvoyGatewayFileProvider
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewayProvider">EnvoyGatewayProvider</a>)
</p>
<p>EnvoyGatewayFileProvider defines configuration for the File provider.</p>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewayKubernetesProvider">EnvoyGatewayKubernetesProvider
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewayProvider">EnvoyGatewayProvider</a>)
</p>
<p>EnvoyGatewayKubernetesProvider defines configuration for the Kubernetes provider.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>rateLimitDeployment</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.KubernetesDeploymentSpec">
KubernetesDeploymentSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RateLimitDeployment defines the desired state of the Envoy ratelimit deployment resource.
If unspecified, default settings for the manged Envoy ratelimit deployment resource
are applied.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewayProvider">EnvoyGatewayProvider
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewaySpec">EnvoyGatewaySpec</a>)
</p>
<p>EnvoyGatewayProvider defines the desired configuration of a provider.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ProviderType">
ProviderType
</a>
</em>
</td>
<td>
<p>Type is the type of provider to use. Supported types are &ldquo;Kubernetes&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>kubernetes</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewayKubernetesProvider">
EnvoyGatewayKubernetesProvider
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Kubernetes defines the configuration of the Kubernetes provider. Kubernetes
provides runtime configuration via the Kubernetes API.</p>
</td>
</tr>
<tr>
<td>
<code>file</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewayFileProvider">
EnvoyGatewayFileProvider
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>File defines the configuration of the File provider. File provides runtime
configuration defined by one or more files. This type is not implemented
until <a href="https://github.com/envoyproxy/gateway/issues/1001">https://github.com/envoyproxy/gateway/issues/1001</a> is fixed.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewaySpec">EnvoyGatewaySpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGateway">EnvoyGateway</a>)
</p>
<p>EnvoyGatewaySpec defines the desired state of Envoy Gateway.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>gateway</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.Gateway">
Gateway
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Gateway defines desired Gateway API specific configuration. If unset,
default configuration parameters will apply.</p>
</td>
</tr>
<tr>
<td>
<code>provider</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewayProvider">
EnvoyGatewayProvider
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Provider defines the desired provider and provider-specific configuration.
If unspecified, the Kubernetes provider is used with default configuration
parameters.</p>
</td>
</tr>
<tr>
<td>
<code>rateLimit</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.RateLimit">
RateLimit
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RateLimit defines the configuration associated with the Rate Limit service
deployed by Envoy Gateway required to implement the Global Rate limiting
functionality. The specific rate limit service used here is the reference
implementation in Envoy. For more details visit <a href="https://github.com/envoyproxy/ratelimit">https://github.com/envoyproxy/ratelimit</a>.
This configuration is unneeded for &ldquo;Local&rdquo; rate limiting.</p>
</td>
</tr>
<tr>
<td>
<code>extension</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.Extension">
Extension
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Extension defines an extension to register for the Envoy Gateway Control Plane.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.EnvoyProxy">EnvoyProxy
</h3>
<p>EnvoyProxy is the schema for the envoyproxies API.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxySpec">
EnvoyProxySpec
</a>
</em>
</td>
<td>
<p>EnvoyProxySpec defines the desired state of EnvoyProxy.</p>
<table>
<tr>
<td>
<code>provider</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxyProvider">
EnvoyProxyProvider
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Provider defines the desired resource provider and provider-specific configuration.
If unspecified, the &ldquo;Kubernetes&rdquo; resource provider is used with default configuration
parameters.</p>
</td>
</tr>
<tr>
<td>
<code>logging</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ProxyLogging">
ProxyLogging
</a>
</em>
</td>
<td>
<p>Logging defines logging parameters for managed proxies. If unspecified,
default settings apply. This type is not implemented until
<a href="https://github.com/envoyproxy/gateway/issues/280">https://github.com/envoyproxy/gateway/issues/280</a> is fixed.</p>
</td>
</tr>
<tr>
<td>
<code>bootstrap</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Bootstrap defines the Envoy Bootstrap as a YAML string.
Visit <a href="https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/bootstrap/v3/bootstrap.proto#envoy-v3-api-msg-config-bootstrap-v3-bootstrap">https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/bootstrap/v3/bootstrap.proto#envoy-v3-api-msg-config-bootstrap-v3-bootstrap</a>
to learn more about the syntax.
If set, this is the Bootstrap configuration used for the managed Envoy Proxy fleet instead of the default Bootstrap configuration
set by Envoy Gateway.
Some fields within the Bootstrap that are required to communicate with the xDS Server (Envoy Gateway) and receive xDS resources
from it are not configurable and will result in the <code>EnvoyProxy</code> resource being rejected.
Backward compatibility across minor versions is not guaranteed.
We strongly recommend using <code>egctl x translate</code> to generate a <code>EnvoyProxy</code> resource with the <code>Bootstrap</code> field set to the default
Bootstrap configuration used. You can edit this configuration, and rerun <code>egctl x translate</code> to ensure there are no validation errors.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxyStatus">
EnvoyProxyStatus
</a>
</em>
</td>
<td>
<p>EnvoyProxyStatus defines the actual state of EnvoyProxy.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.EnvoyProxyKubernetesProvider">EnvoyProxyKubernetesProvider
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxyProvider">EnvoyProxyProvider</a>)
</p>
<p>EnvoyProxyKubernetesProvider defines configuration for the Kubernetes resource
provider.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>envoyDeployment</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.KubernetesDeploymentSpec">
KubernetesDeploymentSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>EnvoyDeployment defines the desired state of the Envoy deployment resource.
If unspecified, default settings for the manged Envoy deployment resource
are applied.</p>
</td>
</tr>
<tr>
<td>
<code>envoyService</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.KubernetesServiceSpec">
KubernetesServiceSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>EnvoyService defines the desired state of the Envoy service resource.
If unspecified, default settings for the manged Envoy service resource
are applied.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.EnvoyProxyProvider">EnvoyProxyProvider
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxySpec">EnvoyProxySpec</a>)
</p>
<p>EnvoyProxyProvider defines the desired state of a resource provider.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ProviderType">
ProviderType
</a>
</em>
</td>
<td>
<p>Type is the type of resource provider to use. A resource provider provides
infrastructure resources for running the data plane, e.g. Envoy proxy, and
optional auxiliary control planes. Supported types are &ldquo;Kubernetes&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>kubernetes</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxyKubernetesProvider">
EnvoyProxyKubernetesProvider
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Kubernetes defines the desired state of the Kubernetes resource provider.
Kubernetes provides infrastructure resources for running the data plane,
e.g. Envoy proxy. If unspecified and type is &ldquo;Kubernetes&rdquo;, default settings
for managed Kubernetes resources are applied.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.EnvoyProxySpec">EnvoyProxySpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxy">EnvoyProxy</a>)
</p>
<p>EnvoyProxySpec defines the desired state of EnvoyProxy.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>provider</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxyProvider">
EnvoyProxyProvider
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Provider defines the desired resource provider and provider-specific configuration.
If unspecified, the &ldquo;Kubernetes&rdquo; resource provider is used with default configuration
parameters.</p>
</td>
</tr>
<tr>
<td>
<code>logging</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ProxyLogging">
ProxyLogging
</a>
</em>
</td>
<td>
<p>Logging defines logging parameters for managed proxies. If unspecified,
default settings apply. This type is not implemented until
<a href="https://github.com/envoyproxy/gateway/issues/280">https://github.com/envoyproxy/gateway/issues/280</a> is fixed.</p>
</td>
</tr>
<tr>
<td>
<code>bootstrap</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Bootstrap defines the Envoy Bootstrap as a YAML string.
Visit <a href="https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/bootstrap/v3/bootstrap.proto#envoy-v3-api-msg-config-bootstrap-v3-bootstrap">https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/bootstrap/v3/bootstrap.proto#envoy-v3-api-msg-config-bootstrap-v3-bootstrap</a>
to learn more about the syntax.
If set, this is the Bootstrap configuration used for the managed Envoy Proxy fleet instead of the default Bootstrap configuration
set by Envoy Gateway.
Some fields within the Bootstrap that are required to communicate with the xDS Server (Envoy Gateway) and receive xDS resources
from it are not configurable and will result in the <code>EnvoyProxy</code> resource being rejected.
Backward compatibility across minor versions is not guaranteed.
We strongly recommend using <code>egctl x translate</code> to generate a <code>EnvoyProxy</code> resource with the <code>Bootstrap</code> field set to the default
Bootstrap configuration used. You can edit this configuration, and rerun <code>egctl x translate</code> to ensure there are no validation errors.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.EnvoyProxyStatus">EnvoyProxyStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxy">EnvoyProxy</a>)
</p>
<p>EnvoyProxyStatus defines the observed state of EnvoyProxy. This type is not implemented
until <a href="https://github.com/envoyproxy/gateway/issues/1007">https://github.com/envoyproxy/gateway/issues/1007</a> is fixed.</p>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.Extension">Extension
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewaySpec">EnvoyGatewaySpec</a>)
</p>
<p>Extension defines the configuration for registering an extension to
the Envoy Gateway control plane.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>resources</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.GroupVersionKind">
[]GroupVersionKind
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Resources defines the set of K8s resources the extension will handle.</p>
</td>
</tr>
<tr>
<td>
<code>hooks</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ExtensionHooks">
ExtensionHooks
</a>
</em>
</td>
<td>
<p>Hooks defines the set of hooks the extension supports</p>
</td>
</tr>
<tr>
<td>
<code>service</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ExtensionService">
ExtensionService
</a>
</em>
</td>
<td>
<p>Service defines the configuration of the extension service that the Envoy
Gateway Control Plane will call through extension hooks.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.ExtensionHooks">ExtensionHooks
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.Extension">Extension</a>)
</p>
<p>ExtensionHooks defines extension hooks across all supported runners</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>xdsTranslator</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.XDSTranslatorHooks">
XDSTranslatorHooks
</a>
</em>
</td>
<td>
<p>XDSTranslator defines all the supported extension hooks for the xds-translator runner</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.ExtensionService">ExtensionService
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.Extension">Extension</a>)
</p>
<p>ExtensionService defines the configuration for connecting to a registered extension service.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>host</code><br>
<em>
string
</em>
</td>
<td>
<p>Host define the extension service hostname.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Port defines the port the extension service is exposed on.</p>
</td>
</tr>
<tr>
<td>
<code>tls</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ExtensionTLS">
ExtensionTLS
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS defines TLS configuration for communication between Envoy Gateway and
the extension service.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.ExtensionTLS">ExtensionTLS
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ExtensionService">ExtensionService</a>)
</p>
<p>ExtensionTLS defines the TLS configuration when connecting to an extension service</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>certificateRef</code><br>
<em>
<a href="https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.SecretObjectReference">
Gateway API v1beta1.SecretObjectReference
</a>
</em>
</td>
<td>
<p>CertificateRef contains a references to objects (Kubernetes objects or otherwise) that
contains a TLS certificate and private keys. These certificates are used to
establish a TLS handshake to the extension server.</p>
<p>CertificateRef can only reference a Kubernetes Secret at this time.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.Gateway">Gateway
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewaySpec">EnvoyGatewaySpec</a>)
</p>
<p>Gateway defines the desired Gateway API configuration of Envoy Gateway.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>controllerName</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ControllerName defines the name of the Gateway API controller. If unspecified,
defaults to &ldquo;gateway.envoyproxy.io/gatewayclass-controller&rdquo;. See the following
for additional details:
<a href="https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GatewayClass">https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GatewayClass</a></p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.GroupVersionKind">GroupVersionKind
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.Extension">Extension</a>)
</p>
<p>GroupVersionKind unambiguously identifies a Kind.
It can be converted to k8s.io/apimachinery/pkg/runtime/schema.GroupVersionKind</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>group</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>version</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>kind</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.KubernetesContainerSpec">KubernetesContainerSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.KubernetesDeploymentSpec">KubernetesDeploymentSpec</a>)
</p>
<p>KubernetesContainerSpec defines the desired state of the Kubernetes container resource.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>resources</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Resources required by this container.
More info: <a href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/">https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/</a></p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#securitycontext-v1-core">
Kubernetes core/v1.SecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecurityContext defines the security options the container should be run with.
If set, the fields of SecurityContext override the equivalent fields of PodSecurityContext.
More info: <a href="https://kubernetes.io/docs/tasks/configure-pod-container/security-context/">https://kubernetes.io/docs/tasks/configure-pod-container/security-context/</a></p>
</td>
</tr>
<tr>
<td>
<code>image</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image specifies the EnvoyProxy container image to be used, instead of the default image.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.KubernetesDeploymentSpec">KubernetesDeploymentSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewayKubernetesProvider">EnvoyGatewayKubernetesProvider</a>, 
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxyKubernetesProvider">EnvoyProxyKubernetesProvider</a>)
</p>
<p>KubernetesDeploymentSpec defines the desired state of the Kubernetes deployment resource.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>replicas</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Replicas is the number of desired pods. Defaults to 1.</p>
</td>
</tr>
<tr>
<td>
<code>pod</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.KubernetesPodSpec">
KubernetesPodSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Pod defines the desired annotations and securityContext of container.</p>
</td>
</tr>
<tr>
<td>
<code>container</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.KubernetesContainerSpec">
KubernetesContainerSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Container defines the resources and securityContext of container.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.KubernetesPodSpec">KubernetesPodSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.KubernetesDeploymentSpec">KubernetesDeploymentSpec</a>)
</p>
<p>KubernetesPodSpec defines the desired state of the Kubernetes pod resource.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>annotations</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Annotations are the annotations that should be appended to the pods.
By default, no pod annotations are appended.</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecurityContext holds pod-level security attributes and common container settings.
Optional: Defaults to empty.  See type description for default values of each field.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.KubernetesServiceSpec">KubernetesServiceSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxyKubernetesProvider">EnvoyProxyKubernetesProvider</a>)
</p>
<p>KubernetesServiceSpec defines the desired state of the Kubernetes service resource.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>annotations</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Annotations that should be appended to the service.
By default, no annotations are appended.</p>
</td>
</tr>
<tr>
<td>
<code>type</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ServiceType">
ServiceType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Type determines how the Service is exposed. Defaults to LoadBalancer.
Valid options are ClusterIP and LoadBalancer.
&ldquo;LoadBalancer&rdquo; means a service will be exposed via an external load balancer (if the cloud provider supports it).
&ldquo;ClusterIP&rdquo; means a service will only be accessible inside the cluster, via the cluster IP.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.LogComponent">LogComponent
(<code>string</code> alias)</h3>
<p>LogComponent defines a component that supports a configured logging level.
This type is not implemented until <a href="https://github.com/envoyproxy/gateway/issues/280">https://github.com/envoyproxy/gateway/issues/280</a>
is fixed.</p>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.LogLevel">LogLevel
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ProxyLogging">ProxyLogging</a>)
</p>
<p>LogLevel defines a log level for system logs. This type is not implemented until
<a href="https://github.com/envoyproxy/gateway/issues/280">https://github.com/envoyproxy/gateway/issues/280</a> is fixed.</p>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.ProviderType">ProviderType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewayProvider">EnvoyGatewayProvider</a>, 
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxyProvider">EnvoyProxyProvider</a>)
</p>
<p>ProviderType defines the types of providers supported by Envoy Gateway.</p>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.ProxyLogging">ProxyLogging
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxySpec">EnvoyProxySpec</a>)
</p>
<p>ProxyLogging defines logging parameters for managed proxies. This type is not
implemented until <a href="https://github.com/envoyproxy/gateway/issues/280">https://github.com/envoyproxy/gateway/issues/280</a> is fixed.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>level</code><br>
<em>
map[<a href="#config.gateway.envoyproxy.io/v1alpha1.LogComponent">
LogComponent
</a>][<a href="#config.gateway.envoyproxy.io/v1alpha1.LogLevel">
LogLevel
</a>]
</em>
</td>
<td>
<p>Level is a map of logging level per component, where the component is the key
and the log level is the value. If unspecified, defaults to &ldquo;System: Info&rdquo;.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.RateLimit">RateLimit
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewaySpec">EnvoyGatewaySpec</a>)
</p>
<p>RateLimit defines the configuration associated with the Rate Limit Service
used for Global Rate Limiting.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>backend</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.RateLimitDatabaseBackend">
RateLimitDatabaseBackend
</a>
</em>
</td>
<td>
<p>Backend holds the configuration associated with the
database backend used by the rate limit service to store
state associated with global ratelimiting.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.RateLimitDatabaseBackend">RateLimitDatabaseBackend
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.RateLimit">RateLimit</a>)
</p>
<p>RateLimitDatabaseBackend defines the configuration associated with
the database backend used by the rate limit service.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.RateLimitDatabaseBackendType">
RateLimitDatabaseBackendType
</a>
</em>
</td>
<td>
<p>Type is the type of database backend to use. Supported types are:
* Redis: Connects to a Redis database.</p>
</td>
</tr>
<tr>
<td>
<code>redis</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.RateLimitRedisSettings">
RateLimitRedisSettings
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Redis defines the settings needed to connect to a Redis database.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.RateLimitDatabaseBackendType">RateLimitDatabaseBackendType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.RateLimitDatabaseBackend">RateLimitDatabaseBackend</a>)
</p>
<p>RateLimitDatabaseBackendType specifies the types of database backend
to be used by the rate limit service.</p>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.RateLimitRedisSettings">RateLimitRedisSettings
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.RateLimitDatabaseBackend">RateLimitDatabaseBackend</a>)
</p>
<p>RateLimitRedisSettings defines the configuration for connecting to
a Redis database.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>url</code><br>
<em>
string
</em>
</td>
<td>
<p>URL of the Redis Database.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.ServiceType">ServiceType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.KubernetesServiceSpec">KubernetesServiceSpec</a>)
</p>
<p>ServiceType string describes ingress methods for a service</p>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.XDSTranslatorHook">XDSTranslatorHook
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.XDSTranslatorHooks">XDSTranslatorHooks</a>)
</p>
<p>XDSTranslatorHook defines the types of hooks that an Envoy Gateway extension may support
for the xds-translator</p>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.XDSTranslatorHooks">XDSTranslatorHooks
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ExtensionHooks">ExtensionHooks</a>)
</p>
<p>XDSTranslatorHooks contains all the pre and post hooks for the xds-translator runner.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
<table class="docutils align-default">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>pre</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.XDSTranslatorHook">
[]XDSTranslatorHook
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>post</code><br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.XDSTranslatorHook">
[]XDSTranslatorHook
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<div class="admonition note">
<p class="last">This page was automatically generated with <code>gen-crd-api-reference-docs</code></p>
</div>
