#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# Outputs a display title for a schema file base name (e.g. controlcatalog -> "Control Catalog").
case "$1" in
base) echo "Base" ;;
metadata) echo "Metadata" ;;
mapping-inline) echo "Mapping Primitives" ;;
guidancecatalog) echo "Guidance Catalog" ;;
vectorcatalog) echo "Vector Catalog" ;;
controlcatalog) echo "Control Catalog" ;;
threatcatalog) echo "Threat Catalog" ;;
riskcatalog) echo "Risk Catalog" ;;
policy) echo "Policy" ;;
mappingdocument) echo "Mapping Document" ;;
evaluationlog) echo "Evaluation Log" ;;
enforcementlog) echo "Enforcement Log" ;;
*) echo "$1" ;;
esac
