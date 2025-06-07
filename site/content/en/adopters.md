+++
title = "Adopters"
linktitle = "Adopters"
description = "Organizations using Envoy Gateway in production"
weight = 10
+++

{{% blocks/cover title="Envoy Gateway Adopters" height="auto" color="primary" %}}
<div class="container">
<p class="lead">
Join the growing community of organizations trusting Envoy Gateway for their production traffic management needs.
</p>
</div>
{{% /blocks/cover %}}

{{% blocks/section color="white" %}}
<div class="row justify-content-center">
<div class="col-md-10">
<h2 class="text-center mb-5">Organizations Using Envoy Gateway</h2>
<p class="text-center mb-5">
These organizations are successfully running Envoy Gateway in production environments, from startups to enterprises across various industries.
</p>
</div>
</div>

<div class="adopters-grid">
  <!-- Example adopter logos would go here -->
  <div class="adopter-placeholder">
    <div class="placeholder-content">
      <i class="fas fa-building fa-3x mb-3"></i>
      <h4>Your Organization</h4>
      <p>Join the adopters list!</p>
    </div>
  </div>
</div>
{{% /blocks/section %}}

{{% blocks/section color="light" %}}
<div class="row justify-content-center">
<div class="col-md-8 text-center">
<h2 class="mb-4">
  <i class="fas fa-plus-circle me-3"></i>Add Your Logo Here!
</h2>
<p class="lead mb-4">
Are you using Envoy Gateway in production? We'd love to showcase your organization and help others learn from your success!
</p>

<div class="cta-buttons">
<a class="btn btn-lg btn-primary me-3" href="https://github.com/envoyproxy/gateway/blob/main/ADOPTERS.md">
  <i class="fas fa-file-alt me-2"></i>View Adopters List
</a>
<a class="btn btn-lg btn-outline-primary" href="https://github.com/envoyproxy/gateway/edit/main/ADOPTERS.md">
  <i class="fas fa-edit me-2"></i>Add Your Organization
</a>
</div>

<div class="mt-5">
<h3>How to Add Your Organization</h3>
<div class="row mt-4">
  <div class="col-md-4">
    <div class="step-card">
      <div class="step-number">1</div>
      <h4>Prepare Your Info</h4>
      <p>Gather your organization name, logo (SVG preferred), website URL, and a brief description of how you use Envoy Gateway.</p>
    </div>
  </div>
  <div class="col-md-4">
    <div class="step-card">
      <div class="step-number">2</div>
      <h4>Submit a PR</h4>
      <p>Edit the <code>ADOPTERS.md</code> file in our GitHub repository and submit a pull request with your organization details.</p>
    </div>
  </div>
  <div class="col-md-4">
    <div class="step-card">
      <div class="step-number">3</div>
      <h4>Get Featured</h4>
      <p>Once merged, your organization will be featured on this page and help inspire other adopters!</p>
    </div>
  </div>
</div>
</div>
</div>
</div>
{{% /blocks/section %}}

{{% blocks/section color="white" %}}
<div class="row justify-content-center">
<div class="col-md-10">
<h2 class="text-center mb-5">Why Share Your Adoption Story?</h2>

<div class="benefits-grid">
  <div class="benefit-card">
    <div class="icon-container">
      <i class="fas fa-users"></i>
    </div>
    <h3>Build Community</h3>
    <p>Help build a stronger Envoy Gateway community by sharing your success story and inspiring others to adopt the technology.</p>
  </div>

  <div class="benefit-card">
    <div class="icon-container">
      <i class="fas fa-star"></i>
    </div>
    <h3>Gain Recognition</h3>
    <p>Showcase your organization as an innovative technology adopter and thought leader in cloud-native networking.</p>
  </div>

  <div class="benefit-card">
    <div class="icon-container">
      <i class="fas fa-handshake"></i>
    </div>
    <h3>Connect & Learn</h3>
    <p>Connect with other adopters, share best practices, and learn from real-world production experiences.</p>
  </div>

  <div class="benefit-card">
    <div class="icon-container">
      <i class="fas fa-chart-line"></i>
    </div>
    <h3>Drive Innovation</h3>
    <p>Your feedback and use cases help shape the future development of Envoy Gateway features and capabilities.</p>
  </div>
</div>
</div>
</div>
{{% /blocks/section %}}

{{% blocks/section color="light" %}}
<div class="row justify-content-center">
<div class="col-md-8 text-center">
<h2 class="mb-4">Questions About Adopting Envoy Gateway?</h2>
<p class="mb-4">
Check out our resources to help with your evaluation and adoption journey:
</p>

<div class="resource-links">
<a href="/docs/evaluator-guide/" class="btn btn-outline-secondary me-2 mb-2">
  <i class="fas fa-clipboard-list me-2"></i>Evaluator Guide
</a>
<a href="/docs/concepts/reference-architecture/" class="btn btn-outline-secondary me-2 mb-2">
  <i class="fas fa-sitemap me-2"></i>Reference Architecture
</a>
<a href="/docs/tasks/quickstart/" class="btn btn-outline-secondary me-2 mb-2">
  <i class="fas fa-rocket me-2"></i>Quick Start
</a>
<a href="https://github.com/envoyproxy/gateway/discussions" class="btn btn-outline-secondary me-2 mb-2">
  <i class="fas fa-comments me-2"></i>Community Discussions
</a>
</div>
</div>
</div>
{{% /blocks/section %}}

<style>
.adopters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 2rem;
  margin: 2rem 0;
}

.adopter-placeholder {
  border: 2px dashed #ddd;
  border-radius: 8px;
  padding: 2rem;
  text-align: center;
  min-height: 150px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.placeholder-content {
  color: #666;
}

.cta-buttons {
  margin: 2rem 0;
}

.step-card {
  text-align: center;
  padding: 1.5rem;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  margin-bottom: 1rem;
  height: 100%;
}

.step-number {
  background: #007bff;
  color: white;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  margin: 0 auto 1rem;
}

.benefits-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 2rem;
  margin: 2rem 0;
}

.benefit-card {
  text-align: center;
  padding: 2rem;
}

.icon-container {
  background: #f8f9fa;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1rem;
  color: #007bff;
  font-size: 1.5rem;
}

.resource-links {
  margin: 2rem 0;
}
</style>
