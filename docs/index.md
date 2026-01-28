---
layout: home
title: Home
---

# Gemara <span class="pronunciation">(Juh-MAH-ruh)</span>

<img src="{{ '/assets/gemara-logo.png' | relative_url }}" alt="Gemara Logo" class="gemara-logo" />

**GRC Engineering Model for Automated Risk Assessment**

Gemara provides a logical model to describe the categories of compliance activities, how they interact, and the schemas to enable automated interoperability between them.

In order to better facilitate cross-functional communication, the Gemara Model seeks to outline the categorical layers of activities related to automated governance.

<!--
## Quick Start

- **New to Gemara?** Start with our [About page](/about) to understand the model
- **Want to dive deeper?** Explore the [Seven Layers](/layers) of the model
- **Ready to build?** Check out our [Tutorial](/tutorial) for a hands-on example
- **Want to contribute?** See our [Contributing Guide](/contributing)
-->

## The Three Components

Gemara delivers three core components that work together to support automated GRC:

<div class="component-grid">
  <a href="./model/" class="component-card">
      <h2>The Model</h2>
      <p class="component-description">
        The foundational layer model that describes the seven categorical layers of GRC activities. 
        This model is <strong>stable and rarely changes</strong>, as it reflects the longstanding 
        reality of GRC activity types.
      </p>
      <p class="component-content">
        Provides the conceptual framework for understanding how different types of compliance 
        activities relate to each other. Establishes the seven layers: Guidance, Controls, Policy, 
        Evaluation, Enforcement, and Audit.
      </p>
  </a>

  <a href="./schema/" class="component-card">
      <h2>The Schemas</h2>
      <p class="component-description">
        Schemas (CUE format) that standardize the expression of elements in the model.
      </p>
      <p class="component-content">
        Provides CUE schemas for validation across all layers. Enables automated validation and 
        interoperability between tools.
      </p>
  </a>

  <a href="./sdk/" class="component-card">
     <h2>The SDKs</h2>
     <p class="component-description">
        Language-specific SDKs that provide programmatic access to Gemara documents and tooling 
        to accelerate automated tool development.
      </p>
      <p class="component-content">
        Currently provides Go SDK for reading, writing, and manipulating Gemara documents. 
        Active development area.
      </p>
  </a>
</div>


## Quick Start

Choose your starting point based on your needs:

- **Understanding GRC structure?** Start with **[The Model](./model)** component
- **Validating documents?** Use **[The Schemas](./schema/)** component
- **Building tools?** Jump to **[The SDKs](./sdk/)** component

All three components work together - you'll likely use elements from each as you work with Gemara.

## Real-World Usage

Gemara is being used today in production environments:

- **[FINOS Common Cloud Controls](https://www.finos.org/common-cloud-controls-project)** - Layer 2 controls for cloud environments
- **[Open Source Project Security Baseline](https://baseline.openssf.org/)** - Layer 2 security baseline for open source projects
- **[Privateer](https://github.com/privateerproj/privateer)** - Layer 5 evaluation framework with plugins like the [OSPS Baseline Plugin](https://github.com/revanite-io/pvtr-github-repo)
