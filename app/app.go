package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

type App struct {
	router http.Handler
	db     *pgx.Conn
}

func New() *App {
	conn, err := pgx.Connect(context.Background(), "postgresql://admin:admin@localhost:5432")
	if err != nil {
		log.Fatal(err)
	}

	app := &App{
		db: conn,
	}

	app.loadRouter()

	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":8080",
		Handler: a.router,
	}

	err := a.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}

	err = a.createTables()
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	defer func() {
		if err := a.db.Close(context.Background()); err != nil {
			fmt.Println("failed to close db", err)
		}
	}()

	fmt.Println("Starting server")

	ch := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		return server.Shutdown(timeout)
	}
}

func (a *App) createTables() error {
	dropTable := `DROP TABLE IF EXISTS files`
	_, err := a.db.Exec(context.Background(), dropTable)
	if err != nil {
		return err
	}

	createTableQuery := `
		CREATE TABLE IF NOT EXISTS files (
				id SERIAL PRIMARY KEY,
				batch_id UUID NOT NULL,
				name TEXT NOT NULL,
				file BYTEA NOT NULL
		)
	`
	_, err = a.db.Exec(context.Background(), createTableQuery)
	if err != nil {
		return err
	}

	return nil
}