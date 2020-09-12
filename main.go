package main

import (
	"log"

	"github.com/tlwr/todo-aggregator/agg/pivotal"
	. "github.com/tlwr/todo-aggregator/todo"
)

func main() {
	pivotalTrackerProjects := []string{"1133984"}

	todos := []Todo{}

	pivotalTodos, err := pivotal.FetchPivotalTodos(pivotalTrackerProjects)
	if err != nil {
		log.Fatal(err)
	}
	todos = append(todos, pivotalTodos...)

	for _, todo := range todos {
		log.Printf("pivotal: %s %s", todo.Name(), todo.URL())
	}
}
