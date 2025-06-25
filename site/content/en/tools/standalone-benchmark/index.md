---
title: "Performance Benchmark Report Explorer"
description: "Envoy Gateway Performance Benchmark Report Explorer Tool"
type: "tools"
includeBenchmark: true
---

<style>
  .standalone-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  }

  .standalone-title {
    font-size: 2.5rem;
    font-weight: 600;
    color: #333;
    margin-bottom: 1rem;
  }

  .standalone-description {
    font-size: 1.25rem;
    color: #666;
    margin-bottom: 2rem;
    line-height: 1.6;
  }
</style>
<div class="standalone-container">
  <h1 class="standalone-title">Performance Benchmark Report Explorer</h1>
  <p class="standalone-description">
    Explore benchmark results from Envoy Gateway Releleases. The test code is open source and available for you to run and contribute to.
  </p>
  <p class="standalone-description">Curious to learn more? Join the conversation in <code>#gateway-users</code> channel in <a href="https://communityinviter.com/apps/envoyproxy/envoy">Envoy Slack</a></p>

  {{< benchmark-dashboard
    version="latest"
    showHeader="false"
    class="mt-4"
  >}}
</div>
