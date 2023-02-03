# gateway.envoyproxy.io/v1alpha1
<p>Packages:</p>
<ul>
<li>
<a href="#gateway.envoyproxy.io%2fv1alpha1">gateway.envoyproxy.io/v1alpha1</a>
</li>
</ul>
<h2 id="gateway.envoyproxy.io/v1alpha1">gateway.envoyproxy.io/v1alpha1</h2>
<div>
<p>Package v1alpha1 contains API Schema definitions for the gateway.envoyproxy.io API group.</p>
</div>
Resource Types:
<ul></ul>
<h3 id="gateway.envoyproxy.io/v1alpha1.AuthenticationFilter">AuthenticationFilter
<a class="headerlink" href="#gateway.envoyproxy.io%2fv1alpha1.AuthenticationFilter" title="Permanent link">¶</a>
</h3>
<p>
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
<a href="#gateway.envoyproxy.io/v1alpha1.AuthenticationFilterSpec">
AuthenticationFilterSpec
</a>
</em>
</td>
<td>
<p>Spec defines the desired state of the AuthenticationFilter type.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>type</code></br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.AuthenticationFilterType">
AuthenticationFilterType
</a>
</em>
</td>
<td>
<p>Type defines the type of authentication provider to use. Supported provider types are:</p>
<ul>
<li>JWT: A provider that uses JSON Web Token (JWT) for authenticating requests.</li>
</ul>
</td>
</tr>
<tr>
<td>
<code>jwtProviders</code></br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.JwtAuthenticationFilterProvider">
[]JwtAuthenticationFilterProvider
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>JWT defines the JSON Web Token (JWT) authentication provider type. When multiple
jwtProviders are specified, the JWT is considered valid if any of the providers
successfully validate the JWT. For additional details, see:</p>
<p><a href="https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html">https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html</a></p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.envoyproxy.io/v1alpha1.AuthenticationFilterSpec">AuthenticationFilterSpec
<a class="headerlink" href="#gateway.envoyproxy.io%2fv1alpha1.AuthenticationFilterSpec" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.AuthenticationFilter">AuthenticationFilter</a>)
</p>
<p>
<p>AuthenticationFilterSpec defines the desired state of the AuthenticationFilter type.</p>
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
<a href="#gateway.envoyproxy.io/v1alpha1.AuthenticationFilterType">
AuthenticationFilterType
</a>
</em>
</td>
<td>
<p>Type defines the type of authentication provider to use. Supported provider types are:</p>
<ul>
<li>JWT: A provider that uses JSON Web Token (JWT) for authenticating requests.</li>
</ul>
</td>
</tr>
<tr>
<td>
<code>jwtProviders</code></br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.JwtAuthenticationFilterProvider">
[]JwtAuthenticationFilterProvider
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>JWT defines the JSON Web Token (JWT) authentication provider type. When multiple
jwtProviders are specified, the JWT is considered valid if any of the providers
successfully validate the JWT. For additional details, see:</p>
<p><a href="https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html">https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.envoyproxy.io/v1alpha1.AuthenticationFilterType">AuthenticationFilterType
(<code>string</code> alias)</p><a class="headerlink" href="#gateway.envoyproxy.io%2fv1alpha1.AuthenticationFilterType" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.AuthenticationFilterSpec">AuthenticationFilterSpec</a>)
</p>
<p>
<p>AuthenticationFilterType is a type of authentication provider.</p>
</p>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;JWT&#34;</p></td>
<td><p>JwtAuthenticationFilterProviderType is the JWT authentication provider type.</p>
</td>
</tr></tbody>
</table>
<h3 id="gateway.envoyproxy.io/v1alpha1.GlobalRateLimit">GlobalRateLimit
<a class="headerlink" href="#gateway.envoyproxy.io%2fv1alpha1.GlobalRateLimit" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitFilterSpec">RateLimitFilterSpec</a>)
</p>
<p>
<p>GlobalRateLimit defines the global rate limit configuration.</p>
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
<code>rules</code></br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitRule">
[]RateLimitRule
</a>
</em>
</td>
<td>
<p>Rules are a list of RateLimit selectors and limits.
Each rule and its associated limit is applied
in a mutually exclusive way i.e. if multiple
rules get selected, each of their associated
limits get applied, so a single traffic request
might increase the rate limit counters for multiple
rules if selected.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.envoyproxy.io/v1alpha1.HeaderMatch">HeaderMatch
<a class="headerlink" href="#gateway.envoyproxy.io%2fv1alpha1.HeaderMatch" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitSelectCondition">RateLimitSelectCondition</a>)
</p>
<p>
<p>HeaderMatch defines the match attributes within the HTTP Headers of the request.</p>
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
<a href="#gateway.envoyproxy.io/v1alpha1.HeaderMatchType">
HeaderMatchType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Type specifies how to match against the value of the header.</p>
</td>
</tr>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name of the HTTP header.</p>
</td>
</tr>
<tr>
<td>
<code>value</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Value within the HTTP header. Due to the
case-insensitivity of header names, &ldquo;foo&rdquo; and &ldquo;Foo&rdquo; are considered equivalent.
Do not set this field when Type=&ldquo;Distinct&rdquo;, implying matching on any/all unique values within the header.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.envoyproxy.io/v1alpha1.HeaderMatchType">HeaderMatchType
(<code>string</code> alias)</p><a class="headerlink" href="#gateway.envoyproxy.io%2fv1alpha1.HeaderMatchType" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.HeaderMatch">HeaderMatch</a>)
</p>
<p>
<p>HeaderMatchType specifies the semantics of how HTTP header values should be
compared. Valid HeaderMatchType values are:</p>
<ul>
<li>&ldquo;Exact&rdquo;: Use this type to match the exact value of the Value field against the value of the specified HTTP Header.</li>
<li>&ldquo;RegularExpression&rdquo;: Use this type to match a regular expression against the value of the specified HTTP Header.
The regex string must adhere to the syntax documented in <a href="https://github.com/google/re2/wiki/Syntax">https://github.com/google/re2/wiki/Syntax</a>.</li>
<li>&ldquo;Distinct&rdquo;: Use this type to match any and all possible unique values encountered in the specified HTTP Header.
Note that each unique value will receive its own rate limit bucket.</li>
</ul>
</p>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Distinct&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Exact&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;RegularExpression&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="gateway.envoyproxy.io/v1alpha1.JwtAuthenticationFilterProvider">JwtAuthenticationFilterProvider
<a class="headerlink" href="#gateway.envoyproxy.io%2fv1alpha1.JwtAuthenticationFilterProvider" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.AuthenticationFilterSpec">AuthenticationFilterSpec</a>)
</p>
<p>
<p>JwtAuthenticationFilterProvider defines the JSON Web Token (JWT) authentication provider type
and how JWTs should be verified:</p>
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
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name defines a unique name for the JWT provider. A name can have a variety of forms,
including RFC1123 subdomains, RFC 1123 labels, or RFC 1035 labels.</p>
</td>
</tr>
<tr>
<td>
<code>issuer</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Issuer is the principal that issued the JWT and takes the form of a URL or email address.
For additional details, see:</p>
<p>URL format: <a href="https://tools.ietf.org/html/rfc7519#section-4.1.1">https://tools.ietf.org/html/rfc7519#section-4.1.1</a>
Email format: <a href="https://rfc-editor.org/rfc/rfc5322.html">https://rfc-editor.org/rfc/rfc5322.html</a></p>
<p>URL Example:
issuer: <a href="https://auth.example.com">https://auth.example.com</a></p>
<p>Email Example:
issuer: jdoe@example.com</p>
<p>If not provided, the JWT issuer is not checked.</p>
</td>
</tr>
<tr>
<td>
<code>audiences</code></br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Audiences is a list of JWT audiences allowed to access. For additional details, see:</p>
<p><a href="https://tools.ietf.org/html/rfc7519#section-4.1.3">https://tools.ietf.org/html/rfc7519#section-4.1.3</a></p>
<p>Example:
audiences:
- foo.apps.example.com
bar.apps.example.com</p>
<p>If not provided, JWT audiences are not checked.</p>
</td>
</tr>
<tr>
<td>
<code>remoteJWKS</code></br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.RemoteJWKS">
RemoteJWKS
</a>
</em>
</td>
<td>
<p>RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote
HTTP/HTTPS endpoint.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.envoyproxy.io/v1alpha1.RateLimitFilter">RateLimitFilter
<a class="headerlink" href="#gateway.envoyproxy.io%2fv1alpha1.RateLimitFilter" title="Permanent link">¶</a>
</h3>
<p>
<p>RateLimitFilter allows the user to limit the number of incoming requests
to a predefined value based on attributes within the traffic flow.</p>
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
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitFilterSpec">
RateLimitFilterSpec
</a>
</em>
</td>
<td>
<p>Spec defines the desired state of RateLimitFilter.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>type</code></br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitType">
RateLimitType
</a>
</em>
</td>
<td>
<p>Type decides the scope for the RateLimits.
Valid RateLimitType values are:</p>
<ul>
<li>&ldquo;Global&rdquo; - In this mode, the rate limits are applied across all Envoy proxy instances.</li>
</ul>
</td>
</tr>
<tr>
<td>
<code>global</code></br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.GlobalRateLimit">
GlobalRateLimit
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Global rate limit configuration.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.envoyproxy.io/v1alpha1.RateLimitFilterSpec">RateLimitFilterSpec
<a class="headerlink" href="#gateway.envoyproxy.io%2fv1alpha1.RateLimitFilterSpec" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitFilter">RateLimitFilter</a>)
</p>
<p>
<p>RateLimitFilterSpec defines the desired state of RateLimitFilter.</p>
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
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitType">
RateLimitType
</a>
</em>
</td>
<td>
<p>Type decides the scope for the RateLimits.
Valid RateLimitType values are:</p>
<ul>
<li>&ldquo;Global&rdquo; - In this mode, the rate limits are applied across all Envoy proxy instances.</li>
</ul>
</td>
</tr>
<tr>
<td>
<code>global</code></br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.GlobalRateLimit">
GlobalRateLimit
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Global rate limit configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.envoyproxy.io/v1alpha1.RateLimitRule">RateLimitRule
<a class="headerlink" href="#gateway.envoyproxy.io%2fv1alpha1.RateLimitRule" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.GlobalRateLimit">GlobalRateLimit</a>)
</p>
<p>
<p>RateLimitRule defines the semantics for matching attributes
from the incoming requests, and setting limits for them.</p>
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
<code>clientSelectors</code></br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitSelectCondition">
[]RateLimitSelectCondition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ClientSelectors holds the list of select conditions to select
specific clients using attributes from the traffic flow.
All individual select conditions must hold True for this rule
and its limit to be applied.
If this field is empty, it is equivalent to True, and
the limit is applied.</p>
</td>
</tr>
<tr>
<td>
<code>limit</code></br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitValue">
RateLimitValue
</a>
</em>
</td>
<td>
<p>Limit holds the rate limit values.
This limit is applied for traffic flows when the selectors
compute to True, causing the request to be counted towards the limit.
The limit is enforced and the request is ratelimited, i.e. a response with
429 HTTP status code is sent back to the client when
the selected requests have reached the limit.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.envoyproxy.io/v1alpha1.RateLimitSelectCondition">RateLimitSelectCondition
<a class="headerlink" href="#gateway.envoyproxy.io%2fv1alpha1.RateLimitSelectCondition" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitRule">RateLimitRule</a>)
</p>
<p>
<p>RateLimitSelectCondition specifies the attributes within the traffic flow that can
be used to select a subset of clients to be ratelimited.
All the individual conditions must hold True for the overall condition to hold True.</p>
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
<code>headers</code></br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.HeaderMatch">
[]HeaderMatch
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Headers is a list of request headers to match. Multiple header values are ANDed together,
meaning, a request MUST match all the specified headers.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.envoyproxy.io/v1alpha1.RateLimitType">RateLimitType
(<code>string</code> alias)</p><a class="headerlink" href="#gateway.envoyproxy.io%2fv1alpha1.RateLimitType" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitFilterSpec">RateLimitFilterSpec</a>)
</p>
<p>
<p>RateLimitType specifies the types of RateLimiting.</p>
</p>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Global&#34;</p></td>
<td><p>GlobalRateLimitType allows the rate limits to be applied across all Envoy proxy instances.</p>
</td>
</tr></tbody>
</table>
<h3 id="gateway.envoyproxy.io/v1alpha1.RateLimitUnit">RateLimitUnit
(<code>string</code> alias)</p><a class="headerlink" href="#gateway.envoyproxy.io%2fv1alpha1.RateLimitUnit" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitValue">RateLimitValue</a>)
</p>
<p>
<p>RateLimitUnit specifies the intervals for setting rate limits.
Valid RateLimitUnit values are:</p>
<ul>
<li>&ldquo;Second&rdquo;</li>
<li>&ldquo;Minute&rdquo;</li>
<li>&ldquo;Hour&rdquo;</li>
<li>&ldquo;Day&rdquo;</li>
</ul>
</p>
<h3 id="gateway.envoyproxy.io/v1alpha1.RateLimitValue">RateLimitValue
<a class="headerlink" href="#gateway.envoyproxy.io%2fv1alpha1.RateLimitValue" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitRule">RateLimitRule</a>)
</p>
<p>
<p>RateLimitValue defines the limits for rate limiting.</p>
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
<code>requests</code></br>
<em>
uint
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>unit</code></br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitUnit">
RateLimitUnit
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.envoyproxy.io/v1alpha1.RemoteJWKS">RemoteJWKS
<a class="headerlink" href="#gateway.envoyproxy.io%2fv1alpha1.RemoteJWKS" title="Permanent link">¶</a>
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.JwtAuthenticationFilterProvider">JwtAuthenticationFilterProvider</a>)
</p>
<p>
<p>RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote
HTTP/HTTPS endpoint.</p>
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
<code>uri</code></br>
<em>
string
</em>
</td>
<td>
<p>URI is the HTTPS URI to fetch the JWKS. Envoy&rsquo;s system trust bundle is used to
validate the server certificate.</p>
<p>Example:
uri: <a href="https://www.foo.com/oauth2/v1/certs">https://www.foo.com/oauth2/v1/certs</a></p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>2a527f7</code>.
</em></p>
