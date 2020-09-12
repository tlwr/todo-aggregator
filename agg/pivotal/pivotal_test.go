package pivotal_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tlwr/todo-aggregator/agg/pivotal"

	"github.com/jarcoal/httpmock"
)

func TestPivotal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pivotal Suite")
}

var _ = Describe("Pivotal", func() {
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
			"https://www.pivotaltracker.com/services/v5/projects/1234/stories",
			httpmock.NewStringResponder(
				200,
				`[{
    "kind": "story",
    "id": 12345,
    "story_type": "feature",
    "name": "a story",
    "current_state": "accepted",
    "url": "https://www.pivotaltracker.com/story/show/12345",
    "project_id": 1234,
    "labels": []
  }, {
    "kind": "story",
    "id": 123456,
    "story_type": "feature",
    "name": "another story",
    "current_state": "started",
    "url": "https://www.pivotaltracker.com/story/show/123456",
    "project_id": 1234,
    "labels": []
  }]`,
			),
		)

		By("Getting stories")
		todos, err := pivotal.FetchPivotalTodos([]string{"1234"})
		Expect(err).NotTo(HaveOccurred())

		Expect(todos).To(HaveLen(2))

		By("Checking names")
		Expect(todos[0].Name()).To(Equal("a story"))
		Expect(todos[1].Name()).To(Equal("another story"))

		By("Checking URLs")
		Expect(todos[0].URL()).To(Equal("https://www.pivotaltracker.com/story/show/12345"))
		Expect(todos[1].URL()).To(Equal("https://www.pivotaltracker.com/story/show/123456"))
	})
})
