// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")

// Package bundle describes how Gemara documents are stored as an OCI artifact.
// Each document is one entry in the manifest's OCI `layers` array; OCI layers are
// unrelated to the layers of the Gemara model.
// The rules that CUE cannot express, including the content check, are in
// README.md next to this file.
package bundle

import "list"

// ArtifactTypeV1 is the OCI artifactType of a version 1 bundle manifest.
#ArtifactTypeV1: "application/vnd.gemara.bundle.v1"

// ConfigMediaTypeV1 is the media type of the bundle config blob.
#ConfigMediaTypeV1: "application/vnd.gemara.manifest.v1+json"

// DocumentMediaTypeV1 is the media type of a layers entry that holds one Gemara YAML document.
#DocumentMediaTypeV1: "application/vnd.gemara.artifact.v1+yaml"

// ImageManifestMediaType is the OCI image manifest media type.
#ImageManifestMediaType: "application/vnd.oci.image.manifest.v1+json"

// RoleAnnotation names the document annotation that carries the document's role.
#RoleAnnotation: "org.gemara.artifact.role"

// TypeAnnotation names the optional document annotation that carries the document's artifact type.
// Its value should be a Gemara #ArtifactType; consumers do not reject other values.
#TypeAnnotation: "org.gemara.artifact.type"

// TitleAnnotation names the document annotation that carries the document's file name.
#TitleAnnotation: "org.opencontainers.image.title"

// LicensesAnnotation names the optional manifest annotation that carries the publication licence.
#LicensesAnnotation: "org.opencontainers.image.licenses"

// Role is the closed set of document roles in version 1.
// "artifact" marks the primary document, the one the bundle is about; "import" marks a document it depends on.
#Role: "artifact" | "import"

// FileName is a document's file name: one path segment, not "." or "..",
// and not "bundle.json", "imports" or "references", which the unpacked layout reserves.
#FileName: string & =~"^[^/\\\\]+$" & !="." & !=".." & !="bundle.json" & !="imports" & !="references"

// Digest is an OCI content digest, for example "sha256:<64 hex>".
#Digest: string & =~"^[a-z0-9]+(?:[+._-][a-z0-9]+)*:[a-zA-Z0-9=_-]+$"

// Descriptor is an OCI content descriptor. Fields this format does not use are allowed.
#Descriptor: {
	mediaType: string
	digest:    #Digest
	size:      int & >=0
	annotations?: [string]: string
	...
}

// DocumentDescriptor is the layers entry for one Gemara document.
#DocumentDescriptor: #Descriptor & {
	mediaType!: #DocumentMediaTypeV1
	annotations: {
		(#TitleAnnotation): #FileName
		(#RoleAnnotation):  #Role
		(#TypeAnnotation)?: string
		...
	}
}

// ForeignDescriptor is a layers entry of a non-Gemara media type. Consumers ignore it.
// A Gemara document of another version is not foreign: it makes the manifest malformed.
#ForeignDescriptor: #Descriptor & {
	mediaType: !~"^application/vnd\\.gemara\\.artifact\\."
}

// Manifest is the OCI image manifest of a version 1 bundle.
#Manifest: {
	schemaVersion!: 2
	mediaType!:     #ImageManifestMediaType
	artifactType!:  #ArtifactTypeV1
	config!: #Descriptor & {mediaType!: #ConfigMediaTypeV1}
	layers: [_, ...] & [...(#DocumentDescriptor | #ForeignDescriptor)]
	annotations?: {
		(#LicensesAnnotation)?: string
		[string]:               string
	}

	// Exactly one document has role "artifact" (the primary document).
	// The guard defers the rule until `layers` is known.
	if len(layers) > 0 {
		_onePrimary: 1 & len([for l in layers if l.mediaType == #DocumentMediaTypeV1 && l.annotations[#RoleAnnotation] == "artifact" {l}])
	}

	// Within a role, every document has a different file name, so each document has
	// its own path in the unpacked layout.
	for r in ["artifact", "import"] {
		_uniqueNames: (r): list.UniqueItems() & [for l in layers if l.mediaType == #DocumentMediaTypeV1 && l.annotations[#RoleAnnotation] == r {l.annotations[#TitleAnnotation]}]
	}
	...
}

// A tool that unpacks a bundle may save it to disk as bundle.json, possibly reformatted. The content check hashes the config blob as fetched.
#Config: {
	// bundle-version is the primary document's metadata.version.
	"bundle-version": string
	// gemara-version is the Gemara specification version the bundled documents target.
	"gemara-version": string
	// revision is an optional producer-chosen revision label.
	revision?: string
	// artifacts describes each bundled document and its dependencies.
	artifacts?: [...#ConfigArtifact]
	...
}

// ConfigArtifact describes one bundled document in the config.
#ConfigArtifact: {
	// name is the document's file name, matching its title annotation.
	name: #FileName
	// type is the document's Gemara artifact type.
	type: string
	// id is the document's metadata.id.
	id: string
	// role is the document's role annotation.
	role: #Role
	// dependencies lists the names of the documents this one references.
	dependencies?: [...#FileName]
}
