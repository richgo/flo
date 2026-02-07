package spec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		wantValid       bool
		wantMissing     []string
		wantErrorCount  int
	}{
		{
			name: "valid spec with all required sections",
			content: `# My Feature

## Goal
Build a new feature.

## Context
This is needed because...

## Success Criteria
- Feature works
- Tests pass
`,
			wantValid:       true,
			wantMissing:     []string{},
			wantErrorCount:  0,
		},
		{
			name: "missing Goal section",
			content: `# My Feature

## Context
This is needed because...

## Success Criteria
- Feature works
`,
			wantValid:       false,
			wantMissing:     []string{"Goal"},
			wantErrorCount:  0,
		},
		{
			name: "missing Context section",
			content: `# My Feature

## Goal
Build a new feature.

## Success Criteria
- Feature works
`,
			wantValid:       false,
			wantMissing:     []string{"Context"},
			wantErrorCount:  0,
		},
		{
			name: "missing Success Criteria section",
			content: `# My Feature

## Goal
Build a new feature.

## Context
This is needed because...
`,
			wantValid:       false,
			wantMissing:     []string{"Success Criteria"},
			wantErrorCount:  0,
		},
		{
			name:            "empty file",
			content:         "",
			wantValid:       false,
			wantMissing:     []string{"Goal", "Context", "Success Criteria"},
			wantErrorCount:  1,
		},
		{
			name:            "no markdown headings",
			content:         "This is just plain text without any headings.",
			wantValid:       false,
			wantMissing:     []string{"Goal", "Context", "Success Criteria"},
			wantErrorCount:  1,
		},
		{
			name: "case insensitive section matching",
			content: `# My Feature

## GOAL
Build a new feature.

## context
This is needed because...

## Success criteria
- Feature works
`,
			wantValid:       true,
			wantMissing:     []string{},
			wantErrorCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			result := v.Validate(tt.content)

			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.wantValid)
			}

			if len(result.MissingSections) != len(tt.wantMissing) {
				t.Errorf("MissingSections count = %d, want %d", len(result.MissingSections), len(tt.wantMissing))
			}

			for _, missing := range tt.wantMissing {
				found := false
				for _, m := range result.MissingSections {
					if m == missing {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected missing section %q not found in %v", missing, result.MissingSections)
				}
			}

			if len(result.Errors) != tt.wantErrorCount {
				t.Errorf("Errors count = %d, want %d. Errors: %v", len(result.Errors), tt.wantErrorCount, result.Errors)
			}
		})
	}
}

func TestValidator_ValidateFile(t *testing.T) {
	v := NewValidator()

	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Test valid file
	validPath := filepath.Join(tmpDir, "valid_spec.md")
	validContent := `# Test Feature

## Goal
Build a test feature.

## Context
Testing the validator.

## Success Criteria
- Validation passes
`
	if err := os.WriteFile(validPath, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := v.ValidateFile(validPath)
	if err != nil {
		t.Errorf("ValidateFile() unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("ValidateFile() Valid = false, want true. Missing: %v, Errors: %v", 
			result.MissingSections, result.Errors)
	}

	// Test non-existent file
	_, err = v.ValidateFile(filepath.Join(tmpDir, "nonexistent.md"))
	if err == nil {
		t.Error("ValidateFile() expected error for non-existent file, got nil")
	}

	// Test invalid file
	invalidPath := filepath.Join(tmpDir, "invalid_spec.md")
	invalidContent := `# Test Feature

## Goal
Build a test feature.
`
	if err := os.WriteFile(invalidPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err = v.ValidateFile(invalidPath)
	if err != nil {
		t.Errorf("ValidateFile() unexpected error: %v", err)
	}
	if result.Valid {
		t.Error("ValidateFile() Valid = true, want false for invalid spec")
	}
	if len(result.MissingSections) == 0 {
		t.Error("ValidateFile() expected missing sections, got none")
	}
}
