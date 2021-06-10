package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/darkside1809/gosql/cmd/app"
	"github.com/darkside1809/gosql/pkg/customers"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/gorilla/mux"
	"go.uber.org/dig"
)

func main() {
	host := "0.0.0.0"
	port := "9999"

	dsn := "postgres://app:pass@localhost:5432/db"

	if err := execute(host, port, dsn); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(server, port, dsn string) (err error) {
	deps := []interface{}{
		app.NewServer,
		mux.NewRouter,
		customers.NewService,
		func() (*pgxpool.Pool, error) {
			connCtx, _ := context.WithTimeout(context.Background(), time.Second * 5)
			return pgxpool.Connect(connCtx, dsn)
		},
		func(serverHandler *app.Server) *http.Server {
			return &http.Server{
				Addr:    net.JoinHostPort(server, port),
				Handler: serverHandler,
			}
		},
	}

	container := dig.New()
	for _, dep := range deps {
		err = container.Provide(dep)
		if err != nil {
			return err
		}
	}

	err = container.Invoke(func(server *app.Server) { 
		server.Init() 
	})
	if err != nil {
		return err
	}

	return container.Invoke(func(s *http.Server) error { 
		return s.ListenAndServe() 
	})
}