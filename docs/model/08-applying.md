---
layout: page
title: The Need for Machine-Optimized Documentation Standards
---

## Going Beyond Machine-Readable

As mentioned in the _[Prior Work](./03-foundational-concepts)_, the *GRC Engineering Model for Automated Risk Assessments* is the product of many previous iterations and learnings. 

Immense value has come from the work on OSCAL, where a large body of machine-readable documents have been crafted both publicly and privately by highly regulated organizations. The work done in the OSCAL space cannot be overlooked or discounted, as it has provided a great benefit to countless firms.

However, there is something to be said about the benefits of *optimizing* for machines when crafting machine-readable GRC documents. Previous efforts to craft relevant schemas have fallen short in two key areas.

The first, and most obvious, shortcoming comes from tools that misappropriate the term “policy” to describe security tooling configurations. As useful as these resources may be, far too many highly-trained GRC and security professionals have been stuck in jobs where they are writing *tooling configurations* under the pretense of writing policy-as-code.

The second is more difficult to spot, and it comes in the form of a specification which allows for customized fields. As innocent — or even necessary — as it may sound, the addition of custom fields in GRC documents results in the need for customized tools that have foreknowledge of those fields. The tradeoff for this flexibility is a sacrifice of the structural clarity required to build widely adopted tooling around the data. While a schema should be flexible enough to support a variety of uses, it should also be opinionated enough to allow an automation ecosystem to rise up and thrive around it.

By drawing on the positive lessons and addressing known issues, we can enable organizations to engineer highly efficient GRC ecosystems that are robust enough to address complex problems, but comprehensible enough for any professional to quickly get up to speed — all while maintaining the ability to export structured data to other formats, such as OSCAL.

## Improvements for Humans; Improvements for AI

An organization that adopts an optimized GRC Engineering strategy can gain benefits for both humans and artificial intelligence.

Communication and handoffs are a well-known obstacle for highly regulated firms. Field observations consistently reveal challenges distributing policies to impacted parties, managing changes, accidental rework, and even months spent on unnecessary work due to misunderstandings.

Firms that standardize document schemas can develop the necessary tools to get documents to the right place at the right time. The database and APIs necessary to build these communication optimizations provide a secondary benefit: they act as the foundation for Model Context Protocol (MCP).

When a properly designed MCP system sits on top of a machine-optimized GRC database, AI capabilities compound with multiplicative benefits. In a trial reported by Morgan Stanley in collaboration with the Financial Technology Open Source Foundation, an MCP server backed by properly structured data was able to produce 90% deterministic results from an AI agent.

When partnered with an Agent2Agent (A2A) enabled system — allowing GRC-specialized systems to provide precisely contextualized information to engineering-specialized systems in a near-deterministic manner — our AI systems will require dramatically fewer resources, our security tools increase in accuracy, less time is spent on false positives or negatives, and more time is spent delivering real value to the organization.

- **< Previous Page**: [Layer 7](./07.3-Layer-7)
<!-- - **> Next Page**: [Gemara SDKs](TODO)-->
- **> Next Page**: [Authors & Acknowledgments](./10-acknowledgement)
