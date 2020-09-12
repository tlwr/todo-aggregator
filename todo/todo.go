package todo

import (
	"fmt"
)

// Todo represents a generic cross-platform Todo
type Todo interface {
	Name() string
	Source() string

	Labels() []string

	URI() string
	URL() string

	Started() bool
	Finished() bool
}

type fakeTodo struct {
	id string

	name   string
	labels []string

	started  bool
	finished bool
}

func (t *fakeTodo) Name() string {
	return t.name
}

func (t *fakeTodo) Source() string {
	return "fake"
}

func (t *fakeTodo) Labels() []string {
	return t.labels
}

func (t *fakeTodo) URI() string {
	return fmt.Sprintf("fake-todo://%s", t.id)
}

func (t *fakeTodo) URL() string {
	return fmt.Sprintf("https://fake-todo-list.local/%s", t.id)
}

func (t *fakeTodo) Started() bool {
	return t.started
}

func (t *fakeTodo) Finished() bool {
	return t.finished
}
