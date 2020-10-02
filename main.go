package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/tlwr/todo-aggregator/agg/github"
	"github.com/tlwr/todo-aggregator/agg/pivotal"
	"github.com/tlwr/todo-aggregator/agg/trello"
	"github.com/tlwr/todo-aggregator/todo"

	nlogrus "github.com/meatballhat/negroni-logrus"
	"github.com/phyber/negroni-gzip/gzip"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sethvargo/go-signalcontext"
	"github.com/sirupsen/logrus"
	"github.com/unrolled/render"
	nsecure "github.com/unrolled/secure"
	"github.com/urfave/negroni"
	nprom "github.com/zbindenren/negroni-prometheus"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	githubUsername := flag.String("github-username", "", "GitHub username")

	pivotalKey := flag.String("pivotal-api-key", "", "API key for Pivotal Tracker")
	rawPivotalOwners := flag.String("pivotal-owners", "", "Comma separated list of Pivotal Tracker owner IDs")
	rawPivotalProjects := flag.String("pivotal-projects", "", "Comma separated list of Pivotal Tracker project IDs")

	trelloKey := flag.String("trello-api-key", "", "API key for Trello")
	trelloToken := flag.String("trello-api-token", "", "API token for Trello")
	rawTrelloUsernames := flag.String("trello-usernames", "", "Comma separated list of Trello usernames")

	flag.Parse()

	pivotalOwners := []string{}
	for _, ownerID := range strings.Split(*rawPivotalOwners, ",") {
		if owner := strings.TrimSpace(ownerID); owner != "" {
			pivotalOwners = append(pivotalOwners, owner)
		}
	}

	pivotalProjects := []string{}
	for _, projectID := range strings.Split(*rawPivotalProjects, ",") {
		if proj := strings.TrimSpace(projectID); proj != "" {
			pivotalProjects = append(pivotalProjects, proj)
		}
	}

	trelloUsernames := []string{}
	for _, username := range strings.Split(*rawTrelloUsernames, ",") {
		if user := strings.TrimSpace(username); username != "" {
			trelloUsernames = append(trelloUsernames, user)
		}
	}

	var (
		todoSyncTime = time.Now()
		currentTodos []todo.Todo
		lock         sync.RWMutex
	)

	go func() {
		for {
			todos := []todo.Todo{}

			if *githubUsername != "" {
				ghAssigneeTodos, err := github.FetchGitHubAssigneeTodos(*githubUsername)
				if err != nil {
					logger.Fatal(err)
				}
				todos = append(todos, ghAssigneeTodos...)

				ghAuthorTodos, err := github.FetchGitHubAuthorTodos(*githubUsername)
				if err != nil {
					logger.Fatal(err)
				}
				todos = append(todos, ghAuthorTodos...)
			}

			if len(pivotalProjects) > 0 {
				pivotalTodos, err := pivotal.FetchPivotalTodos(
					*pivotalKey,
					pivotalOwners,
					pivotalProjects,
				)
				if err != nil {
					logger.Fatal(err)
				}
				todos = append(todos, pivotalTodos...)
			}

			if *trelloKey != "" {
				trelloTodos, err := trello.FetchTrelloTodos(
					*trelloKey,
					*trelloToken,
					trelloUsernames,
				)
				if err != nil {
					logger.Fatal(err)
				}
				todos = append(todos, trelloTodos...)
			}

			lock.Lock()
			todoSyncTime = time.Now()
			currentTodos = todos
			lock.Unlock()

			numTodo := float64(0)
			numDoing := float64(0)
			numDone := float64(0)
			for _, todo := range todos {
				if !todo.Started() && !todo.Finished() {
					numTodo++
				}
				if todo.Started() && !todo.Finished() {
					numDoing++
				}
				if todo.Finished() {
					numDoing++
				}
			}
			todoCountMetric.With(prom.Labels{"state": "todo"}).Set(numTodo)
			todoCountMetric.With(prom.Labels{"state": "doing"}).Set(numDoing)
			todoCountMetric.With(prom.Labels{"state": "done"}).Set(numDone)

			time.Sleep(3 * time.Minute)
		}
	}()

	renderer := render.New(render.Options{
		Directory: "templates",
	})

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "healthy")
	})

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		lock.RLock()
		defer lock.RUnlock()

		renderer.HTML(
			w, http.StatusOK,
			"todos",
			struct {
				SecondsAgo int64
				Todos      []todo.Todo
			}{
				SecondsAgo: int64(time.Now().Sub(todoSyncTime).Seconds()),
				Todos:      currentTodos,
			},
		)
	})

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(nlogrus.NewMiddlewareFromLogger(logger, "web"))
	n.Use(gzip.Gzip(gzip.DefaultCompression))
	n.Use(negroni.HandlerFunc(nsecure.New().HandlerFuncWithNext))
	n.Use(nprom.NewMiddleware("todo-aggregator"))
	n.Use(negroni.NewStatic(http.Dir("public")))
	n.UseHandler(mux)

	ctx, cancel := signalcontext.On(syscall.SIGTERM)
	defer cancel()

	server := &http.Server{Addr: ":8080", Handler: n}

	go func() {
		server.ListenAndServe()
	}()
	logger.Println("server is listening")

	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	os.Exit(0)
}
