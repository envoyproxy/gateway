# Security Policy

## Reporting a Vulnerability or Crash

Please report suspected vulnerabilities privately.

To report a security issue:

1. **Do not open a public issue** on the GitHub repository to disclose a vulnerability.
2. Send an email to our security team at [envoy-gateway-security@googlegroups.com](mailto:envoy-gateway-security@googlegroups.com).
3. Include the following details in your email:
    - Description of the vulnerability.
    - Steps to reproduce it.
    - Potential impact.
    - Suggested fixes or patches, if any.

We aim to respond within **48 hours**. If the issue is confirmed, we will coordinate the release timeline with you and credit you if applicable and with your consent.

## Vulnerability Review and Fix Process

The Envoy Gateway security team at [envoy-gateway-security@googlegroups.com](mailto:envoy-gateway-security@googlegroups.com) reviews reports privately. The team checks scope, validates the issue, asks for details when needed, and identifies affected versions and impact.

Accepted vulnerabilities are tracked in a draft GitHub Security Advisory after initial validation. The advisory tracks affected versions, severity, CVSS assessment, credits, fix status, and disclosure plan. When appropriate, the team requests or attaches a CVE before public disclosure.

The team assigns fixes to security team members or Envoy Gateway maintainers with the right expertise. Fixes for accepted, undisclosed vulnerabilities are developed privately until coordinated disclosure, using a GitHub Security Advisory temporary private fork when appropriate, and coordinated with release managers for affected active branches.

Once fixes are ready for supported versions, we publish the advisory and announce the patched releases through the channels below.

## Public Vulnerability Disclosure

The Envoy Gateway security team and release managers coordinate public disclosure for affected release branches.

We announce advisories and patched releases through:

- [GitHub Security Advisories](https://github.com/envoyproxy/gateway/security/advisories)
- The [GitHub Releases page](https://github.com/envoyproxy/gateway/releases)
- The `#gateway-users` channel in the [Envoy Slack workspace](https://communityinviter.com/apps/envoyproxy/envoy)
- The [envoy-gateway-announce mailing list](https://groups.google.com/g/envoy-gateway-announce)

Security fixes are merged into active, non-EOL release branches as patch releases when the affected versions are still supported. See the [release matrix](https://gateway.envoyproxy.io/news/releases/matrix/) for support windows and the [patch release process](https://gateway.envoyproxy.io/community/releasing/#patch-release) for release mechanics.

Subscribe to these channels for security updates.

## Best Practices for Secure Usage

To reduce risk:

- Use the latest supported version of Envoy Gateway.
- Regularly monitor for updates and apply patches promptly.

## Contact

Questions? Email [envoy-gateway-security@googlegroups.com](mailto:envoy-gateway-security@googlegroups.com).
