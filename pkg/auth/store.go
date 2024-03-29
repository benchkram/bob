package auth

import (
	"errors"
	"os"
	"path"

	"github.com/benchkram/errz"
	"gopkg.in/yaml.v3"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Store struct {
	dir        string
	configPath string
}

type YamlAuthContext struct {
	Name    string `yaml:"name"`
	Token   string `yaml:"token"`
	Current bool   `yaml:"current"`
}

// New creates a filestore. The caller is responsible to pass an
// existing directory.
func New(dir string) *Store {
	s := &Store{
		dir:        dir,
		configPath: path.Join(dir, "auth.yaml"),
	}

	return s
}

func (s *Store) Contexts() (names []Context, err error) {
	defer errz.Recover(&err)

	ctxs, err := s.contexts()
	errz.Fatal(err)

	names = []Context{}
	for _, c := range ctxs {
		names = append(names, Context{
			Name:    c.Name,
			Token:   c.Token,
			Current: c.Current,
		})
	}

	return names, nil
}

func (s *Store) Context(name string) (ctx Context, err error) {
	defer errz.Recover(&err)

	ctxs, err := s.contexts()
	errz.Fatal(err)

	for _, c := range ctxs {
		if c.Name == name {
			return Context{
				Name:  c.Name,
				Token: c.Token,
			}, nil
		}
	}

	errz.Fatal(ErrNotFound)
	return
}

func (s *Store) CreateContext(name, token string) (err error) {
	defer errz.Recover(&err)

	ctxs, err := s.contexts()
	errz.Fatal(err)

	var exists bool
	for _, c := range ctxs {
		if c.Name == name {
			exists = true
			break
		}
	}

	if exists {
		errz.Fatal(ErrAlreadyExists)
	}

	ctxs = append(ctxs, &YamlAuthContext{
		Name:    name,
		Token:   token,
		Current: len(ctxs) == 0,
	})

	return s.save(ctxs)
}
func (s *Store) DeleteContext(name string) (err error) {
	defer errz.Recover(&err)

	ctxs, err := s.contexts()
	errz.Fatal(err)

	var found bool
	for i, c := range ctxs {
		if c.Name == name {
			found = true

			// remove this context
			ctxs = append(ctxs[:i], ctxs[i+1:]...)
			if c.Current && len(ctxs) > 0 {
				// mark the next context as current (if any left)
				ctxs[len(ctxs)-1].Current = true
			}

			break
		}
	}

	if !found {
		errz.Fatal(ErrNotFound)
	}

	return s.save(ctxs)
}

func (s *Store) CurrentContext() (c Context, err error) {
	defer errz.Recover(&err)

	ctxs, err := s.contexts()
	errz.Fatal(err)

	for _, c := range ctxs {
		if c.Current {
			return Context{
				Name:  c.Name,
				Token: c.Token,
			}, nil
		}
	}

	return Context{}, ErrNotFound
}

func (s *Store) SetCurrentContext(name string) (err error) {
	defer errz.Recover(&err)

	ctxs, err := s.contexts()
	errz.Fatal(err)

	var exists bool
	for _, c := range ctxs {
		if c.Name == name {
			exists = true
		}
	}

	if !exists {
		errz.Fatal(ErrNotFound)
	}

	for _, c := range ctxs {
		if c.Name == name {
			c.Current = true
		} else {
			c.Current = false
		}
	}

	return s.save(ctxs)
}

func (s *Store) UpdateContext(name, token string) (err error) {
	defer errz.Recover(&err)

	ctxs, err := s.contexts()
	errz.Fatal(err)

	var found bool
	for _, c := range ctxs {
		if c.Name == name {
			c.Token = token
			found = true
			break
		}
	}

	if !found {
		errz.Fatal(ErrNotFound)
	}

	return s.save(ctxs)
}

func (s *Store) contexts() (ctxs []*YamlAuthContext, err error) {
	defer errz.Recover(&err)

	b, err := os.ReadFile(path.Join(s.dir, "auth.yaml"))
	if errors.Is(err, os.ErrNotExist) {
		return []*YamlAuthContext{}, nil
	}
	errz.Fatal(err)

	err = yaml.Unmarshal(b, &ctxs)
	errz.Fatal(err)

	return ctxs, nil
}

func (s *Store) save(ctxs []*YamlAuthContext) (err error) {
	defer errz.Recover(&err)

	b, err := yaml.Marshal(ctxs)
	errz.Fatal(err)

	err = os.WriteFile(s.configPath, b, os.ModePerm)
	errz.Fatal(err)

	return nil
}
