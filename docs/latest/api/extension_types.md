# gateway.envoyproxy.io/v1alpha1
<p>Packages:</p>
<ul class="simple">
<li>
<a href="#gateway.envoyproxy.io%2fv1alpha1">gateway.envoyproxy.io/v1alpha1</a>
</li>
</ul>
<h2 id="gateway.envoyproxy.io/v1alpha1">gateway.envoyproxy.io/v1alpha1</h2>
<p>Package v1alpha1 contains API schema definitions for the gateway.envoyproxy.io API group.</p>
Resource Types:
<ul class="simple"></ul>
<h3 id="gateway.envoyproxy.io/v1alpha1.AuthenticationFilter">AuthenticationFilter
</h3>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
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
<a href="#gateway.envoyproxy.io/v1alpha1.AuthenticationFilterSpec">
AuthenticationFilterSpec
</a>
</em>
</td>
<td>
<p>Spec defines the desired state of the AuthenticationFilter type.</p>
<table>
<tr>
<td>
<code>type</code><br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.AuthenticationFilterType">
AuthenticationFilterType
</a>
</em>
</td>
<td>
<p>Type defines the type of authentication provider to use. Supported provider types
are &ldquo;JWT&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>jwtProviders</code><br>
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
successfully validate the JWT. For additional details, see
<a href="https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html">https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html</a>.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="gateway.envoyproxy.io/v1alpha1.AuthenticationFilterSpec">AuthenticationFilterSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.AuthenticationFilter">AuthenticationFilter</a>)
</p>
<p>AuthenticationFilterSpec defines the desired state of the AuthenticationFilter type.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
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
<code>type</code><br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.AuthenticationFilterType">
AuthenticationFilterType
</a>
</em>
</td>
<td>
<p>Type defines the type of authentication provider to use. Supported provider types
are &ldquo;JWT&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>jwtProviders</code><br>
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
successfully validate the JWT. For additional details, see
<a href="https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html">https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html</a>.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="gateway.envoyproxy.io/v1alpha1.AuthenticationFilterType">AuthenticationFilterType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.AuthenticationFilterSpec">AuthenticationFilterSpec</a>)
</p>
<p>AuthenticationFilterType is a type of authentication provider.</p>
<h3 id="gateway.envoyproxy.io/v1alpha1.GlobalRateLimit">GlobalRateLimit
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitFilterSpec">RateLimitFilterSpec</a>)
</p>
<p>GlobalRateLimit defines global rate limit configuration.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
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
<code>rules</code><br>
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
</div>
</div>
<h3 id="gateway.envoyproxy.io/v1alpha1.HeaderMatch">HeaderMatch
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitSelectCondition">RateLimitSelectCondition</a>)
</p>
<p>HeaderMatch defines the match attributes within the HTTP Headers of the request.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
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
<code>type</code><br>
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
<code>name</code><br>
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
<code>value</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Value within the HTTP header. Due to the
case-insensitivity of header names, &ldquo;foo&rdquo; and &ldquo;Foo&rdquo; are considered equivalent.
Do not set this field when Type=&ldquo;Distinct&rdquo;, implying matching on any/all unique
values within the header.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="gateway.envoyproxy.io/v1alpha1.HeaderMatchType">HeaderMatchType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.HeaderMatch">HeaderMatch</a>)
</p>
<p>HeaderMatchType specifies the semantics of how HTTP header values should be compared.
Valid HeaderMatchType values are &ldquo;Exact&rdquo;, &ldquo;RegularExpression&rdquo;, and &ldquo;Distinct&rdquo;.</p>
<h3 id="gateway.envoyproxy.io/v1alpha1.JwtAuthenticationFilterProvider">JwtAuthenticationFilterProvider
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.AuthenticationFilterSpec">AuthenticationFilterSpec</a>)
</p>
<p>JwtAuthenticationFilterProvider defines the JSON Web Token (JWT) authentication provider type
and how JWTs should be verified:</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
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
<code>name</code><br>
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
<code>issuer</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Issuer is the principal that issued the JWT and takes the form of a URL or email address.
For additional details, see <a href="https://tools.ietf.org/html/rfc7519#section-4.1.1">https://tools.ietf.org/html/rfc7519#section-4.1.1</a> for
URL format and <a href="https://rfc-editor.org/rfc/rfc5322.html">https://rfc-editor.org/rfc/rfc5322.html</a> for email format. If not provided,
the JWT issuer is not checked.</p>
</td>
</tr>
<tr>
<td>
<code>audiences</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Audiences is a list of JWT audiences allowed access. For additional details, see
<a href="https://tools.ietf.org/html/rfc7519#section-4.1.3">https://tools.ietf.org/html/rfc7519#section-4.1.3</a>. If not provided, JWT audiences
are not checked.</p>
</td>
</tr>
<tr>
<td>
<code>remoteJWKS</code><br>
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
</div>
</div>
<h3 id="gateway.envoyproxy.io/v1alpha1.RateLimitFilter">RateLimitFilter
</h3>
<p>RateLimitFilter allows the user to limit the number of incoming requests
to a predefined value based on attributes within the traffic flow.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
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
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitFilterSpec">
RateLimitFilterSpec
</a>
</em>
</td>
<td>
<p>Spec defines the desired state of RateLimitFilter.</p>
<table>
<tr>
<td>
<code>type</code><br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitType">
RateLimitType
</a>
</em>
</td>
<td>
<p>Type decides the scope for the RateLimits.
Valid RateLimitType values are &ldquo;Global&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>global</code><br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.GlobalRateLimit">
GlobalRateLimit
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Global defines global rate limit configuration.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="gateway.envoyproxy.io/v1alpha1.RateLimitFilterSpec">RateLimitFilterSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitFilter">RateLimitFilter</a>)
</p>
<p>RateLimitFilterSpec defines the desired state of RateLimitFilter.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
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
<code>type</code><br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitType">
RateLimitType
</a>
</em>
</td>
<td>
<p>Type decides the scope for the RateLimits.
Valid RateLimitType values are &ldquo;Global&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>global</code><br>
<em>
<a href="#gateway.envoyproxy.io/v1alpha1.GlobalRateLimit">
GlobalRateLimit
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Global defines global rate limit configuration.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="gateway.envoyproxy.io/v1alpha1.RateLimitRule">RateLimitRule
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.GlobalRateLimit">GlobalRateLimit</a>)
</p>
<p>RateLimitRule defines the semantics for matching attributes
from the incoming requests, and setting limits for them.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
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
<code>clientSelectors</code><br>
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
<code>limit</code><br>
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
</div>
</div>
<h3 id="gateway.envoyproxy.io/v1alpha1.RateLimitSelectCondition">RateLimitSelectCondition
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitRule">RateLimitRule</a>)
</p>
<p>RateLimitSelectCondition specifies the attributes within the traffic flow that can
be used to select a subset of clients to be ratelimited.
All the individual conditions must hold True for the overall condition to hold True.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
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
<code>headers</code><br>
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
<tr>
<td>
<code>sourceIP</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SourceIP is the IP CIDR that represents the range of Source IP Addresses of the client.
These could also be the intermediate addresses through which the request has flown through and is part of the  <code>X-Forwarded-For</code> header.
For example, <code>192.168.0.1/32</code>, <code>192.168.0.0/24</code>, <code>001:db8::/64</code>.
All IP Addresses within the specified SourceIP CIDR are treated as a single client selector and share the same rate limit bucket.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="gateway.envoyproxy.io/v1alpha1.RateLimitType">RateLimitType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitFilterSpec">RateLimitFilterSpec</a>)
</p>
<p>RateLimitType specifies the types of RateLimiting.</p>
<h3 id="gateway.envoyproxy.io/v1alpha1.RateLimitUnit">RateLimitUnit
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitValue">RateLimitValue</a>)
</p>
<p>RateLimitUnit specifies the intervals for setting rate limits.
Valid RateLimitUnit values are &ldquo;Second&rdquo;, &ldquo;Minute&rdquo;, &ldquo;Hour&rdquo;, and &ldquo;Day&rdquo;.</p>
<h3 id="gateway.envoyproxy.io/v1alpha1.RateLimitValue">RateLimitValue
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.RateLimitRule">RateLimitRule</a>)
</p>
<p>RateLimitValue defines the limits for rate limiting.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
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
<code>requests</code><br>
<em>
uint
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>unit</code><br>
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
</div>
</div>
<h3 id="gateway.envoyproxy.io/v1alpha1.RemoteJWKS">RemoteJWKS
</h3>
<p>
(<em>Appears on:</em>
<a href="#gateway.envoyproxy.io/v1alpha1.JwtAuthenticationFilterProvider">JwtAuthenticationFilterProvider</a>)
</p>
<p>RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote
HTTP/HTTPS endpoint.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table td-content">
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
<code>uri</code><br>
<em>
string
</em>
</td>
<td>
<p>URI is the HTTPS URI to fetch the JWKS. Envoy&rsquo;s system trust bundle is used to
validate the server certificate.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<div class="admonition note">
<p class="last">This page was automatically generated with <code>gen-crd-api-reference-docs</code></p>
</div>
