package bobsync

func CheckForConflicts(current []Sync, new Sync) error {
	for _, s := range current {
		if s.Path == new.Path {
			return ErrSyncPathTaken
		}
	}
	return nil
}
