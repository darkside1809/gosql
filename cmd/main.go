package main

import (
	"database/sql"
	"github.com/darkside1809/gosql/cmd/app"
	"github.com/darkside1809/gosql/pkg/customers"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"net"
	"net/http"
	"os"
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

func execute(host string, port string, dsn string) (err error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := db.Close(); cerr != nil {
			if err != nil {
				err = cerr
			}
		}
	}()
	
	mux := http.NewServeMux()
	customersSvc := customers.NewService(db)

	server := app.NewServer(mux, customersSvc)
	server.Init()
	srv := &http.Server{
		Addr: net.JoinHostPort(host, port),
		Handler: server,
	}
	return srv.ListenAndServe()
}