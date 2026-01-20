---
layout: page
title: Creating Control Catalogs
---

## Learning Objectives

- Create Layer 2 Control Catalogs with capabilities, threats, and controls
- Map threats to capabilities
- Map controls to threats and Layer 1 Guidance

## Prerequisites

This tutorial builds on the Guidance document created in [Creating and Extending Guidance](01-guidance). You should have:
- `access-control-guidance.yaml` - The organizational access control Guidance

## Step 1: Create Layer 2 Control Catalog

Save the following as `cloud-access-controls.yaml`:

```yaml
title: Example Cloud Control Catalog
metadata:
  id: EXAMPLE-CLOUD
  description: Technology-specific controls for cloud-based access management
  author:
    id: author
    name: Author
    type: Human
  version: "v1.0.0"
  date: "2025-01-15"
  applicability-categories:
    - id: cloud-services
      title: Cloud Services
      description: Cloud-hosted applications and services
    - id: sso-providers
      title: SSO Providers
      description: Single sign-on and identity providers
  mapping-references:
    - id: ORG-GUIDANCE
      title: Organizational Access Control Guidance
      version: "v1.0.0"
      description: Layer 1 Guidance for access control

capabilities:
  - id: EXAMPLE-CLOUD-CAP-01
    title: User Authentication
    description: |
      Cloud services can authenticate users using username/password credentials
      or federated identity providers.
  - id: EXAMPLE-CLOUD-CAP-02
    title: Multi-Factor Authentication Configuration
    description: |
      Cloud services can configure and enforce multi-factor authentication
      using authenticator apps, hardware tokens, or SMS-based methods.

threats:
  - id: EXAMPLE-CLOUD-TH-01
    title: Credential Compromise
    description: |
      Compromised passwords or credentials can be used by attackers to gain
      unauthorized access to cloud services and sensitive data.
    capabilities:
      - reference-id: EXAMPLE-CLOUD
        entries:
          - reference-id: EXAMPLE-CLOUD-CAP-01
            remarks: Password-based authentication is vulnerable to compromise
  - id: EXAMPLE-CLOUD-TH-02
    title: Weak MFA Implementation
    description: |
      Cloud services with weak or misconfigured MFA allow attackers to bypass
      multi-factor requirements or use less secure authentication methods.
    capabilities:
      - reference-id: EXAMPLE-CLOUD
        entries:
          - reference-id: EXAMPLE-CLOUD-CAP-02
            remarks: MFA configuration can be disabled or use weak methods

families:
  - id: AC
    title: Access Control
    description: Controls for managing access to cloud systems

controls:
  - id: EXAMPLE-CLOUD-01
    family: AC
    title: Enforce Multi-Factor Authentication
    objective: |
      Ensure that all sensitive activities require two or more identity
      factors during authentication to prevent unauthorized access.
    threat-mappings:
      - reference-id: EXAMPLE-CLOUD
        entries:
          - reference-id: EXAMPLE-CLOUD-TH-01
            remarks: |
              Requiring MFA mitigates the threat of credential compromise by
              adding an additional authentication factor beyond passwords.
          - reference-id: EXAMPLE-CLOUD-TH-02
            remarks: |
              Enforcing strong MFA mitigates weak implementation threats.
    guideline-mappings:
      - reference-id: ORG-GUIDANCE
        entries:
          - reference-id: ORG-GUIDANCE-01
            remarks: |
              This control directly implements the organizational Guidance requirement 
              for multi-factor authentication in cloud environments.
    assessment-requirements:
      - id: EXAMPLE-CLOUD-01.01
        text: |
          When a cloud service authenticates users, MFA MUST be required for
          all access attempts using organization-approved authenticators.
        applicability:
          - cloud-services
          - sso-providers
        recommendation: |
          Configure cloud services to require MFA and disable SMS-based methods
          in favor of authenticator apps or hardware tokens.
      - id: EXAMPLE-CLOUD-01.02
        text: |
          When MFA is configured, SMS-based authentication methods MUST be
          disabled and only authenticator apps or hardware tokens permitted.
        applicability:
          - cloud-services
          - sso-providers
        recommendation: |
          Review MFA configuration settings and ensure SMS is disabled. Migrate
          users to authenticator apps or hardware tokens.
```

## Step 2: Verify

Validate Layer 2 Control Catalog:

```bash
cue vet schemas/layer-2.cue cloud-access-controls.yaml
```

## Outcome

Created Layer 2 Control Catalog with Capabilities describing cloud authentication features, Threats mapped to those Capabilities, and Controls that address Threats while mapping back to Layer 1 Guidance. This Control Catalog can now be referenced by Layer 3 Policies.

**Previous Tutorial:** [Creating and Extending Guidance](01-guidance.html) - Create Layer 1 Guidance documents.

**Next Tutorial:** [Creating Policy](03-policy.html) - Incorporate Guidance and Controls into a Policy document.
