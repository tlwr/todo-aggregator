package github

import (
	"context"
	"fmt"

	. "github.com/tlwr/todo-aggregator/todo"

	"github.com/google/go-github/v32/github"
)

type githubAssigneeTodo struct {
	id     int64
	name   string
	url    string
	labels []string
}

func (t *githubAssigneeTodo) Name() string {
	return t.name
}

func (t *githubAssigneeTodo) Source() string {
	return "github"
}

func (t *githubAssigneeTodo) Labels() []string {
	return t.labels
}

func (t *githubAssigneeTodo) URL() string {
	return t.url
}

func (t *githubAssigneeTodo) URI() string {
	return fmt.Sprintf("github://issue/%d", t.id)
}

func (t *githubAssigneeTodo) Started() bool {
	return true
}

func (t *githubAssigneeTodo) Finished() bool {
	return false
}

func FetchGitHubAssigneeTodos(
	username string,
) ([]Todo, error) {
	client := github.NewClient(nil)
	search := client.Search

	todos := []Todo{}

	page := 1
	q := fmt.Sprintf("assignee:%s is:open", username)

	for {
		opts := &github.SearchOptions{}
		opts.Page = page

		res, resp, err := search.Issues(context.TODO(), q, opts)
		if err != nil {
			return todos, err
		}

		for _, issue := range res.Issues {
			todos = append(todos, &githubAssigneeTodo{
				id:     *issue.ID,
				name:   *issue.Title,
				url:    *issue.HTMLURL,
				labels: []string{},
			})
		}

		if *res.IncompleteResults {
			page = resp.NextPage
		} else {
			break
		}
	}

	return todos, nil
}

type githubAuthorTodo struct {
	id     int64
	name   string
	url    string
	labels []string
}

func (t *githubAuthorTodo) Name() string {
	return t.name
}

func (t *githubAuthorTodo) Source() string {
	return "github"
}

func (t *githubAuthorTodo) Labels() []string {
	return t.labels
}

func (t *githubAuthorTodo) URL() string {
	return t.url
}

func (t *githubAuthorTodo) URI() string {
	return fmt.Sprintf("github://pull-request/%d", t.id)
}

func (t *githubAuthorTodo) Started() bool {
	return true
}

func (t *githubAuthorTodo) Finished() bool {
	return false
}

func FetchGitHubAuthorTodos(
	username string,
) ([]Todo, error) {
	client := github.NewClient(nil)
	search := client.Search

	todos := []Todo{}

	page := 1
	q := fmt.Sprintf("is:pr author:%s is:open", username)

	for {
		opts := &github.SearchOptions{}
		opts.Page = page

		res, resp, err := search.Issues(context.TODO(), q, opts)
		if err != nil {
			return todos, err
		}

		for _, pr := range res.Issues {
			todos = append(todos, &githubAuthorTodo{
				id:     *pr.ID,
				name:   *pr.Title,
				url:    *pr.HTMLURL,
				labels: []string{},
			})
		}

		if *res.IncompleteResults {
			page = resp.NextPage
		} else {
			break
		}
	}

	return todos, nil
}
