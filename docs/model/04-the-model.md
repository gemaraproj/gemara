---
layout: page
title: The Model
---

### Introduction to the Model

A fundamental principle of this architecture is that each layer builds upon the one below it. Higher-level layers leverage the outputs and services of lower layers to perform their functions.

For instance, an Evaluation at Layer 5 is designed to test for conformance with a Policy defined at Layer 3, which is in turn informed by Controls at Layer 2. This hierarchical dependency creates a clear logical flow of information and authority through the Governance lifecycle. 

This structure reverses the traditional document-centric view with Policy building upon Control. Since many organizations start with a Policy which includes Controls, the idea that Policy would follow Controls may seem counter-intuitive. However, this example is demonstrative of the model: A Policy is not ready for implementation and Evaluation until the Controls are written.

## The Complete Model 

The model is organized into two primary categories of activity: definitions and measurements. These categories will, ideally, each be represented by documents or logs pertaining to the corresponding action. 
Activities in the definition layers, Layer 1 through Layer 3, should each produce document assets that may be referenced by higher layers or within their own layer. Activities in the measurement layers, Layer 5 through Layer 7, should each produce timestamped logs as outputs.

{% include gemara-layer-diagram.html %}

As noted in [Foundational Concepts](./03-foundational-concepts), Layer 4 defines Sensitive Activities connecting the two halves. The first three layers point toward the Sensitive Activities to define what is acceptable and what is unacceptable. Meanwhile, the last three layers look back at the Sensitive Activities or their outcomes to determine whether they comply with defined expectations. In Layer 7, Audit activities determine the quality of results from the lower layers.

The model's architectural components and their elements are detailed in the following sections, starting with [The Definition Layers](./05-definition-layers), followed by [The Pivot Point](./06-sensitive-activities) examining sensitive activities, and finally [The Measurement Layers](./07.1-measurement-layers).

{% include themed-image.html
  light="/assets/model-images/figure-4.1-light.png"
  dark="/assets/model-images/figure-4.1-dark.png"
  alt="Figure 4.1: Model Relationships and the Logical Flow"
  caption="Figure 4.1: Model Relationships and the Logical Flow"
  width="60%"
%}

---

## Continue Reading

- **< Previous Page**: [Foundational Concepts](./03-foundational-concepts)
- **> Next Page**: [The Definition Layers](./05-definition-layers)

---
