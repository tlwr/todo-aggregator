package todo

import (
	"fmt"
)

// Todo represents a generic cross-platform Todo
type Todo interface {
	Name() string

	Labels() []string

	URI() string
	URL() string
}

type fakeTodo struct {
	id string

	name   string
	labels []string
}

func (t *fakeTodo) Name() string {
	return t.name
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
