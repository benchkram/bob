package bob

func (b B) Clean() error {
	return b.cleanBuildInfoStore()
}

func (b B) cleanBuildInfoStore() error {
	return b.buildInfoStore.Clean()
}
