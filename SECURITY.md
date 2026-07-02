# Security Policy

## Reporting a Vulnerability or a Crash

Please report suspected vulnerabilities privately.

### What to report

In scope:

- Crashes or denial-of-service in Envoy Gateway itself.
- Remote code execution, privilege escalation, or authentication/authorization bypass in the control plane.
- Information disclosure of secrets or other sensitive data handled by Envoy Gateway.

Out of scope (please report elsewhere):

- Vulnerabilities in upstream Envoy proxy — report to [envoy-security@googlegroups.com](mailto:envoy-security@googlegroups.com).
- Issues in third-party dependencies — report to the respective project.
- Findings that only affect a misconfigured or self-managed deployment.

### How to report

1. **Do not open a public issue** on the GitHub repository to disclose a vulnerability.
2. Report privately through one of:
    - GitHub's [private vulnerability reporting](https://github.com/envoyproxy/gateway/security/advisories/new) (preferred).
    - Email to our security team at [envoy-gateway-security@googlegroups.com](mailto:envoy-gateway-security@googlegroups.com).
3. Include the following details:
    - Description of the vulnerability.
    - Steps to reproduce it.
    - Potential impact.
    - Suggested fixes or patches, if any.

### Responsible disclosure

When investigating, please act in good faith: do not access, modify, or exfiltrate data that is not yours, and do not degrade or disrupt running services. We will not pursue legal action against researchers who follow this policy and report in good faith.

We aim to **acknowledge receipt within 48 hours** (calendar time); this is an acknowledgment, not a full assessment or fix. If the issue is confirmed, we will coordinate the release timeline with you and, with your consent, credit you in the advisory.

## Vulnerability Review and Fix Process

The Envoy Gateway security team reviews reports privately. The team checks scope, validates the issue, asks for details when needed, and identifies affected versions and impact.

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

## Best Practices for Secure Usage

To reduce risk:

- Use the latest supported version of Envoy Gateway.
- Regularly monitor for updates and apply patches promptly.
- Subscribe to the [envoy-gateway-announce mailing list](https://groups.google.com/g/envoy-gateway-announce) for security updates.

## Contact

Questions? Email [envoy-gateway-security@googlegroups.com](mailto:envoy-gateway-security@googlegroups.com).
