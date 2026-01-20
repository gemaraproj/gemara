---
layout: page
title: Creating Policy
---

## Learning Objectives

- Create Layer 3 Policies that consume Layer 1 Guidance and Layer 2 Controls
- Modify assessment requirements to specify organizational requirements
- Define Assessment Plans for Evaluation

## Prerequisites

This tutorial builds on previous tutorials. You should have:
- `access-control-guidance.yaml` - From [Creating and Extending Guidance](01-guidance.html)
- `cloud-access-controls.yaml` - From [Creating Control Catalogs](02-controls.html)

## Step 1: Create Policy Document

Save the following as `cloud-access-policy.yaml`:

```yaml
title: Cloud Access Security Policy
metadata:
  id: CLOUD-ACCESS-POLICY-001
  description: Ensure secure access to cloud services through multi-factor authentication
  version: "v1.0.0"
  author:
    id: org-123
    name: Organization 123
    type: Human
    contact:
      name: Author
      affiliation: IT Security
      email: author@example.com
  date: "2025-01-20"
  mapping-references:
    - id: ORG-GUIDANCE
      title: Organizational Access Control Guidance
      version: "v1.0.0"
      description: Layer 1 Guidance for access control
    - id: EXAMPLE-CLOUD
      title: Example Cloud Control Catalog
      version: "v1.0.0"
      description: Layer 2 controls for cloud access management

scope:
  in:
    technologies:
      - "Cloud Services"
      - "SSO Providers"
      - "Identity Management Systems"
  out:
    technologies:
      - "On-Premises Systems"

contacts:
  responsible:
    - name: "Identity and Access Management Team"
      primary: true
      affiliation: "IT Security"
      email: "iam-team@example.com"
  accountable:
    - name: "Author"
      primary: true
      affiliation: "IT Security"
      email: "author@example.com"
  consulted:
    - name: "Security Team"
      affiliation: "IT Security"
      email: "security@example.com"

imports:
  guidance:
    - reference-id: ORG-GUIDANCE
      constraints:
        - id: CONSTRAINT-ORG-GUIDANCE-01
          target-id: ORG-GUIDANCE-01
          text: |
            All cloud service access MUST require multi-factor authentication.
  catalogs:
    - reference-id: EXAMPLE-CLOUD
      assessment-requirement-modifications:
        - id: MOD-EXAMPLE-CLOUD-01-02
          target-id: EXAMPLE-CLOUD-01.02
          modification-type: "modify"
          modification-rationale: |
            This policy requires stricter MFA requirements for administrative access. While the 
            base control allows authenticator apps or hardware tokens, this policy restricts 
            admin access to organization-issued hardware tokens only to ensure physical control 
            over authentication devices for privileged accounts.
          text: |
            When MFA is configured, SMS-based authentication methods MUST be disabled.
            For administrative access to cloud services and SSO providers, only organization-issued 
            hardware tokens (FIDO2/YubiKey) are permitted. Authenticator apps may be used for 
            non-administrative access.
          applicability: ["cloud-services", "sso-providers"]
          recommendation: |
            Configure cloud services to disable SMS methods. For administrative accounts, ensure 
            only organization-issued hardware tokens are enabled for MFA. Non-administrative 
            accounts may use authenticator apps or hardware tokens.

adherence:
  assessment-plans:
    - id: AP-EXAMPLE-CLOUD-01-01
      requirement-id: EXAMPLE-CLOUD-01.01
      frequency: "continuous"
      evaluation-methods:
        - type: automated
          description: Automated MFA requirement checking
          actor:
            id: mfa-checker
            name: MFA Checker Service
            type: Software
      evidence-requirements: |
        MFA configuration status must be logged with timestamps for audit purposes.
    - id: AP-EXAMPLE-CLOUD-01-02
      requirement-id: EXAMPLE-CLOUD-01.02
      frequency: "daily"
      evaluation-methods:
        - type: automated
          description: Automated MFA method validation
          actor:
            id: mfa-validator
            name: MFA Validator Service
            type: Software
      evidence-requirements: |
        MFA configuration status and authentication method usage must be logged 
        with timestamps for audit purposes.
```

## Step 2: Verify

Validate the policy document:

```bash
cue vet schemas/layer-3.cue cloud-access-policy.yaml
```

## Outcome

Created a Layer 3 Policy that imports Layer 1 Guidance and Layer 2 Controls, modifies assessment requirements to specify organizational requirements, and defines Assessment Plans that configure how Evaluations are performed. These Assessment Plans can now be referenced by Layer 4 Evaluation tools.

**Previous Tutorial:** [Creating Control Catalogs](02-controls.html) - Create Layer 2 Control Catalogs.
