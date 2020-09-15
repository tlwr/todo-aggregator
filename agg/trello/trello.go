package trello

import (
	"fmt"

	. "github.com/tlwr/todo-aggregator/todo"

	"github.com/adlio/trello"
)

type trelloTodo struct {
	cardID  string
	boardID string

	name   string
	labels []string

	url string
}

func (t *trelloTodo) Name() string {
	return t.name
}

func (t *trelloTodo) Source() string {
	return "trello"
}

func (t *trelloTodo) Labels() []string {
	return t.labels
}

func (t *trelloTodo) URL() string {
	return t.url
}

func (t *trelloTodo) URI() string {
	return fmt.Sprintf("trello://%s/%s", t.boardID, t.cardID)
}

func (t *trelloTodo) Started() bool {
	return true
}

func (t *trelloTodo) Finished() bool {
	return false
}

func FetchTrelloTodos(
	apiKey string,
	apiToken string,
	usernames []string,
) ([]Todo, error) {
	client := trello.NewClient(apiKey, apiToken)

	todos := []Todo{}

	q := ""

	for _, username := range usernames {
		q += fmt.Sprintf("@%s ", username)
	}

	q += "-archived "

	cards, err := client.SearchCards(q, trello.Defaults())
	if err != nil {
		return todos, err
	}

	for _, card := range cards {
		labels := []string{}
		for _, label := range card.Labels {
			labels = append(labels, label.Name)
		}

		todos = append(todos, &trelloTodo{
			cardID:  card.ID,
			boardID: card.IDBoard,

			name:   card.Name,
			labels: labels,

			url: card.URL,
		})
	}

	return todos, nil
}
