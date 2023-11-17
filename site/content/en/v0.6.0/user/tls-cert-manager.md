---
title: "Using cert-manager For TLS Termination"
---

This guide shows how to set up [cert-manager](https://cert-manager.io/) to automatically create certificates and secrets for use by Envoy Gateway.
It will first show how to enable the self-sign issuer, which is useful to test that cert-manager and Envoy Gateway can talk to each other.
Then it shows how to use [Let's Encrypt's staging environment](https://letsencrypt.org/docs/staging-environment/).
Changing to the Let's Encrypt production environment is straight-forward after that.

## Prerequisites

* A Kubernetes cluster and a configured `kubectl`.
* The `helm` command.
* The `curl` command or similar for testing HTTPS requests.
* For the ACME HTTP-01 challenge to work
  * your Gateway must be reachable on the public Internet.
  * the domain name you use (we use `www.example.com`) must point to the Gateway's external IP(s).

## Installation

Follow the steps from the [Quickstart Guide](../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Deploying cert-manager

*This is a summary of [cert-manager Installation with Helm](https://cert-manager.io/docs/installation/helm/).*

Installing cert-manager is straight-forward, but currently (v1.12) requires setting a feature gate to enable the Gateway API support.

```console
$ helm repo add jetstack https://charts.jetstack.io
$ helm upgrade --install --create-namespace --namespace cert-manager --set installCRDs=true --set featureGates=ExperimentalGatewayAPISupport=true cert-manager jetstack/cert-manager
```

You should now have `cert-manager` running with nothing to do:

```console
$ kubectl wait --for=condition=Available deployment -n cert-manager --all
deployment.apps/cert-manager condition met
deployment.apps/cert-manager-cainjector condition met
deployment.apps/cert-manager-webhook condition met

$ kubectl get -n cert-manager deployment
NAME                      READY   UP-TO-DATE   AVAILABLE   AGE
cert-manager              1/1     1            1           42m
cert-manager-cainjector   1/1     1            1           42m
cert-manager-webhook      1/1     1            1           42m
```

## A Self-Signing Issuer

cert-manager can have any number of *issuer* configurations.
The simplest issuer type is [SelfSigned](https://cert-manager.io/docs/configuration/selfsigned/).
It simply takes the certificate request and signs it with the private key it generates for the TLS Secret.

```
Self-signed certificates don't provide any help in establishing trust between certificates.
However, they are great for initial testing, due to their simplicity.
```

To install self-signing, run

```console
$ kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned
spec:
  selfSigned: {}
EOF
```

## Creating a TLS Gateway Listener

We now have to patch the example Gateway to reference cert-manager:

```console
$ kubectl patch gateway/eg --patch-file - <<EOF
metadata:
  annotations:
    cert-manager.io/cluster-issuer: selfsigned
    cert-manager.io/common-name: "Hello World!"
spec:
  listeners:
  - name: https
    protocol: HTTPS
    hostname: www.example.com
    port: 443
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        name: eg-https
EOF
```

You could instead create a new Gateway serving HTTPS, if you'd prefer.
cert-manager doesn't care, but we'll keep it all together in this guide.

Nowadays, X.509 certificates don't use the subject Common Name for hostname matching, so you can set it to whatever you want, or leave it empty.
The important parts here are

* the annotation referencing the "selfsigned" ClusterIssuer we created above,
* the `hostname`, which is required (but see [#6051](https://github.com/cert-manager/cert-manager/issues/6051) for computing it based on attached HTTPRoutes), and
* the named Secret, which is what cert-manager will create for us.

The annotations are documented in [Supported Annotations](https://cert-manager.io/docs/usage/gateway/#supported-annotations).

Patching the Gateway makes cert-manager create a self-signed certificate within a few seconds.
Eventually, the Gateway becomes `Programmed` again:

```console
$ kubectl wait --for=condition=Programmed gateway/eg
gateway.gateway.networking.k8s.io/eg condition met
```

### Testing The Gateway

See [Testing in Secure Gateways](secure-gateways.md#testing) for the general idea.

Since we have a self-signed certificate, `curl` will by default reject it, requiring the `-k` flag:

```console
$ curl -kv -HHost:www.example.com https://127.0.0.1/get
...
* Server certificate:
*  subject: CN=Hello World!
...
< HTTP/2 200
...
```

### How cert-manager and Envoy Gateway Interact

*This explains [cert-manager Concepts](https://cert-manager.io/docs/concepts/) in an Envoy Gateway context.*

In the interaction between the two, cert-manager does all the heavy lifting.
It subscribes to changes to Gateway resources (using the [`gateway-shim` component](https://github.com/cert-manager/cert-manager/tree/master/pkg/controller/certificate-shim/gateways).)
For any Gateway it finds, it looks for any [TLS listeners](https://gateway-api.sigs.k8s.io/guides/tls/#listeners-and-tls), and the associated `tls.certificateRefs`.
Note that while Gateway API supports multiple refs here, Envoy Gateway only uses one.
cert-manager also looks at the `hostname` of the listener to figure out which hosts the certificate is expected to cover.
More than one listener can use the same certificate Secret, which means cert-manager needs to find all listeners using the same Secret before deciding what to do.
If the `certificatRef` points to a valid certificate, given the hostnames found in listeners, cert-manager has nothing to do.

If there is no valid certificate, or it is about to expire, cert-manager's `gateway-shim` creates a Certificate resource, or updates the existing one.
cert-manager then follows the [Certificate Lifecycle](https://cert-manager.io/docs/concepts/certificate/#certificate-lifecycle).
To know how to issue the certificate, an ClusterIssuer is configured, and referenced through annotations on the Gateway resource, which you did above.
Once a matching ClusterIssuer is found, that plugin does what needs to be done to acquire a signed certificate.

In the case of the ACME protocol (used by Let's Encrypt,) cert-manager can also use an HTTP Gateway to solve the HTTP-01 challenge type.
This is the other side of cert-manager's Gateway API support:
the [ACME issuer](https://github.com/cert-manager/cert-manager/tree/master/pkg/issuer/acme/http/httproute.go) creates a temporary [HTTPRoute](https://gateway-api.sigs.k8s.io/api-types/httproute/), lets the ACME server(s) query it, and deletes it again.

cert-manager then updates the Secret that the Gateway's listener points to in `tls.certificateRefs`.
Envoy Gateway picks up that the Secret has changed, and reloads the corresponding Envoy Proxy Deployments with the new private key and certificate.

As you can imagine, cert-manager requires quite broad permissions to update Secrets in any namespace, so the security-minded reader may want to look at the RBAC resources the Helm chart creates.

## Using the ACME Issuer With Let's Encrypt and HTTP-01

We will start using the Let's Encrypt staging environment, to spare their production environment.
Our Gateway already contains an HTTP listener, so we will use that for the HTTP-01 challenges.

```console
$ CERT_MANAGER_CONTACT_EMAIL=$(git config user.email)  # Or whatever...
$ kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: "$CERT_MANAGER_CONTACT_EMAIL"
    privateKeySecretRef:
      name: letsencrypt-staging-account-key
    solvers:
    - http01:
        gatewayHTTPRoute:
          parentRefs:
          - kind: Gateway
            name: eg
            namespace: default
EOF
```

The important parts are

* using `spec.acme` with a server URI and contact email address, and
* referencing our plain HTTP gateway so the challenge HTTPRoute is attached to the right place.

Check the account registration process using the Ready condition:

```console
$ kubectl wait --for=condition=Ready clusterissuer/letsencrypt-staging
$ kubectl describe clusterissuer/letsencrypt-staging
...
Status:
  Acme:
    Uri:                   https://acme-staging-v02.api.letsencrypt.org/acme/acct/123456789
  Conditions:
    Message:               The ACME account was registered with the ACME server
    Reason:                ACMEAccountRegistered
    Status:                True
    Type:                  Ready
...
```

Now we're ready to update the Gateway annotation to use this issuer instead:

```console
$ kubectl annotate --overwrite gateway/eg cert-manager.io/cluster-issuer=letsencrypt-staging
```

The Gateway should be picked up by cert-manager, which will create a new certificate for you, and replace the Secret.

You should see a new CertificateRequest to track:

```console
$ kubectl get certificaterequest
NAME             APPROVED   DENIED   READY   ISSUER                REQUESTOR                                         AGE
eg-https-xxxxx   True                True    selfsigned            system:serviceaccount:cert-manager:cert-manager   42m
eg-https-xxxxx   True                True    letsencrypt-staging   system:serviceaccount:cert-manager:cert-manager   42m
```

### Testing The Gateway

We still require the `-k` flag, since the Let's Encrypt staging environment CA is not widely trusted.

```console
$ curl -kv -HHost:www.example.com https://127.0.0.1/get
...
* Server certificate:
*  subject: CN=Hello World!
*  issuer: C=US; O=(STAGING) Let's Encrypt; CN=(STAGING) Ersatz Edamame E1
...
< HTTP/2 200
...
```

## Using The Let's Encrypt Production Environment

Changing to the production environment is just a matter of replacing the server URI:

```console
$ CERT_MANAGER_CONTACT_EMAIL=$(git config user.email)  # Or whatever...
$ kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory  # Removed "-staging".
    email: "$CERT_MANAGER_CONTACT_EMAIL"
    privateKeySecretRef:
      name: letsencrypt-account-key                         # Removed "-staging".
    solvers:
    - http01:
        gatewayHTTPRoute:
          parentRefs:
          - kind: Gateway
            name: eg
            namespace: default
EOF
```

And now you can update the Gateway listener to point to `letsencrypt` instead:

```console
$ kubectl annotate --overwrite gateway/eg cert-manager.io/cluster-issuer=letsencrypt
```

As before, track it by looking at CertificateRequests.

### Testing The Gateway

Once the certificate has been replaced, we should finally be able to get rid of the `-k` flag:

```console
$ curl -v -HHost:www.example.com --resolve www.example.com:127.0.0.1 https://www.example.com/get
...
* Server certificate:
*  subject: CN=Hello World!
*  issuer: C=US; O=Let's Encrypt; CN=R3
...
< HTTP/2 200
...
```

## Collecting Garbage

You probably want to set the `cert-manager.io/revision-history-limit` annotation on your Gateway to make cert-manager prune the CertificateRequest history.

cert-manager [deletes unused Certificate resources](https://github.com/cert-manager/cert-manager/blob/c5e6bf39d688d202592318eaf91988466a7ee37b/pkg/controller/certificate-shim/sync.go#L171), and they are updated in-place when possible, so there should be no need for cleaning up Certificate resources.
The deletion is based on whether a Gateway still holds a `tls.certificateRefs` that requires the Certificate.

If you remove a TLS listener from a Gateway, you may still have a Secret lingering.
cert-manager can clean it up using [a flag](https://cert-manager.io/docs/usage/certificate/#cleaning-up-secrets-when-certificates-are-deleted).

## Issuer Namespaces

We have used ClusterIssuer resources in this tutorial.
They are not bound to any namespace, and will read annotations from Gateways in any namespace.
You could also use [Issuer](https://cert-manager.io/docs/concepts/issuer/), which is bound to a namespace.
This is useful e.g. if you want to use different ACME accounts for different namespaces.

If you change the issuer kind, you also need to change the annotation key from `cert-manager.io/clusterissuer` to `cert-manager.io/issuer`.

## Using ExternalDNS

The [ExternalDNS](https://kubernetes-sigs.github.io/external-dns/v0.6.0/) controller maintains DNS records based on Kubernetes resources.
Together with cert-manager, this can be used to fully automate hostname management.
It can use various source resources, among them Gateway Routes.
Just specify a Gateway Route resource, let ExternalDNS create the domain records, and then cert-manager the TLS certificate.

[The tutorial on Gateway API](https://kubernetes-sigs.github.io/external-dns/v0.6.0/tutorials/gateway-api/) uses kubectl.
They also have a [Helm chart](https://github.com/kubernetes-sigs/external-dns/blob/master/charts/external-dns/README.md), which is easier to customize.
The only thing relevant to Envoy Gateway is to set the sources:

```yaml
# values.yaml
sources:
- gateway-httproute
- gateway-grpcroute
- gateway-tcproute
- gateway-tlsroute
- gateway-udproute
```

## Monitoring Progress / Troubleshooting

You can monitor progress in several ways:

The Issuer has a Ready condition (though this is rather [boring](https://github.com/cert-manager/cert-manager/blob/c5e6bf39d688d202592318eaf91988466a7ee37b/pkg/issuer/selfsigned/setup.go#L32) for the `selfSigned` type):

```console
$ kubectl get issuer --all-namespaces
NAMESPACE   NAME         READY   AGE
default     selfsigned   True    42m
```

The Gateway will say when it has an invalid certificate:

```console
$ kubectl describe gateway/eg
...
    Conditions:
      Message:               Secret default/eg-https does not exist.
      Reason:                InvalidCertificateRef
      Status:                False
      Type:                  ResolvedRefs
...
      Message:               Listener is invalid, see other Conditions for details.
      Reason:                Invalid
      Status:                False
      Type:                  Programmed
...
Events:
  Type     Reason     Age    From                       Message
  ----     ------     ----   ----                       -------
  Warning  BadConfig  42m    cert-manager-gateway-shim  Skipped a listener block: spec.listeners[1].hostname: Required value: the hostname cannot be empty
```

The main question is if cert-manager has picked up on the Gateway.
I.e., has it created a Certificate for it?
The above `describe` contains an event from `cert-manager-gateway-shim` telling you of one such issue.
Be aware that if you have a non-TLS listener in the Gateway, like we did, there will be events saying it is not eligible, which is of course expected.

Another option is looking at Deployment logs.
The cert-manager logs are not very verbose by default, but setting the Helm value `global.logLevel` to 6 will enable all debug logs (the default is 2.)
This will make them verbose enough to say why a Gateway wasn't considered (e.g. missing `hostname`, or `tls.mode` is not `Terminate`.)

```console
$ kubectl logs -n cert-manager deployment/cert-manager
...
```

Simply listing Certificate resources may be useful, even if it just gives a yes/no answer:

```console
$ kubectl get certificate --all-namespaces
NAMESPACE   NAME       READY   SECRET     AGE
default     eg-https   True    eg-https   42m
```

If there is a Certificate, then the `gateway-shim` has recognized the Gateway.
But is there a CertificateRequest for it?
(BTW, don't confuse this with a CertificateSigningRequest, which is a Kubernetes core resource type representing the same thing.)

```console
$ kubectl get certificaterequest --all-namespaces
NAMESPACE   NAME             APPROVED   DENIED   READY   ISSUER       REQUESTOR                                         AGE
default     eg-https-xxxxx   True                True    selfsigned   system:serviceaccount:cert-manager:cert-manager   42m
```

The ACME issuer also has `Order` and `Challenge` resources to watch:

```console
$ kubectl get order --all-namespaces -o wide
NAME                                                     STATE     ISSUER                REASON   AGE
order.acme.cert-manager.io/envoy-https-xxxxx-123456789   pending   letsencrypt-staging            42m

$ kubectl get challenge --all-namespaces
NAME                                                                    STATE     DOMAIN            AGE
challenge.acme.cert-manager.io/envoy-https-xxxxx-123456789-1234567890   pending   www.example.com   42m
```

Using `kubetctl get -o wide` or `kubectl describe` on the Challenge will explain its state more.

```console
$ kubectl get order --all-namespaces -o wide
NAME                                                     STATE   ISSUER                REASON   AGE
order.acme.cert-manager.io/envoy-https-xxxxx-123456789   valid   letsencrypt-staging            42m
```

Finally, since cert-manager creates the Secret referenced by the Gateway listener as its last step, we can also look for that:

```console
$ kubectl get secret secret/eg-https
NAME       TYPE                DATA   AGE
eg-https   kubernetes.io/tls   3      42m
```

## Clean Up

* Uninstall cert-manager: `helm uninstall --namespace cert-manager cert-manager`
* Delete the `cert-manager` namespace: `kubectl delete namespace/cert-manager`
* Delete the `https` listener from `gateway/eg`.
* Delete `secret/eg-https`.

## See Also

* [Secure Gateways](../secure-gateways/)
* [Securing gateway.networking.k8s.io Gateway Resources](https://cert-manager.io/docs/usage/gateway/)
