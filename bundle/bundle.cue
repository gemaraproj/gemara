// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")

// Package bundle describes how Gemara artifacts are stored as an OCI artifact.
// The rules that CUE cannot express, including the content check, are in
// README.md next to this file.
package bundle

import "list"

// ArtifactTypeV1 is the OCI artifactType of a version 1 bundle manifest.
#ArtifactTypeV1: "application/vnd.gemara.bundle.v1"

// ConfigMediaTypeV1 is the media type of the bundle config blob.
#ConfigMediaTypeV1: "application/vnd.gemara.manifest.v1+json"

// LayerMediaTypeV1 is the media type of a layer that holds one Gemara YAML document.
#LayerMediaTypeV1: "application/vnd.gemara.artifact.v1+yaml"

// ImageManifestMediaType is the OCI image manifest media type.
#ImageManifestMediaType: "application/vnd.oci.image.manifest.v1+json"

// RoleAnnotation names the layer annotation that carries a layer's role.
#RoleAnnotation: "org.gemara.artifact.role"

// TypeAnnotation names the optional layer annotation that carries the artifact type.
// Its value should be a Gemara #ArtifactType; consumers do not reject other values.
#TypeAnnotation: "org.gemara.artifact.type"

// TitleAnnotation names the layer annotation that carries the layer's file name.
#TitleAnnotation: "org.opencontainers.image.title"

// LicensesAnnotation names the optional manifest annotation that carries the publication licence.
#LicensesAnnotation: "org.opencontainers.image.licenses"

// Role is the closed set of layer roles in version 1.
// "artifact" is the document the bundle is about; "import" is a dependency of it.
#Role: "artifact" | "import"

// FileName is a layer's file name: one path segment, not "." or "..",
// and not "bundle.json" or "references", which the unpacked layout reserves.
#FileName: string & =~"^[^/\\\\]+$" & !="." & !=".." & !="bundle.json" & !="references"

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

// LayerDescriptor is a descriptor for a Gemara artifact layer.
#LayerDescriptor: #Descriptor & {
	mediaType!: #LayerMediaTypeV1
	annotations: {
		(#TitleAnnotation): #FileName
		(#RoleAnnotation):  #Role
		(#TypeAnnotation)?: string
		...
	}
}

// ForeignDescriptor is a layer of any other media type. Consumers ignore it.
#ForeignDescriptor: #Descriptor & {
	mediaType: !=#LayerMediaTypeV1
}

// Manifest is the OCI image manifest of a version 1 bundle.
#Manifest: {
	schemaVersion!: 2
	mediaType!:     #ImageManifestMediaType
	artifactType!:  #ArtifactTypeV1
	config!: #Descriptor & {mediaType!: #ConfigMediaTypeV1}
	layers: [_, ...] & [...(#LayerDescriptor | #ForeignDescriptor)]
	annotations?: {
		(#LicensesAnnotation)?: string
		[string]:               string
	}

	// Exactly one layer has role "artifact".
	// The guard defers the rule until the layers are known.
	if len(layers) > 0 {
		_oneArtifact: 1 & len([for l in layers if l.mediaType == #LayerMediaTypeV1 && l.annotations[#RoleAnnotation] == "artifact" {l}])
	}

	// Within a role, every layer has a different file name, so each layer has
	// its own path in the unpacked layout.
	for r in ["artifact", "import"] {
		_uniqueNames: (r): list.UniqueItems() & [for l in layers if l.mediaType == #LayerMediaTypeV1 && l.annotations[#RoleAnnotation] == r {l.annotations[#TitleAnnotation]}]
	}
	...
}

// A tool that unpacks a bundle may save it to disk as bundle.json, possibly reformatted. The content check hashes the config blob as fetched.
#Config: {
	// bundle-version is the bundled artifact's metadata.version, elevated from within the primary document in the bundle.
	"bundle-version": string
	// gemara-version is the Gemara specification version the bundled artifacts target.
	"gemara-version": string
	// revision is an optional producer-chosen revision label.
	revision?: string
	// metadata holds producer extensions, not to be confused with the Gemara document's metadata.
	metadata?: {...}
	// artifacts describes each bundled document and its dependencies.
	artifacts?: [...#ConfigArtifact]
	...
}

// ConfigArtifact describes one bundled document in the config.
#ConfigArtifact: {
	// name is the document's file name, matching its layer's title annotation.
	name: #FileName
	// type is the document's Gemara artifact type.
	type: string
	// id is the document's metadata.id.
	id: string
	// role is the document's layer role.
	role: #Role
	// dependencies lists the names of the documents this one references.
	dependencies?: [...#FileName]
}
