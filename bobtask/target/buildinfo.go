package target

// func (t *T) Buildinfo() (bi buildinfo.TargetBuildInfo, err error) {

// 	for _, f := range t.PathsSerialize {
// 		target := filepath.Join(t.dir, f)

// 		if !file.Exists(target) {
// 			return empty, usererror.Wrapm(fmt.Errorf("target does not exist %q", f), "failed to hash target")
// 		}
// 		fi, err := os.Stat(target)
// 		if err != nil {
// 			return empty, fmt.Errorf("failed to get file info %q: %w", f, err)
// 		}
// 	}

// 	return bi, nil
// }
