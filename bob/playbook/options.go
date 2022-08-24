package playbook

type Option func(p *Playbook)

func WithCachingEnabled(enable bool) Option {
	return func(p *Playbook) {
		p.enableCaching = enable
	}
}

func WithPredictedNumOfTasks(tasks int) Option {
	return func(p *Playbook) {
		p.predictedNumOfTasks = tasks
	}
}
