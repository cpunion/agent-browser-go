package agentbrowser

import (
	"fmt"
	"strings"
	"sync/atomic"
)

// RefMap maps ref IDs to element info.
type RefMap map[string]RefData

// RefData contains information about a referenced element.
type RefData struct {
	Selector string `json:"selector"`
	Role     string `json:"role"`
	Name     string `json:"name,omitempty"`
	Nth      int    `json:"nth,omitempty"`
}

// EnhancedSnapshot contains the accessibility tree with refs.
type EnhancedSnapshot struct {
	Tree string `json:"tree"`
	Refs RefMap `json:"refs"`
}

// SnapshotOptions configures snapshot generation.
type SnapshotOptions struct {
	Interactive bool   `json:"interactive,omitempty"`
	MaxDepth    int    `json:"maxDepth,omitempty"`
	Compact     bool   `json:"compact,omitempty"`
	Selector    string `json:"selector,omitempty"`
}

// Role classifications
var (
	// InteractiveRoles are roles that get refs and are included in interactive-only mode.
	InteractiveRoles = map[string]bool{
		"button":           true,
		"link":             true,
		"textbox":          true,
		"checkbox":         true,
		"radio":            true,
		"combobox":         true,
		"listbox":          true,
		"menuitem":         true,
		"menuitemcheckbox": true,
		"menuitemradio":    true,
		"option":           true,
		"searchbox":        true,
		"slider":           true,
		"spinbutton":       true,
		"switch":           true,
		"tab":              true,
		"treeitem":         true,
	}

	// ContentRoles are roles that provide structure/context.
	ContentRoles = map[string]bool{
		"heading":      true,
		"cell":         true,
		"gridcell":     true,
		"columnheader": true,
		"rowheader":    true,
		"listitem":     true,
		"article":      true,
		"region":       true,
		"main":         true,
		"navigation":   true,
	}

	// StructuralRoles are purely structural elements.
	StructuralRoles = map[string]bool{
		"generic":      true,
		"group":        true,
		"list":         true,
		"table":        true,
		"row":          true,
		"rowgroup":     true,
		"grid":         true,
		"treegrid":     true,
		"menu":         true,
		"menubar":      true,
		"toolbar":      true,
		"tablist":      true,
		"tree":         true,
		"directory":    true,
		"document":     true,
		"application":  true,
		"presentation": true,
		"none":         true,
	}
)

// refCounter for generating unique refs.
var refCounter atomic.Int64

// resetRefs resets the ref counter.
func resetRefs() {
	refCounter.Store(0)
}

// nextRef generates the next ref ID.
func nextRef() string {
	return fmt.Sprintf("e%d", refCounter.Add(1))
}

// buildSelector creates a selector string for a role+name.
func buildSelector(role, name string) string {
	if name != "" {
		escapedName := strings.ReplaceAll(name, `"`, `\"`)
		return fmt.Sprintf(`[role="%s"][aria-label="%s"]`, role, escapedName)
	}
	return fmt.Sprintf(`[role="%s"]`, role)
}

// AXNode represents an accessibility node from the browser.
type AXNode struct {
	Role       string                 `json:"role"`
	Name       string                 `json:"name"`
	Children   []*AXNode              `json:"children"`
	Properties map[string]interface{} `json:"properties"`
}

// BuildSnapshotFromNodes builds an enhanced snapshot from a raw accessibility tree.
func BuildSnapshotFromNodes(root *AXNode, opts SnapshotOptions) *EnhancedSnapshot {
	resetRefs()
	refs := make(RefMap)

	if root == nil {
		return &EnhancedSnapshot{Tree: "(empty)", Refs: refs}
	}

	// Track role+name combinations for nth handling
	roleNameCounts := make(map[string]int)

	// Build tree
	var builder strings.Builder
	buildTreeNodeFromAX(&builder, root, refs, roleNameCounts, opts, 0)

	tree := builder.String()
	if tree == "" {
		if opts.Interactive {
			tree = "(no interactive elements)"
		} else {
			tree = "(empty)"
		}
	}

	return &EnhancedSnapshot{Tree: strings.TrimSpace(tree), Refs: refs}
}

// buildTreeNodeFromAX recursively builds the tree representation.
func buildTreeNodeFromAX(
	builder *strings.Builder,
	node *AXNode,
	refs RefMap,
	roleNameCounts map[string]int,
	opts SnapshotOptions,
	depth int,
) {
	if node == nil {
		return
	}

	// Check max depth
	if opts.MaxDepth > 0 && depth > opts.MaxDepth {
		return
	}

	role := strings.ToLower(node.Role)
	name := node.Name

	isInteractive := InteractiveRoles[role]
	isContent := ContentRoles[role]
	isStructural := StructuralRoles[role]

	// Filter for interactive-only mode
	if opts.Interactive && !isInteractive {
		// Still process children to find interactive elements
		for _, child := range node.Children {
			buildTreeNodeFromAX(builder, child, refs, roleNameCounts, opts, depth)
		}
		return
	}

	// Skip unnamed structural elements in compact mode
	if opts.Compact && isStructural && name == "" {
		for _, child := range node.Children {
			buildTreeNodeFromAX(builder, child, refs, roleNameCounts, opts, depth)
		}
		return
	}

	// Skip generic/none roles without names
	if (role == "generic" || role == "none") && name == "" {
		for _, child := range node.Children {
			buildTreeNodeFromAX(builder, child, refs, roleNameCounts, opts, depth)
		}
		return
	}

	// Build the line
	indent := strings.Repeat("  ", depth)

	// Determine if this node should have a ref
	shouldHaveRef := isInteractive || (isContent && name != "")

	var ref string
	var nth int
	if shouldHaveRef {
		ref = nextRef()
		key := fmt.Sprintf("%s:%s", role, name)
		nth = roleNameCounts[key]
		roleNameCounts[key]++

		refs[ref] = RefData{
			Selector: buildSelector(role, name),
			Role:     role,
			Name:     name,
			Nth:      nth,
		}
	}

	// Build the line content
	line := fmt.Sprintf("%s- %s", indent, role)
	if name != "" {
		line += fmt.Sprintf(` "%s"`, name)
	}
	if ref != "" {
		line += fmt.Sprintf(" [ref=%s]", ref)
		if nth > 0 {
			line += fmt.Sprintf(" [nth=%d]", nth)
		}
	}

	// Add properties like level for headings
	if role == "heading" && node.Properties != nil {
		if level, ok := node.Properties["level"]; ok {
			if v, ok := level.(float64); ok {
				line += fmt.Sprintf(" [level=%d]", int(v))
			}
		}
	}

	builder.WriteString(line)
	builder.WriteString("\n")

	// Process children
	for _, child := range node.Children {
		buildTreeNodeFromAX(builder, child, refs, roleNameCounts, opts, depth+1)
	}
}

// GetSnapshotStats returns statistics about a snapshot.
func GetSnapshotStats(snapshot *EnhancedSnapshot) map[string]int {
	interactiveCount := 0
	for _, ref := range snapshot.Refs {
		if InteractiveRoles[ref.Role] {
			interactiveCount++
		}
	}

	return map[string]int{
		"lines":       len(strings.Split(snapshot.Tree, "\n")),
		"chars":       len(snapshot.Tree),
		"tokens":      len(snapshot.Tree) / 4,
		"refs":        len(snapshot.Refs),
		"interactive": interactiveCount,
	}
}
