package trello_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tlwr/todo-aggregator/agg/trello"

	"github.com/jarcoal/httpmock"
)

func TestPivotal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Trello Suite")
}

var _ = Describe("Trello", func() {
	const (
		apiKey   = "test"
		apiToken = "test"
	)

	BeforeEach(func() {
		httpmock.Activate()
		httpmock.Reset()
	})

	AfterEach(func() {
		httpmock.DeactivateAndReset()
	})

	It("should return some stories", func() {
		httpmock.RegisterResponder(
			"GET",
			"https://api.trello.com/1/search",
			httpmock.NewStringResponder(
				200,
				`{
  "cards": [
    {
      "id": "5f5bb1876a371454ad43f678",
      "desc": "a-description",
      "idBoard": "a-board",
      "idList": "a-list",
      "idShort": 123,
      "name": "a-story",
      "labels": []
    },
    {
      "id": "5f5bb1876a371454ad43f677",
      "desc": "another-description",
      "idBoard": "another-board",
      "idList": "another-list",
      "idShort": 123,
      "name": "another-story",
      "labels": []
    }
  ]
}
  `,
			),
		)

		By("Getting stories")
		todos, err := trello.FetchTrelloTodos(
			apiKey, apiToken,
			[]string{"a-username"}, /* owners */
		)
		Expect(err).NotTo(HaveOccurred())

		Expect(todos).To(HaveLen(2))

		By("Checking names")
		Expect(todos[0].Name()).To(Equal("a-story"))
		Expect(todos[1].Name()).To(Equal("another-story"))
	})
})
