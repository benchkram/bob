package bobtask

// inputHashVersion is an hardcoded stringified number.
// Used to create unique input hashes between version.
// A bump of this const indicates a incompatible
// buildinfo && artifact changes and with previous version.
//
// Increment on each incompatible change and document it here.
//
//	"1" - 1. apr 2023
const inputHashVersion = "1"
