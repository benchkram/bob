package bobfile_test

import (
	"errors"
	"testing"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/bobrun"
	"github.com/benchkram/bob/bobtask"
)

func TestBobfileValidateSelReference(t *testing.T) {
	b := bobfile.NewBobfile()

	b.BTasks["one"] = bobtask.Task{
		DependsOn: []string{"one"},
	}

	if err := b.Validate(); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestBobfileValidateDuplicateName(t *testing.T) {
	b := bobfile.NewBobfile()

	b.BTasks["one"] = bobtask.Task{}

	b.RTasks["one"] = &bobrun.Run{}

	if err := b.Validate(); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestBobfileValidateInvalidVersion(t *testing.T) {
	b := bobfile.NewBobfile()

	b.Version = "invalid-version"

	if err := b.Validate(); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestBobfileValidateValidVersion(t *testing.T) {
	b := bobfile.NewBobfile()

	b.Version = "1.2.3"

	if err := b.Validate(); err != nil {
		t.Error("Expected nil, got error")
	}
}

func TestBobfileProjectName(t *testing.T) {
	var tests = []struct {
		name     string
		expected error
	}{
		{"simple-project", nil},
		{"project-with-inv@lid-chars", bobfile.ErrInvalidProjectName},
		{"bob.build/user/url-project", nil},
		{"https://bob.build/user/schema-url-project", bobfile.ErrInvalidProjectName},
	}

	for _, tt := range tests {
		b := bobfile.NewBobfile()
		b.Project = tt.name

		err := b.Validate()

		if !errors.Is(err, tt.expected) {
			t.Errorf("ValidateProjectName(%s): expected %q, got %q", tt.name, tt.expected, err)
		}
	}
}
