package bobrun

func (r *Run) Init() []string {
	return r.init
}

func (r *Run) InitOnce() []string {
	return r.initOnce
}
