// Package spec provides validation for SPEC.md format.
package spec

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// RequiredSections lists the sections that must be present in a valid SPEC.md
var RequiredSections = []string{"Goal", "Context", "Success Criteria"}

// ValidationResult represents the outcome of spec validation.
type ValidationResult struct {
	Valid          bool
	MissingSections []string
	Errors         []string
}

// Validator validates SPEC.md files.
type Validator struct{}

// NewValidator creates a new spec validator.
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateFile validates a SPEC.md file at the given path.
func (v *Validator) ValidateFile(path string) (*ValidationResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	return v.Validate(string(content)), nil
}

// Validate validates the content of a SPEC.md file.
func (v *Validator) Validate(content string) *ValidationResult {
	result := &ValidationResult{
		Valid:          true,
		MissingSections: []string{},
		Errors:         []string{},
	}

	// Parse markdown structure
	sections := v.extractSections(content)

	// Check for required sections
	for _, required := range RequiredSections {
		if !v.hasSectionCaseInsensitive(sections, required) {
			result.MissingSections = append(result.MissingSections, required)
			result.Valid = false
		}
	}

	// Validate markdown structure
	if err := v.validateMarkdownStructure(content); err != nil {
		result.Errors = append(result.Errors, err.Error())
		result.Valid = false
	}

	return result
}

// extractSections parses markdown headings from content.
func (v *Validator) extractSections(content string) []string {
	var sections []string
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			// Extract heading text
			heading := strings.TrimSpace(strings.TrimLeft(line, "#"))
			if heading != "" {
				sections = append(sections, heading)
			}
		}
	}

	return sections
}

// hasSectionCaseInsensitive checks if a section exists (case-insensitive).
func (v *Validator) hasSectionCaseInsensitive(sections []string, target string) bool {
	targetLower := strings.ToLower(target)
	for _, section := range sections {
		if strings.ToLower(section) == targetLower {
			return true
		}
	}
	return false
}

// validateMarkdownStructure checks for basic markdown validity.
func (v *Validator) validateMarkdownStructure(content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("spec file is empty")
	}

	// Check if it starts with a heading
	lines := strings.Split(content, "\n")
	foundHeading := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			foundHeading = true
			break
		}
		// First non-empty line should be a heading
		if !foundHeading {
			return fmt.Errorf("spec should start with a markdown heading (# Title)")
		}
	}

	if !foundHeading {
		return fmt.Errorf("spec contains no markdown headings")
	}

	return nil
}
