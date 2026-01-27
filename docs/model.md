---
layout: page
nav-title: Model
---

# The Gemara Model

Integrating Governance, Risk, and Compliance (GRC) into software development pipelines presents a formidable challenge. 

Traditional, often manual, approaches to GRC are ill-suited for the pace of contemporary development. This has given rise to the discipline of GRC Engineering, which strategically applies engineering principles to GRC processes to make them more efficient and integrated. Its ultimate goal is to achieve Automated Governance, where compliance tracking is embedded throughout the deployment pipeline, acting as a required quality gate before code reaches production.

The need for a structured approach to this challenge was a key observation during the creation of the CNCF's Automated Governance Maturity Model (AGMM). The authors of that document observed that in a fully automated governance program, maturity can be measured in at least four different areas: "Policy, Evaluation, Enforcement, and Audit." This foundational insight provides a lexicon for describing the core activities of a secure software factory. The Gemara model took shape as industry projects began to apply this lexicon.

The FINOS Common Cloud Controls (CCC) and the OpenSSF's Open Source Project Security Baseline (OSPSB) projects adopted the AGMM's language but identified the need for greater granularity. They added two more conceptual areas to distinguish between high-level, abstract recommendations and technology-specific objectives. This practical expansion from four concepts to six areas that build upon each other formed the genesis of Gemara's layered model.

The Gemara model was created to codify these concepts. As the GRC Engineering Model for Automated Risk Assessment, its core purpose is to provide a logical model to describe the categories of compliance activities, how they interact, and how to enable automated interoperability between them.

## At a Glance

Gemara organizes compliance activities into seven categorical layers, each building upon the previous. 

Layers 1 through 3 provide definitions to inform the execution of sensitive activities, while layers 5 through 8 provide measurements to inform next steps.

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

This model is intentionally stable. Changes are rare and require significant community discussion, as the model reflects fundamental organizational patterns in GRC activities. As noted in ADR-0008, the model is closed to modification. This enables downstream resources to build upon the foundational concepts without concern about essential terms shifting later.

---

## Continue Reading

- **> Next Page**: [Scope](./model/01-scope.html)

---