package target

// // Exists determines if the target exists without
// // validating it's integrety.
// func (t *T) Exists() bool {
// 	return t.existsFile() && t.existsDocker()
// }

// func (t *T) existsFile() bool {
// 	if len(*t.filesystemEntries) == 0 {
// 		return true
// 	}
// 	// check plain existence
// 	for _, f := range t.filesystemEntries {
// 		target := filepath.Join(t.dir, f)
// 		if !file.Exists(target) {
// 			return false
// 		}
// 	}

// 	return true
// }

func (t *T) existsDocker() bool {
	if len(t.dockerImages) == 0 {
		return true
	}

	for _, f := range t.dockerImages {
		exists, err := t.dockerRegistryClient.ImageExists(f)
		if err != nil {
			return false
		}
		if !exists {
			return false
		}
	}

	return true
}
