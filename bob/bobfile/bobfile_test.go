package bobfile_test

import (
	"github.com/Benchkram/bob/bob/bobfile"
	"github.com/Benchkram/bob/bobrun"
	"github.com/Benchkram/bob/bobtask"
	"testing"
)

func TestBobfileValidateSelReference(t *testing.T) {
	b := bobfile.NewBobfile()

	b.Tasks["one"] = bobtask.Task{
		DependsOn: []string{"one"},
	}

	if err := b.Validate(); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestBobfileValidateDuplicateName(t *testing.T) {
	b := bobfile.NewBobfile()

	b.Tasks["one"] = bobtask.Task{}

	b.Runs["one"] = &bobrun.Run{}

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