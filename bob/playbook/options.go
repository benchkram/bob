package playbook

import (
	"github.com/benchkram/bob/pkg/store"
)

type Option func(p *Playbook)

func WithCachingEnabled(enable bool) Option {
	return func(p *Playbook) {
		p.enableCaching = enable
	}
}

func WithPushEnabled(enable bool) Option {
	return func(p *Playbook) {
		p.enablePush = enable
	}
}

func WithPullEnabled(enable bool) Option {
	return func(p *Playbook) {
		p.enablePull = enable
	}
}

func WithPredictedNumOfTasks(tasks int) Option {
	return func(p *Playbook) {
		p.predictedNumOfTasks = tasks
	}
}

func WithMaxParallel(maxParallel int) Option {
	return func(p *Playbook) {
		p.maxParallel = maxParallel
	}
}

func WithRemoteStore(s store.Store) Option {
	return func(p *Playbook) {
		p.remoteStore = s
	}
}

func WithLocalStore(s store.Store) Option {
	return func(p *Playbook) {
		p.localStore = s
	}
}
