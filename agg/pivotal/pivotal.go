package pivotal

import (
	"fmt"
	"strconv"

	. "github.com/tlwr/todo-aggregator/todo"

	"github.com/salsita/go-pivotaltracker/v5/pivotal"
)

type pivotalTodo struct {
	storyID   string
	projectID string

	name   string
	labels []string

	url string

	currentState string
}

func (t *pivotalTodo) Name() string {
	return t.name
}

func (t *pivotalTodo) Source() string {
	return "pivotal"
}

func (t *pivotalTodo) Labels() []string {
	return t.labels
}

func (t *pivotalTodo) URL() string {
	return t.url
}

func (t *pivotalTodo) URI() string {
	return fmt.Sprintf("pivotal://%s/%s", t.projectID, t.storyID)
}

func (t *pivotalTodo) Started() bool {
	switch t.currentState {
	case "rejected", "accepted", "delivered", "finished", "started":
		return true
	default:
		return false
	}
}

func (t *pivotalTodo) Finished() bool {
	switch t.currentState {
	case "rejected", "accepted":
		return true
	default:
		return false
	}
}

func FetchPivotalTodos(
	apiKey string,
	ownerIDs []string,
	projectIDs []string,
) ([]Todo, error) {
	client := pivotal.NewClient(apiKey)

	todos := []Todo{}

	owners := []int{}
	for _, oid := range ownerIDs {
		owner, err := strconv.Atoi(oid)
		if err != nil {
			return nil, err
		}
		owners = append(owners, owner)
	}

	for _, pid := range projectIDs {
		projectID, err := strconv.Atoi(pid)
		if err != nil {
			return nil, err
		}

		stories, err := client.Stories.List(projectID, "-estimated:-1")

		if err != nil {
			return nil, err
		}

		for _, story := range stories {
			owned := false

			for _, storyOwner := range story.OwnerIDs {
				for _, owner := range owners {
					if storyOwner == owner {
						owned = true
					}
				}
			}

			if !owned {
				continue
			}

			todos = append(todos, &pivotalTodo{
				storyID:   fmt.Sprintf("%d", story.ID),
				projectID: fmt.Sprintf("%d", projectID),

				name: story.Name,

				url: story.URL,

				currentState: story.State,
			})
		}
	}

	return todos, nil
}
