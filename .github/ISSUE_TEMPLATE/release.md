---
name: Release
about: Suggest a release for this project
title: ''
labels: release-process
assignees: ''

---

- [ ] release notes
- [ ] cherry pick
- [ ] bump [envoy](https://hub.docker.com/r/envoyproxy/envoy/tags),[ratelimit](https://hub.docker.com/r/envoyproxy/ratelimit/tags) image tag if needed
- [ ] update VERSION in release branch
- [ ] wait for CI
- [ ] push tag
    - [ ] Push tag https://github.com/envoyproxy/gateway/releases/tag/v1.x.x
    - [ ] wait for release CI
- [ ] verify quickstart
- [ ] update doc
- [ ] update release [announcement](https://github.com/envoyproxy/gateway/releases/tag/v1.x.x) 
- [ ] GH Release, Slack announcement, [google group](https://groups.google.com/g/envoy-gateway-announce) announcement