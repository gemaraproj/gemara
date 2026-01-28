---
layout: page
title: Metadata
---

## `Metadata`

Metadata represents common metadata fields shared across all layers

`author` **[Actor](base#actor)** _Required_

author is the person or group primarily responsible for this artifact

`description` **string** _Required_

description provides a high-level summary of the artifact's purpose and scope

`id` **string** _Required_

id allows this entry to be referenced by other elements

`applicability-categories` **array[[Category](base#category)]**

applicability-categories is a list of categories used to classify within this artifact to specify scope

`date` **[Date](base#date)**

date is the publication or effective date of this artifact

`draft` **boolean**

draft indicates whether this artifact is a pre-release version; open to modification

`lexicon` **string**

lexicon is a URI pointing to a controlled vocabulary or glossary relevant to this artifact

`mapping-references` **array[[MappingReference](mapping#mappingreference)]**

mapping-references is a list of external documents referenced within this artifact

`version` **string**

version is the version identifier of this artifact

