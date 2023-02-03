# config.gateway.envoyproxy.io/v1alpha1
<p>Packages:</p>
<ul>
<li>
<a href="#config.gateway.envoyproxy.io%2fv1alpha1">config.gateway.envoyproxy.io/v1alpha1</a>
</li>
</ul>
<h2 id="config.gateway.envoyproxy.io/v1alpha1">config.gateway.envoyproxy.io/v1alpha1</h2>
<div>
<p>Package v1alpha1 contains API Schema definitions for the config v1alpha1 API group.</p>
</div>
Resource Types:
<ul></ul>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.EnvoyGateway">EnvoyGateway
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.EnvoyGateway" title="Permanent link">¶</a>
</h3>
<p>
<p>EnvoyGateway is the Schema for the envoygateways API.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>EnvoyGatewaySpec</code></br>
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
<p>EnvoyGatewaySpec defines the desired state of Envoy Gateway.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewaySpec">EnvoyGatewaySpec
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.EnvoyGatewaySpec" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGateway">EnvoyGateway</a>)
</p>
<p>
<p>EnvoyGatewaySpec defines the desired state of Envoy Gateway.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>gateway</code></br>
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
<code>provider</code></br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.Provider">
Provider
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
<code>rateLimit</code></br>
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
This configuration will not be needed to enable Local Rate limiitng.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.EnvoyProxy">EnvoyProxy
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.EnvoyProxy" title="Permanent link">¶</a>
</h3>
<p>
<p>EnvoyProxy is the Schema for the envoyproxies API</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#objectmeta-v1-meta">
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
<code>spec</code></br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxySpec">
EnvoyProxySpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>provider</code></br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ResourceProvider">
ResourceProvider
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
<code>logging</code></br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ProxyLogging">
ProxyLogging
</a>
</em>
</td>
<td>
<p>Logging defines logging parameters for managed proxies. If unspecified,
default settings apply.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxyStatus">
EnvoyProxyStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.EnvoyProxySpec">EnvoyProxySpec
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.EnvoyProxySpec" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxy">EnvoyProxy</a>)
</p>
<p>
<p>EnvoyProxySpec defines the desired state of EnvoyProxy.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>provider</code></br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ResourceProvider">
ResourceProvider
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
<code>logging</code></br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ProxyLogging">
ProxyLogging
</a>
</em>
</td>
<td>
<p>Logging defines logging parameters for managed proxies. If unspecified,
default settings apply.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.EnvoyProxyStatus">EnvoyProxyStatus
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.EnvoyProxyStatus" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxy">EnvoyProxy</a>)
</p>
<p>
<p>EnvoyProxyStatus defines the observed state of EnvoyProxy</p>
</p>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.FileProvider">FileProvider
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.FileProvider" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.Provider">Provider</a>)
</p>
<p>
<p>FileProvider defines configuration for the File provider.</p>
</p>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.Gateway">Gateway
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.Gateway" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewaySpec">EnvoyGatewaySpec</a>)
</p>
<p>
<p>Gateway defines the desired Gateway API configuration of Envoy Gateway.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>controllerName</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ControllerName defines the name of the Gateway API controller. If unspecified,
defaults to &ldquo;gateway.envoyproxy.io/gatewayclass-controller&rdquo;. See the following
for additional details:</p>
<p><a href="https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GatewayClass">https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GatewayClass</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.KubernetesDeploymentSpec">KubernetesDeploymentSpec
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.KubernetesDeploymentSpec" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.KubernetesResourceProvider">KubernetesResourceProvider</a>)
</p>
<p>
<p>KubernetesDeploymentSpec defines the desired state of the Kubernetes deployment resource.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>replicas</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Replicas is the number of desired pods. Defaults to 1.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.KubernetesProvider">KubernetesProvider
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.KubernetesProvider" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.Provider">Provider</a>)
</p>
<p>
<p>KubernetesProvider defines configuration for the Kubernetes provider.</p>
</p>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.KubernetesResourceProvider">KubernetesResourceProvider
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.KubernetesResourceProvider" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ResourceProvider">ResourceProvider</a>)
</p>
<p>
<p>KubernetesResourceProvider defines configuration for the Kubernetes resource
provider.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>envoyDeployment</code></br>
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
</tbody>
</table>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.LogComponent">LogComponent
(<code>string</code> alias)</p><a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.LogComponent" title="Permanent link">¶</a>
</h3>
<p>
<p>LogComponent defines a component that supports a configured logging level.</p>
</p>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;admin&#34;</p></td>
<td><p>LogComponentAdmin defines defines the &ldquo;admin&rdquo; logging component.</p>
</td>
</tr><tr><td><p>&#34;client&#34;</p></td>
<td><p>LogComponentClient defines defines the &ldquo;client&rdquo; logging component.</p>
</td>
</tr><tr><td><p>&#34;connection&#34;</p></td>
<td><p>LogComponentConnection defines defines the &ldquo;connection&rdquo; logging component.</p>
</td>
</tr><tr><td><p>&#34;filter&#34;</p></td>
<td><p>LogComponentFilter defines defines the &ldquo;filter&rdquo; logging component.</p>
</td>
</tr><tr><td><p>&#34;http&#34;</p></td>
<td><p>LogComponentHTTP defines defines the &ldquo;http&rdquo; logging component.</p>
</td>
</tr><tr><td><p>&#34;main&#34;</p></td>
<td><p>LogComponentMain defines defines the &ldquo;main&rdquo; logging component.</p>
</td>
</tr><tr><td><p>&#34;router&#34;</p></td>
<td><p>LogComponentRouter defines defines the &ldquo;router&rdquo; logging component.</p>
</td>
</tr><tr><td><p>&#34;runtime&#34;</p></td>
<td><p>LogComponentRuntime defines defines the &ldquo;runtime&rdquo; logging component.</p>
</td>
</tr><tr><td><p>&#34;system&#34;</p></td>
<td><p>LogComponentSystem defines the &ldquo;system&rdquo;-wide logging component. When specified,
all other logging components are ignored.</p>
</td>
</tr><tr><td><p>&#34;upstream&#34;</p></td>
<td><p>LogComponentUpstream defines defines the &ldquo;upstream&rdquo; logging component.</p>
</td>
</tr></tbody>
</table>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.LogLevel">LogLevel
(<code>string</code> alias)</p><a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.LogLevel" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ProxyLogging">ProxyLogging</a>)
</p>
<p>
<p>LogLevel defines a log level for system logs.</p>
</p>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;debug&#34;</p></td>
<td><p>LogLevelDebug defines the &ldquo;debug&rdquo; logging level.</p>
</td>
</tr><tr><td><p>&#34;error&#34;</p></td>
<td><p>LogLevelError defines the &ldquo;Error&rdquo; logging level.</p>
</td>
</tr><tr><td><p>&#34;info&#34;</p></td>
<td><p>LogLevelInfo defines the &ldquo;Info&rdquo; logging level.</p>
</td>
</tr></tbody>
</table>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.Provider">Provider
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.Provider" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewaySpec">EnvoyGatewaySpec</a>)
</p>
<p>
<p>Provider defines the desired configuration of a provider.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code></br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ProviderType">
ProviderType
</a>
</em>
</td>
<td>
<p>Type is the type of provider to use. Supported types are:</p>
<ul>
<li>Kubernetes: A provider that provides runtime configuration via the Kubernetes API.</li>
</ul>
</td>
</tr>
<tr>
<td>
<code>kubernetes</code></br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.KubernetesProvider">
KubernetesProvider
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
<code>file</code></br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.FileProvider">
FileProvider
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>File defines the configuration of the File provider. File provides runtime
configuration defined by one or more files.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.ProviderType">ProviderType
(<code>string</code> alias)</p><a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.ProviderType" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.Provider">Provider</a>, 
<a href="#config.gateway.envoyproxy.io/v1alpha1.ResourceProvider">ResourceProvider</a>)
</p>
<p>
<p>ProviderType defines the types of providers supported by Envoy Gateway.</p>
</p>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;File&#34;</p></td>
<td><p>ProviderTypeFile defines the &ldquo;File&rdquo; provider.</p>
</td>
</tr><tr><td><p>&#34;Kubernetes&#34;</p></td>
<td><p>ProviderTypeKubernetes defines the &ldquo;Kubernetes&rdquo; provider.</p>
</td>
</tr></tbody>
</table>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.ProxyLogging">ProxyLogging
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.ProxyLogging" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxySpec">EnvoyProxySpec</a>)
</p>
<p>
<p>ProxyLogging defines logging parameters for managed proxies.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>level</code></br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.LogLevel">
map[github.com/envoyproxy/gateway/api/config/v1alpha1.LogComponent]github.com/envoyproxy/gateway/api/config/v1alpha1.LogLevel
</a>
</em>
</td>
<td>
<p>Level is a map of logging level per component, where the component is the key
and the log level is the value. If unspecified, defaults to &ldquo;System: Info&rdquo;.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.RateLimit">RateLimit
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.RateLimit" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyGatewaySpec">EnvoyGatewaySpec</a>)
</p>
<p>
<p>RateLimit defines the configuration associated with the Rate Limit Service
used for Global Rate Limiting.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>backend</code></br>
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
<h3 id="config.gateway.envoyproxy.io/v1alpha1.RateLimitDatabaseBackend">RateLimitDatabaseBackend
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.RateLimitDatabaseBackend" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.RateLimit">RateLimit</a>)
</p>
<p>
<p>RateLimitDatabaseBackend defines the configuration associated with
the database backend used by the rate limit service.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code></br>
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
<code>redis</code></br>
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
<h3 id="config.gateway.envoyproxy.io/v1alpha1.RateLimitDatabaseBackendType">RateLimitDatabaseBackendType
(<code>string</code> alias)</p><a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.RateLimitDatabaseBackendType" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.RateLimitDatabaseBackend">RateLimitDatabaseBackend</a>)
</p>
<p>
<p>RateLimitDatabaseBackendType specifies the types of database backend
to be used by the rate limit service.</p>
</p>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Redis&#34;</p></td>
<td><p>RedisBackendType uses a redis database for the rate limit service.</p>
</td>
</tr></tbody>
</table>
<h3 id="config.gateway.envoyproxy.io/v1alpha1.RateLimitRedisSettings">RateLimitRedisSettings
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.RateLimitRedisSettings" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.RateLimitDatabaseBackend">RateLimitDatabaseBackend</a>)
</p>
<p>
<p>RateLimitRedisSettings defines the configuration for connecting to
a Redis database.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>url</code></br>
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
<h3 id="config.gateway.envoyproxy.io/v1alpha1.ResourceProvider">ResourceProvider
<a class="headerlink" href="#config.gateway.envoyproxy.io%2fv1alpha1.ResourceProvider" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.EnvoyProxySpec">EnvoyProxySpec</a>)
</p>
<p>
<p>ResourceProvider defines the desired state of a resource provider.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code></br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.ProviderType">
ProviderType
</a>
</em>
</td>
<td>
<p>Type is the type of resource provider to use. A resource provider provides
infrastructure resources for running the data plane, e.g. Envoy proxy, and
optional auxiliary control planes. Supported types are:</p>
<ul>
<li>Kubernetes: Provides infrastructure resources for running the data plane,
e.g. Envoy proxy.</li>
</ul>
</td>
</tr>
<tr>
<td>
<code>kubernetes</code></br>
<em>
<a href="#config.gateway.envoyproxy.io/v1alpha1.KubernetesResourceProvider">
KubernetesResourceProvider
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
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>2a527f7</code>.
</em></p>
