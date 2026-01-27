#!/bin/sh
# Parses docs/schema-nav.yml and outputs page information
# Usage: parse-nav.sh <nav-file> <command>
# Commands: list-pages, get-title <filename>, get-filename <title>

NAV_FILE="$1"
COMMAND="$2"
ARG="$3"

case "$COMMAND" in
  list-pages)
    # Output: filename|title (one per line)
    awk '/^  - title:/ { 
      title = $0; 
      gsub(/.*title: "/, "", title); 
      gsub(/".*/, "", title); 
      getline; 
      if ($0 ~ /filename:/) { 
        filename = $0; 
        gsub(/.*filename: /, "", filename); 
        gsub(/[ "]/, "", filename); 
      } else { 
        filename = ""; 
      } 
      if (filename == "") { 
        filename = tolower(title); 
        gsub(/ /, "-", filename); 
      } 
      print filename "|" title 
    }' "$NAV_FILE"
    ;;
  get-title)
    # Get title for a given filename
    awk -v filename="$ARG" '
    BEGIN { found = 0 }
    /^  - title:/ { 
      title = $0; 
      gsub(/.*title: "/, "", title); 
      gsub(/".*/, "", title); 
      getline; 
      if ($0 ~ /filename:/) { 
        file = $0; 
        gsub(/.*filename: /, "", file); 
        gsub(/[ "]/, "", file); 
      } else { 
        file = ""; 
      } 
      if (file == "") { 
        file = tolower(title); 
        gsub(/ /, "-", file); 
      } 
      if (file == filename) { 
        print title; 
        found = 1; 
        exit 
      } 
    }
    END { if (!found) exit 1 }
    ' "$NAV_FILE"
    ;;
  *)
    echo "Unknown command: $COMMAND" >&2
    exit 1
    ;;
esac
