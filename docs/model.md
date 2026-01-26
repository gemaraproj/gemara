---
layout: page
title: The Gemara Model
---

**Status**: <span class="badge badge-stable">Stable</span>

The Gemara Model describes seven categorical layers of GRC (Governance, Risk, Compliance) activities, representing how GRC activities are organized and interact.

## The Seven Layers

Gemara organizes compliance activities into seven categorical layers, each building upon the previous:

<div class="gemara-layer-diagram">
  <div class="layer-banner layer-7">
    <span class="layer-number">7</span>
    <div class="layer-content">
      <div class="layer-title">Audit & Continuous Monitoring</div>
      <div class="layer-description">Efficacy review of all previous outputs</div>
    </div>
  </div>
  <div class="layer-banner layer-6">
    <span class="layer-number">6</span>
    <div class="layer-content">
      <div class="layer-title">Preventive & Remediative Enforcement</div>
      <div class="layer-description">Corrective actions for noncompliance</div>
    </div>
  </div>
  <div class="layer-banner layer-5">
    <span class="layer-number">5</span>
    <div class="layer-content">
      <div class="layer-title">Intent & Behavior Evaluation</div>
      <div class="layer-description">Inspection of sensitive activities</div>
    </div>
  </div>
  <div class="layer-banner layer-4">
    <span class="layer-number">4</span>
    <div class="layer-content">
      <div class="layer-title">Sensitive Activities</div>
      <div class="layer-description">Actions that might introduce risk</div>
    </div>
  </div>
  <div class="layer-banner layer-3">
    <span class="layer-number">3</span>
    <div class="layer-content">
      <div class="layer-title">Risk & Policy</div>
      <div class="layer-description">Organization-specific rules</div>
    </div>
  </div>
  <div class="layer-banner layer-2">
    <span class="layer-number">2</span>
    <div class="layer-content">
      <div class="layer-title">Threats & Controls</div>
      <div class="layer-description">Technology-specific objectives</div>
    </div>
  </div>
  <div class="layer-banner layer-1">
    <span class="layer-number">1</span>
    <div class="layer-content">
      <div class="layer-title">Vectors & Guidance</div>
      <div class="layer-description">Foundational knowledge or regulations</div>
    </div>
  </div>
</div>

## Model Stability

This model is intentionally stable. Changes are rare and require significant community discussion, as the model reflects fundamental organizational patterns in GRC activities.

**Why Stability Matters:**
- Provides a consistent foundation for all Gemara work
- Allows the Lexicon and Implementation to evolve independently
- Ensures long-term compatibility

## Relationship to Other Components

### [The Lexicon](../lexicon)
Provides definitions for terms used within each layer. The Model describes structure; the Lexicon provides shared vocabulary.

### [The Implementation](../implementation)
Provides schemas and SDKs based on the Model. The Model describes conceptual layers; the Implementation provides machine-readable formats and APIs.

