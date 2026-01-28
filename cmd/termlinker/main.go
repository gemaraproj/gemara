package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/goccy/go-yaml"
)

type Term struct {
	Term       string   `yaml:"term"`
	Definition string   `yaml:"definition"`
	References []string `yaml:"references"`
}

type TermInfo struct {
	OriginalTerm string
	LowerTerm    string
	Slug         string
	Regex        *regexp.Regexp
}

func main() {
	lexiconFile := flag.String("lexicon", "docs/lexicon.yaml", "Input lexicon YAML file")
	docsDir := flag.String("docs", "docs", "Documentation directory to process")
	cleanup := flag.Bool("cleanup", false, "Remove termlinker-generated links instead of adding them")
	flag.Parse()

	// Load terms from lexicon
	terms, err := loadTerms(*lexiconFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading terms: %v\n", err)
		os.Exit(1)
	}

	// Build term info with regex patterns (sorted by length, longest first)
	termInfos := buildTermInfos(terms)

	// Find all markdown files
	mdFiles, err := findMarkdownFiles(*docsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding markdown files: %v\n", err)
		os.Exit(1)
	}

	// Process each file
	processedCount := 0
	for _, file := range mdFiles {
		// Skip the definitions page itself
		if strings.HasSuffix(file, "model/02-definitions.md") {
			continue
		}

		if *cleanup {
			if err := cleanupFile(file, termInfos, *docsDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error cleaning up %s: %v\n", file, err)
				continue
			}
		} else {
			if err := processFile(file, termInfos, *docsDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", file, err)
				continue
			}
		}
		processedCount++
	}

	if *cleanup {
		fmt.Printf("Successfully cleaned up %d markdown files\n", processedCount)
	} else {
		fmt.Printf("Successfully processed %d markdown files\n", processedCount)
	}
}

func loadTerms(lexiconFile string) ([]Term, error) {
	data, err := os.ReadFile(lexiconFile)
	if err != nil {
		return nil, fmt.Errorf("read lexicon file: %w", err)
	}

	var terms []Term
	if err := yaml.Unmarshal(data, &terms); err != nil {
		return nil, fmt.Errorf("parse lexicon YAML: %w", err)
	}

	return terms, nil
}

func buildTermInfos(terms []Term) []TermInfo {
	termInfos := make([]TermInfo, 0, len(terms))

	for _, term := range terms {
		lowerTerm := strings.ToLower(term.Term)
		slug := termToSlug(term.Term)

		// Create regex for whole-word, case-insensitive matching
		// Escape special regex characters in the term
		escapedTerm := regexp.QuoteMeta(term.Term)
		// Use word boundaries for whole-word matching
		pattern := `(?i)\b` + escapedTerm + `\b`
		regex, err := regexp.Compile(pattern)
		if err != nil {
			// Skip terms that can't be compiled (shouldn't happen)
			continue
		}

		termInfos = append(termInfos, TermInfo{
			OriginalTerm: term.Term,
			LowerTerm:    lowerTerm,
			Slug:         slug,
			Regex:        regex,
		})
	}

	// Sort by length (longest first) to avoid partial matches
	sort.Slice(termInfos, func(i, j int) bool {
		return len(termInfos[i].OriginalTerm) > len(termInfos[j].OriginalTerm)
	})

	return termInfos
}

