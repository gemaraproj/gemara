---
layout: page
---

# Gemara Lexicon

The Gemara lexicon establishes stable definitions for compliance activities, describes their interactions, and provides standards for term usage.

## GRC Engineering Terms

| Entity | Definition | Context |
|--------|------------|---------|
| **GRC Engineering** | An approach that strategically applies engineering principles to GRC processes to make them more efficient and integrated. Also known as automated governance. |  |
| **Automated Governance** | An automated process for tracking governance throughout the deployment pipeline. Treating Governance as a required quality gate in the deployment to production (CI/CD). | |
| **GRC** | An integrated strategy for managing an organization's Governance, Risk, and Compliance to reliably achieve objectives, address uncertainty, and ensure integrity. |  |
| **Engineering** | The application of scientific principles to design, build, and maintain efficient, reliable, and scalable systems, structures, and processes. |  |
| **Governance** | The set of rules, policies, processes, and structures through which an organization is directed and controlled to achieve its objectives. |  |
| **Risk** | The potential for loss or damage when a threat is actualized. Typically informed by a threat assessment, calculated as the situational likelihood of a negative outcome, then modified by the resulting impact. | Layer 3 |
| **Compliance** | The act of adhering to stated requirements, which can include laws, regulations, industry standards, and internal corporate policies. |  |
| **Cybersecurity** | The processes and procedures implemented to reduce risk to a software system. |  |
| **Threat** | Any circumstance or event with the potential to adversely impact organizational operations or assets, often by exploiting a vulnerability but some threats are also inherent based on software capabilities. | Layer 2 |
| **Threat Actor** | Any entity or group that can manifest a threat, whether internal or external, human or machine, intentional or malicious. |  |
| **Threat Assessment** | The process and subsequent result of reviewing the possible threats and attack vectors for a particular service. | Layer 2 |
| **Attack Vector** | The specific path or method a threat actor uses to exploit a vulnerability. Examples include phishing emails, a compromised API endpoint, or an exposed server port. |  |
| **Vulnerability** | Any part of a system which creates opportunity for undesirable outcomes to be brought about by neglect, mistakes, or malicious action. |  |
| **Guidance** | High-level, abstract rules and frameworks pertaining to cybersecurity, typically developed by industry groups or government bodies. Because these are not tied to specific technologies, they are not likely to need frequent updates and may remain useful for years after creation. | Layer 1 |
| **Control Catalog** | A version-controlled document containing controls tailored to a specific technology. The catalog may also contain threats for better understanding the rationale behind the controls, and capabilities to increase precision around when and how the controls should be enforced. | Layer 2 |
| **Control Objective** | A technology-specific, threat-informed safeguard or countermeasure. | Layer 2 |
| **Control Family** | A logical grouping of controls which share a common purpose or function. Useful for quickly navigating complex control catalogs, and for ensuring proper coverage within a topical domain. | Layer 2 |
| **Mapping** | The process and subsequent result of identifying correlated entries across different documents, or internal to a document — such as mapping controls to threats or threats to capabilities in a control catalog. One entry may map to multiple others with varying strength levels, indicating how effective the source will satisfy the requirements of the target. |  |
| **Capability** | Highly specific descriptions of software behavior for elements such as command line interface, network accessibility, encryption by default, backups, recovery, and much more. | Layer 2 |
| **Assessment Requirement** | A specific, testable statement within a control that defines the exact conditions or evidence needed to verify its successful implementation. These requirements form the basis for evaluations, when deemed applicable according to the organization's policy. | Layer 2 |
| **Policy** | A formal statement or rule, informed by Guidance and Controls. It is tailored to an organization's specific risk appetite, and dictates the required or prohibited actions. | Layer 3 |
| **Risk Appetite** | The level of risk an organization is willing to accept in pursuit of its objectives. | Maps to Gemara Layer 3 |
| **Evaluation** | The inspection and assessment of code, configurations, and deployments to verify compliance with established policies. | Layer 5 |
| **Assessment** | The manual or automated process of evaluating control compliance following a specific assessment requirement. Multiple assessments may be required for a single control. | Layer 5 |
| **Evaluation Log** | A group of assessment logs corresponding to an evaluation plan. Provides the details necessary for enforcement actions. | Layer 5 |
| **Assessment Log** | The documented result of an assessment, containing details about when and how a specific set of steps was taken in accordance with an assessment requirement. | Layer 5 |
| **Enforcement** | The preventive or remedial actions taken based on evaluation findings, such as blocking a non-compliant deployment or automatically fixing a misconfiguration. | Layer 6 |
| **Enforcement Gate** | A manual or automated process which will prevent the deployment of any resource that cannot demonstrate compliance to achieve a satisfactory degree. | Layer 6 |
| **Remediation** | A manual or automated process which will fix any compliance issues that are identified by drift detection. | Layer 6 |
| **Drift Detection** | The process of evaluating resources after they have been deployed, with the aim of ensuring that there are no changes which might impact compliance or indicate a breach. |  |
| **Audit** | A formal, systematic review of an organization's policies, procedures, and conformance to ensure GRC processes are effective. | Layer 7 |
| **Inherent vs. Residual Risk** | Inherent Risk is the level of risk before any controls or mitigation efforts are applied. Residual Risk is the risk that remains after security controls have been implemented. |  |
| **Risk Treatment** | The process of selecting and implementing measures to modify risk. Common strategies are Mitigation (applying controls), Acceptance (formally acknowledging the risk), Transference (e.g., buying insurance), and Avoidance (ceasing the risky activity). |  |
| **Baseline** | A standardized level of minimum security configuration that is required for all systems of a certain type. A system must at least meet the baseline to be considered compliant. |  |
| **Compensating Control** | An alternative measure that can be used to satisfy a security requirement when the primary control cannot be implemented. It must provide a similar or greater level of defense. |  |
| **Policy as Code (PaC)** | The practice of managing security and compliance policies using a high-level, declarative programming language. This allows policies to be version-controlled, tested, and automatically enforced as part of the development lifecycle. |  |
| **Continuous ATO** | A modern approach to authorization where the Authority To Operate is maintained through continuous monitoring, automated assessments, and real-time risk data, rather than through static, point-in-time audits. |  |
| **OSCAL** | The Open Security Controls Assessment Language. A set of standardized, machine-readable formats (XML, JSON, YAML) for expressing and exchanging security control and assessment information, developed and governed by the United States' National Institute of Standards and Technology (NIST). |  |
| **NIST 800-53** | A publication by the National Institute of Standards and Technology that provides a comprehensive catalog of security and privacy controls for U.S. federal information systems. | Uses approaches from Gemara Layer 1 |
| **ISO 27001** | An international standard specifying the requirements for establishing, implementing, maintaining, and continually improving an Information Security Management System. | Uses approaches from Gemara Layer 1 |
| **NIST CSF** | The National Institute of Standards and Technology Cybersecurity Framework is a set of voluntary guidelines and best practices to help organizations manage cybersecurity risk. | Uses approaches from Gemara Layer 1 |
| **SSDF** | The Secure Software Development Framework (NIST SP 800-218) is a set of secure software development practices intended to be integrated into the software development lifecycle. | Uses approaches from Gemara Layer 1 |
| **FINOS CCC** | A FINOS project that creates a unified set of cybersecurity controls for the financial services industry by harmonizing global regulations and standards to simplify cloud adoption and compliance. | Adheres to Gemara Layer 2 |
| **OSPS Baseline** | A set of minimum security requirements for open-source projects, established by the OpenSSF (Open Source Security Foundation). It provides a clear baseline of practices to improve the security posture of open-source software. | Adheres to Gemara Layer 2 |
| **Privateer** | An open-source, read-only policy engine used to validate actual state and detect drift. It uses plugins to evaluate target resources—from cloud infrastructure to GitHub repositories—against policies and assessment requirements. | Adheres to Gemara Layer 5 |
