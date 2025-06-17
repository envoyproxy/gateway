---
title: "Follow the Code: The Journey of Quality in Envoy Gateway"
date: 2024-03-21
author: "Envoy Gateway Team"
description: "A deep dive into how Envoy Gateway ensures code quality through its comprehensive testing and validation process"
---

Have you ever wondered how open source projects like Envoy Gateway maintain high quality standards while evolving rapidly? Today, we're taking you behind the scenes to show you how code makes its way from an idea to a production-ready feature in Envoy Gateway.

**Built in the Open, Built Together** - this is our philosophy. Every test, every validation, every quality check is transparent and open for the community to see, contribute to, and improve upon.

## The Journey Begins

Every feature in Envoy Gateway starts as an idea - perhaps a new capability, a performance improvement, or a bug fix. But before any code makes it into the project, it must pass through a rigorous quality assurance process that ensures reliability, performance, and compatibility. Let's follow this journey together and see how you can be part of it.

## The Comprehensive Quality Assurance Pipeline

### 1. Development Phase
When a developer has an idea to improve Envoy Gateway, the journey begins with a well-thought-out proposal. During development:
- Code is written following the project's strict coding standards
- Unit tests are written alongside the code
- Local testing ensures basic functionality
- Documentation is prepared

> **🤝 Join Us:** Our coding standards and development practices are completely open. Check out our [contributor guide](https://gateway.envoyproxy.io/latest/contributing/) and start contributing today!

### 2. Pull Request Phase: Where the Magic Happens
The real magic happens when a pull request is submitted. This triggers our comprehensive automated testing suite - all visible and auditable by the community:

#### Conformance Tests
Our conformance tests ensure that Envoy Gateway remains a faithful implementation of the Gateway API specification. These tests:
- Validate that the implementation meets the standard requirements
- Test both experimental and stable features
- Help maintain compatibility with the broader Kubernetes ecosystem

#### End-to-End (E2E) Tests
E2E tests are the workhorses of our quality assurance process. They:
- Test the complete system in a real-world environment
- Validate gateway infrastructure resources
- Ensure HTTP routing works as expected
- Verify TLS configurations
- Test rate limiting functionality
- Validate WebAssembly (Wasm) features
- Check zone-aware routing capabilities

#### CEL Validation Tests
Our Common Expression Language (CEL) validation tests ensure policy correctness:
- Validate security policies
- Test backend traffic policies
- Verify client traffic policies
- Ensure extension policies work correctly
- Test complex routing rules and filters

#### Benchmark Tests
Performance is crucial for a gateway solution. Our benchmark tests:
- Measure the performance impact of changes
- Generate detailed benchmark reports
- Track metrics like route propagation speed and response times
- Ensure changes don't degrade system performance

#### Resilience Tests
A gateway must be reliable. Our resilience tests:
- Verify system behavior under stress
- Validate error handling and recovery
- Ensure system stability
- Test graceful degradation

#### Fuzz Testing
Security is paramount. Our fuzz testing:
- Runs continuously on Google's OSS-Fuzz platform
- Tests for security vulnerabilities
- Validates XDS configuration handling
- Ensures robustness against malformed inputs

#### Helm Deployment Tests
Real-world deployment scenarios are tested through:
- Gateway CRDs Helm chart validation
- Gateway controller Helm chart testing
- Add-ons Helm chart verification
- Multi-environment deployment testing

> **🔍 Transparency in Action:** All our test results are public! You can see every test run, every failure, and every fix in our GitHub Actions. Nothing is hidden - that's what "Built in the Open" means to us.

## Try It Yourself: Run the Tests Locally

Here's the beauty of open source - you don't have to take our word for it! You can clone our repository and run every single test yourself. This level of transparency means you can:

### 🏃‍♂️ Run the Complete Test Suite

```bash
# Clone the repository
git clone https://github.com/envoyproxy/gateway.git
cd gateway

# Run all tests
make test

# Run specific test types
make test.e2e           # End-to-end tests
make test.conformance   # Conformance tests
make test.benchmark     # Benchmark tests
make go.test.fuzz       # Fuzz tests
```

### 🔬 Explore Individual Test Areas

Want to dive deeper into a specific area? You can run tests for individual components:

```bash
# Run CEL validation tests
make test.cel-validation

# Run resilience tests
make test.resilience

# Run Helm chart tests
make test.helm
```

### 📊 Generate Your Own Benchmark Reports

Curious about performance? Generate the same benchmark reports we use:

```bash
# Run performance benchmarks
make benchmark

# View detailed performance metrics
cat test/benchmark/benchmark_report/latest_report.md
```

### 🐛 Test Your Own Scenarios

Found an edge case? Add your own test and run it:

```bash
# Add your test case to the appropriate test directory
# Then run it to see if it passes
go test -v ./test/e2e/tests -run YourTestName
```

> **💡 Real Transparency:** This isn't just marketing - every test we run in CI/CD, you can run on your laptop. Same code, same tests, same results. That's what "Built in the Open" really means.

### 3. Release Phase: Community-Driven Quality
Before a release is cut, we perform additional quality checks with full community visibility:

#### Comprehensive Test Suite
- All tests are run in sequence
- Different environment configurations are tested
- Various Kubernetes versions are validated
- Multiple deployment scenarios are verified

#### Performance Validation
- Benchmark results are compared against previous releases
- Performance regressions are identified and addressed
- Resource usage patterns are analyzed

#### Documentation Review
- API documentation is updated
- Configuration examples are verified
- Migration guides are prepared if needed

> **📊 See It All:** Our benchmark reports and test results are publicly available. Check out our latest performance metrics and see how we're constantly improving!

## Why This Matters to You

For users of Envoy Gateway, this rigorous process means:
- **Reliable production deployments** - because every change is thoroughly tested
- **Predictable behavior** - because our test suite covers real-world scenarios
- **Consistent performance** - because we benchmark every change
- **Easy upgrades** - because we test compatibility extensively
- **Confidence in the software** - because you can see exactly how we ensure quality

## Built Together: Your Role in Our Quality Journey

We invite you to be part of this journey. The beauty of open source is that quality is a community effort:

### 🚀 Ways to Contribute to Our Test Suite

**For Developers:**
- Add new test cases for edge cases you encounter
- Improve existing test coverage
- Contribute performance benchmarks
- Add fuzz testing scenarios

**For Users:**
- Report issues with detailed reproduction steps
- Share your real-world configuration examples
- Suggest new test scenarios based on your use cases

**For DevOps Engineers:**
- Contribute Helm chart tests
- Add deployment scenario tests
- Share infrastructure-specific test cases

**For Security Researchers:**
- Contribute to our fuzz testing suite
- Report security vulnerabilities
- Add security-focused test scenarios

### 🔗 Start Contributing Today

Ready to join our quality journey? Here's how:

1. **Explore our test suite**: Browse through /test directory in our [GitHub repository](https://github.com/envoyproxy/gateway)
2. **Pick a test area**: Choose from E2E, benchmark, conformance, resilience, or fuzz testing
3. **Start small**: Add a test case or improve documentation
4. **Join the conversation**: Participate in our community discussions about testing strategies

> **💡 Pro Tip:** Not sure where to start? Look for "good first issue" labels in our GitHub repository. Many are related to improving our test coverage!

## Transparency Through Collaboration

What sets Envoy Gateway apart is our commitment to transparency:

- **Open test results**: Every test run is visible
- **Public discussions**: All quality decisions are made in the open
- **Collaborative improvement**: Community feedback shapes our testing strategy
- **Shared ownership**: Everyone can contribute to quality

## Conclusion

The extensive test suite and quality checks make Envoy Gateway a production-ready solution that users can trust for their critical infrastructure needs. But more importantly, this commitment to quality is built on transparency and collaboration.

When you use Envoy Gateway, you're not just using software - you're joining a community that believes in building quality together, in the open, with full transparency.

**Ready to be part of our quality journey?**

🌟 **Start here**: [Contributing Guide](https://gateway.envoyproxy.io/latest/contributing/)
🧪 **Explore tests**: [Test Directory on GitHub](https://github.com/envoyproxy/gateway/tree/main/test)
💬 **Join discussions**: [Community Slack](https://envoyproxy.slack.com/)
🐛 **Report issues**: [GitHub Issues](https://github.com/envoyproxy/gateway/issues)

---

*Built in the Open, Built Together - this is how we ensure Envoy Gateway remains a reliable, high-performance gateway solution for everyone.*
