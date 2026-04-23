---
title: "Performance Benchmark Report Explorer"
description: "Explore the Benchmark Reports from Envoy Gateway Releases"
type: "tools"
includeBenchmark: true
---

<style>
  .bt-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  }

  .bt-title {
    font-size: 3.5rem;
    font-weight: 800;
    margin-bottom: 2.5rem;
    background: linear-gradient(135deg, #9333EA 0%, #4F46E5 100%);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
    line-height: 1.4;
    letter-spacing: -0.02em;
  }

  .bt-description {
    font-size: 1.25rem;
    color: #666;
    margin-bottom: 2rem;
    line-height: 1.6;
  }
</style>

<div class="bt-container">
  <h1 class="bt-title">Performance Benchmark Report Explorer</h1>
  <p class="bt-description">
    Explore benchmark results from Envoy Gateway Releleases. The test code is open source and available for you to run and contribute to.
  </p>
  <p class="bt-description">Curious to learn more? Join the conversation in <code>#gateway-users</code> channel in <a href="https://communityinviter.com/apps/envoyproxy/envoy">Envoy Slack</a></p>

  {{< benchmark-dashboard
    version="latest"
    showHeader="false"
    class="mt-4"
  >}}
</div>
