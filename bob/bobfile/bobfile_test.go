package bobfile_test

import (
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
