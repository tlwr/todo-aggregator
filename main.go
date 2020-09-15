package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/tlwr/todo-aggregator/agg/pivotal"
	"github.com/tlwr/todo-aggregator/agg/trello"
	"github.com/tlwr/todo-aggregator/todo"

	nlogrus "github.com/meatballhat/negroni-logrus"
	"github.com/phyber/negroni-gzip/gzip"
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

	todos := []todo.Todo{}

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

	renderer := render.New(render.Options{
		Directory: "templates",
	})

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "healthy")
	})

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		renderer.HTML(w, http.StatusOK, "todos", todos)
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
