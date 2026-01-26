---
layout: page
title: The Model
---

## Introduction to the Model

A fundamental principle of this architecture is that each layer builds upon the one below it. Higher-level layers leverage the outputs and services of lower layers to perform their functions.

For instance, an evaluation at Layer 5 is designed to test for conformance with a policy defined at Layer 3, which is in turn informed by controls at Layer 2. This hierarchical dependency creates a clear logical flow of information and authority through the governance lifecycle.

This structure reverses the traditional document-centric view with policy building upon control. Since many organizations start with a policy which includes controls, the idea that policy would follow controls may seem counter-intuitive. However, this example is demonstrative of the model: A policy is not ready for implementation and evaluation until the controls are written.

## The Complete Model

The model is organized into two primary categories of activity These categories will, ideally, each be represented by documents or logs pertaining to the corresponding action.

Activities in the “definition” layers, one through three, should each produce document assets that may be referenced by higher layers or within their own layer. The “measurement” layers, five through seven, should each produce timestamped logs as outputs.

As noted in [_Foundational Concepts_](./03-foundational-concepts.md), the fourth layer defines sensitive activities connecting the two halves. The first three layers point toward the sensitive activities to define what is "good" and what is "bad". Meanwhile, the last three layers all look back at the sensitive activities — or their outcomes — to determine whether they comply with the definition of good (and in the case of Layer 7: Audit, to also determine the quality of results from all the lower layers). **Figure 4.1** shows these layers in cascading order.

The rest of this documentation describes each layers with examples.

{% include themed-image.html
  light="/assets/model-images/figure-4.1-light.png"
  dark="/assets/model-images/figure-4.1-dark.png"
  alt="Figure 4.1: Model Relationships and the Logical Flow"
  caption="Figure 4.1: Model Relationships and the Logical Flow"
  width="60%"
%}

---

## Continue Reading

- **< Previous Page**: - [Foundational Concepts](./03-foundational-concepts.html)
- **> Next Page**: [The Definition Layers](./05-definition-layers.html)

---
