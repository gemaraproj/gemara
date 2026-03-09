---
layout: page
title: Foundational Concepts
---

## Prior Work

Much prior work exists to guide organizations in the adoption of Compliance Management Systems (CMS), such as *ISO 19600: Compliance management systems — Guidelines* and *ISO 37301:* *Compliance management systems — Requirements with Guidance for use*.

Other similar resources abound, such as the recent *Automated Governance Maturity Model* from the CNCF, which allows organizations to measure themselves and set growth goals for their GRC programs. In addition, much work has been done in the standardization space by the National Institute of Standards and Technology (NIST), a U.S. agency, on the [Open Security Controls Assessment Language](https://pages.nist.gov/OSCAL/) (OSCAL), which built upon the scope of automation protocols like [Security Content Automation Protocol](https://csrc.nist.gov/projects/security-content-automation-protocol) (SCAP).

Still, there is a deficiency in the consistent use of terminology, schemas, and processes — as well as the cross-team interoperability that could reasonably be expected when governing sensitive activities. These divisions have resulted in re-work both within organizations and across industries, with persistent gaps between Policy and reality.

**Section 8.1** discusses how organizations might learn from this prior work in comprehensive automated solutions.

## Assessing the State of Risk Assessments

While **Risk Assessment** may appear to be a straightforward topic at first glance, experience has shown that it takes many forms across a variety of scenarios. Some activities look forward, anticipating scenarios where Risk may occur. Other activities look backward, inspecting activities that are active or deployed to determine whether Risk has materialized and how to respond.

In some scenarios, Risk Assessment occurs even in the same lifecycle that a **Sensitive Activity** is being performed, though we might not directly refer to it as such. 

These Sensitive Activities may take many forms, from analog scenarios such as the daily operations of a bank teller to the seemingly antithetical software development lifecycle (SDLC). Still, the categorical activities performed to Control outcomes and mitigate Risk are similar.

As we’ll see in **Section 4**, Risk Assessments can be categorized into different types based on their inputs and outputs. These categories are cumulative, each calling upon other elements predictably, like layers building upon each other. 

At the beginning of the logical flow, Risk is hypothetical and cannot be made concrete without information about the specific organization and activity being performed. These activities are often performed by consortiums or as a matter of course during the creation of a Policy — sometimes causing the processes to be stunted and the outputs to be created with lower quality or insufficient completeness. Because those elements form the foundation, an incomplete understanding can reduce the effectiveness of entire GRC and security programs.

## Establishing a Layered Approach

The strategic value of adopting a layered architectural model for GRC lies in its ability to create a clear separation of concerns, breaking down a complex domain into manageable, interconnected components. 

The Gemara model's structure was directly inspired by *ISO 7498: Information Technology — Open Systems Interconnection — Basic Reference Model*. Better known as the OSI Model, it is an authoritative and proven framework that describes distinct layers of network functionality. 

By taking a similar approach to GRC, we enable different teams and tools to focus on specific activities, from high level Guidance to low level Enforcement, while ensuring they can interoperate within a cohesive system. Similar to the OSI Model, some tools may operate at multiple layers, and yet they still benefit from the clarity provided by describing their different activities according to the model.

Mirroring the layered approach, the Gemara model defines a contiguous seven-layer architecture. Within this structure, Layer 4 represents a pivot point: implementing Policies in a way that will later be evaluated. This layer captures the sensitive activities, such as software design or infrastructure provisioning that serve as a bridge where requirements and operational reality meet. 

---

## Continue Reading

- **< Previous Page**: [Definitions](./02-definitions)
- **> Next Page**: [The Model](./04-the-model)

---