func termToSlug(term string) string {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(term)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove any other non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range slug {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func findMarkdownFiles(docsDir string) ([]string, error) {
	var files []string
	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func processFile(filePath string, termInfos []TermInfo, docsDir string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// Calculate relative path to definitions page
	relPath := calculateRelativePath(filePath, docsDir)

	// Process the content
	processed := processContent(string(content), termInfos, relPath)

	// Write back
	if err := os.WriteFile(filePath, []byte(processed), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func cleanupFile(filePath string, termInfos []TermInfo, docsDir string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// Calculate relative path to definitions page (for matching)
	relPath := calculateRelativePath(filePath, docsDir)

	// Process the content to remove links
	processed := cleanupContent(string(content), termInfos, relPath)

	// Write back
	if err := os.WriteFile(filePath, []byte(processed), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func calculateRelativePath(filePath, docsDir string) string {
	// Get the directory of the current file
	fileDir := filepath.Dir(filePath)

	// Calculate relative path from file directory to definitions page
	defPagePath := filepath.Join(docsDir, "model/02-definitions.html")
	relPath, err := filepath.Rel(fileDir, defPagePath)
	if err != nil {
		// Fallback to absolute path
		return "/model/02-definitions.html"
	}

	// Normalize path separators for URLs
	return filepath.ToSlash(relPath)
}

func processContent(content string, termInfos []TermInfo, defPath string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder

	inCodeBlock := false
	inFrontMatter := false
	frontMatterCount := 0

	for i, line := range lines {
		originalLine := line

		// Track front matter
		if strings.HasPrefix(strings.TrimSpace(line), "---") {
			if !inFrontMatter {
				inFrontMatter = true
				frontMatterCount = 1
			} else {
				frontMatterCount++
				if frontMatterCount == 2 {
					inFrontMatter = false
				}
			}
		}

		// Skip processing in front matter
		if inFrontMatter {
			result.WriteString(originalLine)
			if i < len(lines)-1 {
				result.WriteString("\n")
			}
			continue
		}

		// Track code blocks (fenced with ```)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
		}

		// Skip code blocks
		if inCodeBlock {
			result.WriteString(originalLine)
			if i < len(lines)-1 {
				result.WriteString("\n")
			}
			continue
		}

		// Skip headers (lines starting with #)
		if strings.HasPrefix(trimmed, "#") {
			result.WriteString(originalLine)
			if i < len(lines)-1 {
				result.WriteString("\n")
			}
			continue
		}

		// Process the line
		processedLine := processLine(line, termInfos, defPath)
		result.WriteString(processedLine)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

func processLine(line string, termInfos []TermInfo, defPath string) string {
	// Find all existing markdown links, inline code, and HTML tags to skip
	linkPattern := regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
	inlineCodePattern := regexp.MustCompile("`[^`]+`")
	// Match HTML tags: <tag>, </tag>, <tag/>, <tag attr="value">, etc.
	// This regex matches opening tags, closing tags, and self-closing tags
	htmlTagPattern := regexp.MustCompile(`<[^>]+>`)

	// First, clean up any existing nested links (e.g., [Term [Subterm](...)](...))
	// Replace them with just the outer link, using the proper slug for the compound term
	// This pattern matches nested links including malformed hash fragments
	// Pattern 1: [[Word](url#word) RestOfTerm](url#[word](url#word)-restofterm)
	splitTermPattern := regexp.MustCompile(`\[\[([^\]]+)\]\(([^\)]+)\)\s+([^\]]+)\]\(([^\)]+)\)`)
	result := splitTermPattern.ReplaceAllStringFunc(line, func(match string) string {
		submatches := splitTermPattern.FindStringSubmatch(match)
		if len(submatches) >= 5 {
			firstWord := strings.TrimSpace(submatches[1])
			_ = submatches[2] // First word URL (unused)
			restOfTerm := strings.TrimSpace(submatches[3])
			_ = submatches[4] // Outer URL with malformed hash (unused)
			
			// Combine to form the full compound term
			fullTerm := firstWord + " " + restOfTerm
			
			// Check if this matches a known compound term
			for _, termInfo := range termInfos {
				if strings.EqualFold(fullTerm, termInfo.OriginalTerm) {
					// This is a known compound term - use its proper slug
					return fmt.Sprintf("[%s](%s#%s)", fullTerm, defPath, termInfo.Slug)
				}
			}
			// If not a known term, create a clean slug
			slug := termToSlug(fullTerm)
			return fmt.Sprintf("[%s](%s#%s)", fullTerm, defPath, slug)
		}
		return match
	})
	
	// Pattern 2: [Term [Subterm](...)](...) - general nested link pattern
	// This pattern needs to handle nested parentheses in the hash fragment
	nestedLinkPattern := regexp.MustCompile(`\[([^\[]*)\[([^\]]+)\]\(([^\)]+)\)\]\(([^\)]+)\)`)
	result = nestedLinkPattern.ReplaceAllStringFunc(result, func(match string) string {
		// Extract the outer term and inner term from the nested link
		submatches := nestedLinkPattern.FindStringSubmatch(match)
		if len(submatches) >= 5 {
			outerTerm := strings.TrimSpace(submatches[1] + submatches[2]) // Combine prefix and inner term
			_ = submatches[3]                                             // The inner link URL (unused)
			_ = submatches[4]                                             // The outer link URL (may contain malformed hash)
			
			// Check if this matches a known compound term
			for _, termInfo := range termInfos {
				if strings.EqualFold(outerTerm, termInfo.OriginalTerm) {
					// This is a known compound term - use its proper slug
					return fmt.Sprintf("[%s](%s#%s)", outerTerm, defPath, termInfo.Slug)
				}
			}
			// If not a known term, try to extract a clean slug from the outer link
			// Remove the nested link structure and create a clean link
			slug := termToSlug(outerTerm)
			return fmt.Sprintf("[%s](%s#%s)", outerTerm, defPath, slug)
		}
		return match // Return unchanged if we can't parse it
	})
	
	// Pattern 3: Clean up extremely long hash fragments that contain way too much text
	// This handles cases where the hash includes entire definitions or sentences
	// Pattern: ](url#extremely-long-hash-that-contains-too-much-text) or ](url#extremely-long-hash-that-contains-too-much-text))
	longHashPattern := regexp.MustCompile(`(\]\([^\)]+#)([a-z0-9-]{50,})(\))\)?`)
	result = longHashPattern.ReplaceAllStringFunc(result, func(match string) string {
		submatches := longHashPattern.FindStringSubmatch(match)
		if len(submatches) >= 4 {
			linkStart := submatches[1] // ](url#
			longHash := submatches[2]   // The extremely long hash
			_ = submatches[3]           // ) or )) (unused - we always return single )
			
			// Try to extract a reasonable slug from the long hash
			// Look for common compound term patterns at the end
			for _, termInfo := range termInfos {
				termSlug := termInfo.Slug
				// Check if the hash ends with this term's slug
				if strings.HasSuffix(longHash, termSlug) {
					// Extract the part before the term slug to see if it's a prefix
					prefix := strings.TrimSuffix(longHash, termSlug)
					// If the prefix ends with the first word of the term, use the full term slug
					firstWord := strings.ToLower(strings.Fields(termInfo.OriginalTerm)[0])
					if strings.HasSuffix(prefix, firstWord) || prefix == "" {
						// Return with just one closing parenthesis (remove any extra)
						return linkStart + termSlug + ")"
					}
				}
			}
			// If we can't find a match, try to extract just the last reasonable part
			// Split by common delimiters and take the last meaningful segment
			parts := strings.Split(longHash, "-")
			if len(parts) >= 2 {
				// Take the last 2-3 parts as a potential compound term slug
				potentialSlug := strings.Join(parts[len(parts)-2:], "-")
				// Check if this matches a known term
				for _, termInfo := range termInfos {
					if termInfo.Slug == potentialSlug {
						// Return with just one closing parenthesis (remove any extra)
						return linkStart + potentialSlug + ")"
					}
				}
			}
		}
		return match
	})
	
	// Clean up trailing patterns left from malformed hash fragments
	// Pattern: ](#term-slug)-last-part) should become just ](#term-slug)
	// This handles cases like [Control Catalog](#control-catalog)-catalog) -> [Control Catalog](#control-catalog)
	// The trailing part matches the last segment of the slug (after the last hyphen)
	trailingSlugPattern := regexp.MustCompile(`(\]\([^\)]+#([a-z0-9-]+)\))-([a-z0-9-]+)\)`)
	result = trailingSlugPattern.ReplaceAllStringFunc(result, func(match string) string {
		submatches := trailingSlugPattern.FindStringSubmatch(match)
		if len(submatches) >= 4 {
			fullSlug := submatches[2]
			trailingPart := submatches[3]
			// Extract the last segment of the slug (after the last hyphen)
			lastHyphenIndex := strings.LastIndex(fullSlug, "-")
			if lastHyphenIndex >= 0 {
				lastSegment := fullSlug[lastHyphenIndex+1:]
				// If the trailing part matches the last segment, remove it
				if lastSegment == trailingPart {
					return submatches[1] + ")"
				}
			} else if fullSlug == trailingPart {
				// If there's no hyphen and they match exactly, remove it
				return submatches[1] + ")"
			}
		}
		return match
	})
	
	// Clean up any extra closing parentheses that might be left after fixing nested links
	// This handles cases where the original nested link had malformed hash fragments
	// Pattern: markdown link followed by extra ) and then **, whitespace, or end of string
	extraParenPattern := regexp.MustCompile(`(\[([^\]]+)\]\([^\)]+\))\)(\s*\*\*|\s|[,.;:!?)]|$)`)
	result = extraParenPattern.ReplaceAllString(result, `$1$3`)
	
	// Also clean up cases where there are double closing parentheses: ](...)))
	// This handles cases like [term](#slug)) or *[term](#slug))*
	// Match ](url)) and replace with ](url) - be careful to only match link endings
	// Pattern: ]( followed by non-) chars, then )), but make sure it's a link ending
	linkDoubleParenPattern := regexp.MustCompile(`(\]\([^\)]+\))\)\)([^*\w]|$)`)
	result = linkDoubleParenPattern.ReplaceAllString(result, `$1)$2`)
	
	// Clean up any double opening brackets that might have been created
	// Pattern: **[[text](...)]** should be **[text](...)]**
	doubleBracketPattern := regexp.MustCompile(`\*\*\[\[([^\]]+)\]\(([^\)]+)\)\]\*\*`)
	result = doubleBracketPattern.ReplaceAllString(result, `**[$1]($2)**`)

	// Process each term (longest first)
	// Terms are already sorted by length in buildTermInfos() to ensure compound terms
	// like "Preventive Enforcement" are processed before shorter terms like "Enforcement"
	// This prevents shorter terms from matching within longer compound terms
	for _, termInfo := range termInfos {
		// Rebuild skip ranges after each replacement
		linkRanges := linkPattern.FindAllStringIndex(result, -1)
		codeRanges := inlineCodePattern.FindAllStringIndex(result, -1)
		htmlRanges := htmlTagPattern.FindAllStringIndex(result, -1)
		
		// Find HTML tag pairs (opening and closing tags) to skip ALL content between them
		htmlContentRanges := findHTMLContentRanges(result, htmlRanges)

		// Combine skip ranges
		var skipRanges []rangeInfo
		for _, r := range linkRanges {
			skipRanges = append(skipRanges, rangeInfo{start: r[0], end: r[1]})
		}
		for _, r := range codeRanges {
			skipRanges = append(skipRanges, rangeInfo{start: r[0], end: r[1]})
		}
		// Add HTML tags themselves
		for _, r := range htmlRanges {
			skipRanges = append(skipRanges, rangeInfo{start: r[0], end: r[1]})
		}
		// Add ALL content between HTML tag pairs - this completely skips linking terms inside HTML
		skipRanges = append(skipRanges, htmlContentRanges...)

		// Sort by start position
		sort.Slice(skipRanges, func(i, j int) bool {
			return skipRanges[i].start < skipRanges[j].start
		})

		matches := termInfo.Regex.FindAllStringIndex(result, -1)
		if len(matches) == 0 {
			continue
		}

		// Process matches from end to start to preserve indices
		for i := len(matches) - 1; i >= 0; i-- {
			match := matches[i]
			start, end := match[0], match[1]

			// Check if this match is in a skip range (includes existing links, code, HTML)
			shouldSkip := false
			for _, skipRange := range skipRanges {
				// Skip if match overlaps with a skip range at all
				// This prevents linking terms that are inside existing links, code blocks, or HTML
				if start < skipRange.end && end > skipRange.start {
					shouldSkip = true
					break
				}
			}

			if shouldSkip {
				continue
			}

			// Additional safety check: if the matched text itself contains a markdown link pattern
			// This catches edge cases where a link might be embedded in the matched text
			matchedText := result[start:end]
			linkInMatchPattern := regexp.MustCompile(`\[[^\]]+\]\([^\)]+\)`)
			if linkInMatchPattern.MatchString(matchedText) {
				// The matched text contains a markdown link, skip it
				continue
			}
			
			// Check if match overlaps with any link's text portion (the part between [ and ])
			// This prevents linking shorter terms that are part of longer terms that were already linked
			for _, linkRange := range linkRanges {
				linkStart, linkEnd := linkRange[0], linkRange[1]
				// Find the link text portion (between [ and ])
				linkTextStart := linkStart + 1 // Skip opening [
				linkTextEnd := linkTextStart
				for linkTextEnd < linkEnd && result[linkTextEnd] != ']' {
					linkTextEnd++
				}
				// Check if our match overlaps with the link text portion
				// We need to check if the match is completely or partially inside the link text
				if start >= linkTextStart && start < linkTextEnd {
					// Match starts inside link text - skip to avoid nested links
					shouldSkip = true
					break
				}
				if end > linkTextStart && end <= linkTextEnd {
					// Match ends inside link text - skip to avoid nested links
					shouldSkip = true
					break
				}
				if start <= linkTextStart && end >= linkTextEnd {
					// Match completely encompasses link text - skip to avoid nested links
					shouldSkip = true
					break
				}
			}
			
			if shouldSkip {
				continue
			}

			// Create the link
			link := fmt.Sprintf("[%s](%s#%s)", matchedText, defPath, termInfo.Slug)

			// Replace in result
			result = result[:start] + link + result[end:]
		}
	}

	return result
}

// rangeInfo represents a range of text to skip
type rangeInfo struct {
	start, end int
}

// findHTMLContentRanges finds ranges of content between opening and closing HTML tags
func findHTMLContentRanges(text string, tagRanges [][]int) []rangeInfo {
	var contentRanges []rangeInfo
	
	if len(tagRanges) == 0 {
		return contentRanges
	}
	
	// Extract tag information
	type tagInfo struct {
		start, end int
		tagName    string
		isClosing  bool
		isSelfClose bool
	}
	
	var tags []tagInfo
	// Improved regex: matches tag name at the start (after </? and optional whitespace)
	// Handles: <tag>, <tag attr="...">, </tag>, <tag/>, etc.
	tagNamePattern := regexp.MustCompile(`</?\s*([a-zA-Z][a-zA-Z0-9]*)`)
	
	for _, r := range tagRanges {
		tagText := text[r[0]:r[1]]
		matches := tagNamePattern.FindStringSubmatch(tagText)
		if len(matches) < 2 {
			// If we can't extract tag name, skip this tag but still mark it as a skip range
			continue
		}
		
		tagName := matches[1]
		isClosing := strings.HasPrefix(tagText, "</")
		isSelfClose := strings.HasSuffix(tagText, "/>") || (strings.HasSuffix(tagText, " />") && !isClosing)
		
		tags = append(tags, tagInfo{
			start:      r[0],
			end:        r[1],
			tagName:    tagName,
			isClosing:  isClosing,
			isSelfClose: isSelfClose,
		})
	}
	
	// Match opening and closing tags using a stack to handle nesting
	type tagStackItem struct {
		tagName string
		start   int
	}
	var stack []tagStackItem
	
	for _, tag := range tags {
		if tag.isSelfClose {
			// Self-closing tags don't have content
			continue
		}
		
		if !tag.isClosing {
			// Opening tag - push to stack
			stack = append(stack, tagStackItem{
				tagName: tag.tagName,
				start:   tag.end, // Content starts after the opening tag
			})
		} else {
			// Closing tag - find matching opening tag (search from top of stack)
			for i := len(stack) - 1; i >= 0; i-- {
				if stack[i].tagName == tag.tagName {
					// Found matching opening tag
					contentStart := stack[i].start
					contentEnd := tag.start // Content ends before the closing tag
					
					if contentEnd > contentStart {
						contentRanges = append(contentRanges, rangeInfo{
							start: contentStart,
							end:   contentEnd,
						})
					}
					
					// Remove this tag and all tags after it (they were nested inside)
					stack = stack[:i]
					break
				}
			}
		}
	}
	
	// Handle any unclosed tags (opening tags without matching closing tags)
	// This can happen with malformed HTML or tags that span multiple lines
	for _, item := range stack {
		// For unclosed tags, skip content from the tag end to the end of the text
		contentRanges = append(contentRanges, rangeInfo{
			start: item.start,
			end:   len(text),
		})
	}
	
	return contentRanges
}

func cleanupContent(content string, termInfos []TermInfo, defPath string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder

	inCodeBlock := false
	inFrontMatter := false
	frontMatterCount := 0

	for i, line := range lines {
		originalLine := line

		// Track front matter
		if strings.HasPrefix(strings.TrimSpace(line), "---") {
			if !inFrontMatter {
				inFrontMatter = true
				frontMatterCount = 1
			} else {
				frontMatterCount++
				if frontMatterCount == 2 {
					inFrontMatter = false
				}
			}
		}

		// Skip processing in front matter
		if inFrontMatter {
			result.WriteString(originalLine)
			if i < len(lines)-1 {
				result.WriteString("\n")
			}
			continue
		}

		// Track code blocks (fenced with ```)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
		}

		// Skip code blocks
		if inCodeBlock {
			result.WriteString(originalLine)
			if i < len(lines)-1 {
				result.WriteString("\n")
			}
			continue
		}

		// Process the line to remove termlinker-generated links
		processedLine := cleanupLine(line, termInfos, defPath)
		result.WriteString(processedLine)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

func cleanupLine(line string, termInfos []TermInfo, defPath string) string {
	// Create a map of term text to term info for quick lookup
	termMap := make(map[string]TermInfo)
	for _, termInfo := range termInfos {
		termMap[strings.ToLower(termInfo.OriginalTerm)] = termInfo
	}

	// Manually parse markdown links to handle parentheses in link text correctly
	// This is more robust than regex for handling nested parentheses
	var result strings.Builder
	i := 0
	for i < len(line) {
		// Look for the start of a markdown link: [
		if line[i] == '[' {
			// Find the matching closing bracket ]
			bracketStart := i
			bracketEnd := -1
			for j := i + 1; j < len(line); j++ {
				if line[j] == ']' {
					bracketEnd = j
					break
				}
			}

			// If we found a closing bracket, check if it's followed by (
			if bracketEnd != -1 && bracketEnd+1 < len(line) && line[bracketEnd+1] == '(' {
				// Extract link text (everything between [ and ])
				linkText := line[bracketStart+1 : bracketEnd]

				// Find the matching closing parenthesis for the URL
				parenStart := bracketEnd + 1
				parenEnd := -1
				parenDepth := 1
				for j := parenStart + 1; j < len(line); j++ {
					if line[j] == '(' {
						parenDepth++
					} else if line[j] == ')' {
						parenDepth--
						if parenDepth == 0 {
							parenEnd = j
							break
						}
					}
				}

				// If we found a matching closing parenthesis, extract the URL
				if parenEnd != -1 {
					linkURL := line[parenStart+1 : parenEnd]

					// Check if this link points to the definitions page
					if strings.Contains(linkURL, "02-definitions") {
						// Extract the slug from the URL (part after #)
						hashIndex := strings.Index(linkURL, "#")
						if hashIndex != -1 {
							slug := linkURL[hashIndex+1:]

							// Check if the link text matches a term (case-insensitive)
							lowerLinkText := strings.ToLower(linkText)
							termInfo, found := termMap[lowerLinkText]
							if found && termInfo.Slug == slug {
								// This is a termlinker-generated link, remove it and return just the text
								result.WriteString(linkText)
								i = parenEnd + 1
								continue
							}
						}
					}
				}
			}
		}

		// Not a termlinker link, or parsing failed, keep the character
		result.WriteByte(line[i])
		i++
	}

	return result.String()
}
