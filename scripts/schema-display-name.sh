#!/bin/sh
# Outputs a display title for a schema file base name (e.g. layer-1 -> "Layer 1").
case "$1" in
base) echo "Base" ;;
metadata) echo "Metadata" ;;
mapping) echo "Mapping" ;;
layer-1) echo "Layer 1" ;;
layer-2) echo "Layer 2" ;;
layer-3) echo "Layer 3" ;;
layer-5) echo "Layer 5" ;;
*) echo "$1" ;;
esac
