# Envoy Gateway

[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/envoyproxy/gateway/badge)](https://securityscorecards.dev/viewer/?uri=github.com/envoyproxy/gateway)
[![Build and Test](https://github.com/envoyproxy/gateway/actions/workflows/build_and_test.yaml/badge.svg)](https://github.com/envoyproxy/gateway/actions/workflows/build_and_test.yaml)
[![codecov](https://codecov.io/gh/envoyproxy/gateway/branch/main/graph/badge.svg)](https://codecov.io/gh/envoyproxy/gateway)
[![CodeQL](https://github.com/envoyproxy/gateway/actions/workflows/codeql.yml/badge.svg)](https://github.com/envoyproxy/gateway/actions/workflows/codeql.yml)
[![OSV-Scanner](https://github.com/envoyproxy/gateway/actions/workflows/osv-scanner.yml/badge.svg)](https://github.com/envoyproxy/gateway/actions/workflows/osv-scanner.yml)
[![Trivy](https://github.com/envoyproxy/gateway/actions/workflows/trivy.yml/badge.svg)](https://github.com/envoyproxy/gateway/actions/workflows/trivy.yml)

![Envoy Gateway Logo](https://github.com/cncf/artwork/blob/main/projects/envoy/envoy-gateway/horizontal/color/envoy-gateway-horizontal-color.svg)

Envoy Gateway is an open source project for managing Envoy Proxy as a standalone or
Kubernetes-based application gateway.
[Gateway API](https://gateway-api.sigs.k8s.io) resources are used to dynamically provision and configure the managed Envoy Proxies.

## Documentation

* [Blog][blog] introducing Envoy Gateway.
* [Goals](GOALS.md)
* [Quickstart](https://gateway.envoyproxy.io/latest/tasks/quickstart/) to use Envoy Gateway in a few simple steps.
* [Roadmap](https://gateway.envoyproxy.io/contributions/roadmap/)
* [Compatibility Matrix](https://gateway.envoyproxy.io/news/releases/matrix/)

## Contact

* [envoy-gateway-announce](https://groups.google.com/g/envoy-gateway-announce): Join our mailing list to receive important announcements.
* Slack: Join the [Envoy Slack workspace][] if you're not already a member. Otherwise, use the
  [Envoy Gateway channel][] to start collaborating with the community.

## Contributing

* [Code of conduct](/CODE_OF_CONDUCT.md)
* [Contributing guide](https://gateway.envoyproxy.io/contributions/contributing/)
* [Developer guide](https://gateway.envoyproxy.io/contributions/develop/)

## Security Reporting

If you've found a security vulnerability or a process crash, please follow the instructions in [SECURITY.md](./SECURITY.md) to submit a report.

## Community Meeting

Envoy Gateway contributor meetings are held on Thursdays and alternate weekly between EU-friendly and APAC-friendly times.

* EU-friendly: Thursdays at 7:00 AM PST / 10:00 AM EST / 4:00 PM CET. [Add to your calendar][eu-calendar]
* APAC-friendly: Thursdays at 4:00 PM PST / 7:00 PM EST / 8:00 AM CST. [Add to your calendar][apac-calendar]
* Full Envoy community calendar is available [here][community-calendar].
* Meeting notes and agenda: [Meeting details][meeting]

[meeting]: https://docs.google.com/document/d/1i5wa1VsxIbQw7jbWvGmvy8C4Zpp7SGV1aVViSLgqU4M/edit?usp=sharing
[eu-calendar]: https://calendar.google.com/calendar/event?action=TEMPLATE&tmeid=djQ1NWRhM2xjbG5vdGdpbXZibHM5MWxiOWdfMjAyNTEyMTFUMTUwMDAwWiA4YnNmYzhtMmRmOG1hZDJqdGdmNmY4b25qNEBn&tmsrc=8bsfc8m2df8mad2jtgf6f8onj4%40group.calendar.google.com&scp=ALL
[apac-calendar]: https://calendar.google.com/calendar/event?action=TEMPLATE&tmeid=MW80N2FmZDhyZmZoNnBrbTdnbnN1Y3M0dGJfMjAyNTEyMTlUMDAwMDAwWiA4YnNmYzhtMmRmOG1hZDJqdGdmNmY4b25qNEBn&tmsrc=8bsfc8m2df8mad2jtgf6f8onj4%40group.calendar.google.com&scp=ALL
[community-calendar]: https://calendar.google.com/calendar/u/0/embed?src=8bsfc8m2df8mad2jtgf6f8onj4@group.calendar.google.com&ctz=America/Chicago
[blog]: https://blog.envoyproxy.io/introducing-envoy-gateway-ad385cc59532
[Envoy Slack workspace]: https://communityinviter.com/apps/envoyproxy/envoy
[Envoy Gateway channel]: https://envoyproxy.slack.com/archives/C03E6NHLESV
