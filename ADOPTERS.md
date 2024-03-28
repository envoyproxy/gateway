
<!--

Insert your entry using this template keeping the list alphabetically sorted:

## <Company/Organization Name>
 * Website: https://www.your-website.com
 * Category: End User, Service Provider, etc
 * Environments: AWS, Azure, Google Cloud, Bare Metal, etc
 * Use Cases:
    - ...
 * Status:
   - [ ] development & testing
   - [ ] production
 * (Option) Logo (show in the official site):
 * (Option) Description:
-->

# Envoy Gateway Adopters

This page contains a list of organizations who are users of Envoy Gateway, following the [definitions provided by the CNCF](https://github.com/cncf/toc/blob/main/FAQ.md#what-is-the-definition-of-an-adopter).

If you would like to be included in this table, please submit a PR to this file or comment to [this issue](https://github.com/envoyproxy/gateway/issues/2781) and your information will be added.

## AllFactors
* Website https://allfactors.com
* Category: End User
* Environments:
* Use Case:
   - Routing all customer traffic to our various backends. Every time a new customer signs up we dynamically add a
     route to a new hostname so Envoy Gateway is deeply integrated with our product.
* Status: production
* Logo: https://allfactors.com/AllFactors-Logo.svg

## Tetrate
* Website: https://www.tetrate.io
* Category: Service Provider
* Environments: AWS
* Use Cases:
   - Tetrate provides Enterprise Gateway (TEG) to end users, which includes a 100% upstream distribution of Envoy Gateway, and management to deliver applications securely, authenticate user traffic, protect services with rate limiting and WAF, and integrate with your observability stack to monitor and observe activity.
* Status: production
* (Option) https://tetrate.io/wp-content/uploads/2023/03/tetrate-logo-dark.svg
* (Option) Description:

## Airspace Link
* Organizatioin: Airspace Link
* Website: https://airspacelink.com/
* Category: End User
* Environments: Azure
* Use Cases:
    - Airspace Link is using Envoy Gateway to route all public APIs to Kubernetes clusters, developers are manipulating routes descriptions using agnostic manifest files, which are then automatically provisioned using Envoy Gateway.
* Status: production
* Logo: https://airhub.airspacelink.com/images/asl-flat-logo.png
