package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/tlwr/todo-aggregator/agg/pivotal"
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

	todos := []todo.Todo{}

	pivotalTrackerProjects := []string{"1133984"}
	pivotalTodos, err := pivotal.FetchPivotalTodos(pivotalTrackerProjects)
	if err != nil {
		logger.Fatal(err)
	}
	todos = append(todos, pivotalTodos...)

	for _, todo := range todos {
		logger.Printf("pivotal: %s %s", todo.Name(), todo.URL())
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

	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	os.Exit(0)
}
