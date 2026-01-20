---
layout: page
title: Creating and Extending Guidance
---

## Learning Objectives

- Create Guidance documents that link to existing standards
- Extend existing Guidance with organizational requirements
- Validate Guidance documents

## Step 1: Create Guidance Linked to Existing Standards

Save the following as `access-control-guidance.yaml`:

```yaml
title: Organizational Access Control Guidance
metadata:
  id: ORG-GUIDANCE
  description: Organizational guidance for access control requirements that links to NIST 800-53
  author:
    id: org-123
    name: Organization 123
    type: Human
  version: "v1.0.0"
  date: "2025-01-15"
document-type: Framework
  applicability-categories:
    - id: production
      title: Production Systems
      description: Applies to all production systems handling sensitive data
    - id: stage
      title: Staging Systems
      description: Applies to staging and pre-production environments
  mapping-references:
    - id: NIST-800-53-Rev5
      title: NIST Special Publication 800-53 Revision 5
      version: "Rev 5"
      description: Complete NIST 800-53 Catalog
      url: "https://csrc.nist.gov/publications/detail/sp/800-53/rev-5/final"

families:
  - id: Access-Control
    title: Access Control
    description: Controls for managing access to information systems

guidelines:
  - id: ORG-GUIDANCE-01
    title: Require Multi-Factor Authentication
    family: Access-Control
    objective: |
      Ensure that all access to sensitive systems requires multi-factor 
      authentication to prevent unauthorized access through compromised credentials.
    statements:
      - id: ORG-GUIDANCE-01.01
        title: MFA Requirement
        text: |
          All access to systems handling sensitive organizational data MUST require 
          multi-factor authentication using organization-approved authenticators.
    guideline-mappings:
      - reference-id: NIST-800-53-Rev5
        entries:
          - reference-id: IA-2
            strength: 9
            remarks: Identification and Authentication (Organizational Users)
          - reference-id: IA-2(1)
            strength: 9
            remarks: Network Access to Privileged Accounts
          - reference-id: IA-2(2)
            strength: 9
            remarks: Network Access to Non-Privileged Accounts
    applicability:
      - production
      - stage
```

## Step 2: Extend Existing Guidance

Add the following guideline to `access-control-guidance.yaml` to extend the organizational MFA requirement (ORG-GUIDANCE-01) with additional session management requirements that apply only to production systems:

```yaml
  - id: ORG-GUIDANCE-01-SESSION
    title: MFA Session Management Extension
    family: Access-Control
    extends:
      entry-id: ORG-GUIDANCE-01
    objective: |
      Extend the organizational MFA requirement with stricter session management and timeout 
      policies for production systems to reduce the risk of unauthorized access from abandoned 
      or compromised sessions.
    statements:
      - id: ORG-GUIDANCE-01-SESSION.01
        title: Inactivity Timeout Requirement
        text: |
          All interactive sessions in production systems protected by MFA MUST automatically 
          timeout after 15 minutes of inactivity and require re-authentication using MFA.
      - id: ORG-GUIDANCE-01-SESSION.02
        title: Maximum Session Duration
        text: |
          All interactive sessions in production systems protected by MFA MUST be terminated 
          after a maximum duration of 4 hours, regardless of activity, and require re-authentication 
          using MFA to establish a new session.
      - id: ORG-GUIDANCE-01-SESSION.03
        title: Re-authentication for Sensitive Operations
        text: |
          Production systems MUST require re-authentication using MFA before performing 
          sensitive operations such as modifying user permissions, changing system configuration, 
          or accessing highly sensitive data, even if the session is still active.
    applicability:
      - production
```

## Step 3: Verify

Validate the Guidance document:

```bash
cue vet schemas/layer-1.cue access-control-guidance.yaml
```

## Outcome

Created technology-agnostic Guidance that links to NIST 800-53 MFA controls. The base MFA requirement applies to both production and staging systems, while the session management extension adds stricter timeout and re-authentication requirements that apply only to production systems. This Guidance document demonstrates both linking to existing standards and extending your own organizational guidance with environment-specific requirements. It can now be referenced by Layer 2 Control Catalogs and Layer 3 Policies.

**Next Tutorial:** [Creating Control Catalogs](02-controls.html) - Break down Guidance into technology-specific Control Catalogs.
