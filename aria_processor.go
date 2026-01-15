package agentbrowser

import (
	"fmt"
	"regexp"
	"strings"
)

// processAriaTree processes ARIA snapshot string and adds refs
// This matches the TypeScript processAriaTree function
func processAriaTree(ariaTree string, opts SnapshotOptions) *EnhancedSnapshot {
	resetRefs()
	refs := make(RefMap)

	lines := strings.Split(ariaTree, "\n")
	var result []string
	roleNameCounts := make(map[string]int)

	// Process each line
	for _, line := range lines {
		processed := processAriaLine(line, refs, roleNameCounts, opts)
		if processed != "" {
			result = append(result, processed)
		}
	}

	tree := strings.Join(result, "\n")
	if tree == "" {
		if opts.Interactive {
			tree = "(no interactive elements)"
		} else {
			tree = "(empty)"
		}
	}

	return &EnhancedSnapshot{
		Tree: strings.TrimSpace(tree),
		Refs: refs,
	}
}

// processAriaLine processes a single line from ARIA snapshot
func processAriaLine(line string, refs RefMap, roleNameCounts map[string]int, opts SnapshotOptions) string {
	// Match lines like:
	//   - button "Submit"
	//   - heading "Title" [level=1]
	//   - link "Click me":
	re := regexp.MustCompile(`^(\s*-\s*)(\w+)(?:\s+"([^"]*)")?(.*)$`)
	match := re.FindStringSubmatch(line)

	if match == nil {
		// Not a role line (metadata or text content)
		if opts.Interactive {
			return "" // Skip in interactive mode
		}
		return line
	}

	prefix := match[1]
	role := match[2]
	name := match[3]
	suffix := match[4]

	roleLower := strings.ToLower(role)

	// Skip metadata lines
	if strings.HasPrefix(role, "/") {
		return line
	}

	isInteractive := InteractiveRoles[roleLower]
	isContent := ContentRoles[roleLower]

	// Filter for interactive-only mode
	if opts.Interactive && !isInteractive {
		return ""
	}

	// Add ref for interactive or named content elements
	shouldHaveRef := isInteractive || (isContent && name != "")

	if shouldHaveRef {
		ref := nextRef()
		key := fmt.Sprintf("%s:%s", roleLower, name)
		nth := roleNameCounts[key]
		roleNameCounts[key]++

		refs[ref] = RefData{
			Selector: buildSelector(roleLower, name),
			Role:     roleLower,
			Name:     name,
			Nth:      nth,
		}

		// Build enhanced line with ref
		enhanced := fmt.Sprintf("%s%s", prefix, role)
		if name != "" {
			enhanced += fmt.Sprintf(` "%s"`, name)
		}
		enhanced += fmt.Sprintf(" [ref=%s]", ref)
		if nth > 0 {
			enhanced += fmt.Sprintf(" [nth=%d]", nth)
		}
		if suffix != "" {
			enhanced += suffix
		}

		return enhanced
	}

	return line
}
