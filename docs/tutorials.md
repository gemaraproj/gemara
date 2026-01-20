---
layout: page
nav-title: Tutorials
title: Tutorials
---

## Overview

These tutorials guide you through using the Gemara CUE schemas to support an automated governance workflow, progressing from high-level Guidance to Evaluation.

Each tutorial builds upon the previous one, showing how compliance activities flow through the layers:

- **Layer 1 (Guidance)**: Start with high-level, technology-agnostic requirements
- **Layer 2 (Objectives)**: Translate Guidance into technology-specific Controls with associated Capabilities and Threats
- **Layer 3 (Policy)**: Create organizational Policies that incorporate Guidance and Controls, modify assessment requirements, and define Assessment Plans
- **Layer 4 (Evaluation)**: Build Evaluation tools that assess compliance of specific resources and generate Evaluation Logs (***Coming Soon***)

Tutorials can be followed independently, but completing them in sequence provides the most comprehensive understanding of how the layers interact.

1. **[Creating and Extending Guidance](tutorials/01-guidance)** - Create Layer 1 Guidance documents that link to existing standards with organizational requirements.

2. **[Creating Control Catalogs](tutorials/02-controls)** - Build Layer 2 Control Catalogs by mapping Capabilities, Threats, and Controls.

3. **[Creating Policy](tutorials/03-policy)** - Define Layer 3 Policies that import Guidance and Controls, modify assessment requirements, and specify Assessment Plans.

## Prerequisites

Before starting any tutorial, ensure you have:

- **Familiarity with YAML syntax** - All Gemara documents use YAML format
- **Understanding of Gemara layers** - Basic knowledge of the six-layer model (see the [Gemara Model documentation](model))
- **Text editor** - Any YAML-capable editor
- **CUE installed** - For schema validation (optional but recommended)

