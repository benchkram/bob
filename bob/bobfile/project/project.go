package project

import "regexp"

// RestrictedProjectNameChars collects the characters allowed to represent a project.
const RestrictedProjectNameChars = `[a-zA-Z0-9/_.\-:]`

// RestrictedProjectNamePattern is a regular expression to validate projectnames.
var RestrictedProjectNamePattern = regexp.MustCompile(`^` + RestrictedProjectNameChars + `+$`)

// ProjectNameDoubleSlashPattern matches a string containing a double slash (useful to check for URL schema)
var ProjectNameDoubleSlashPattern = regexp.MustCompile(`//+`)
