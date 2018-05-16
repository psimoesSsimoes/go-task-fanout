package main

import (
	"database/sql"
	"go-task-fanout/persistence/postgres"

	"fmt"

	"time"

	_ "github.com/lib/pq" // postgreSQL driver
	"gitlab.com/mandalore/go-app/app"
	logger "gitlab.com/vredens/go-logger"
)

func main() {
	pman := app.NewProcessManager()

	conn, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:7000/postgres?sslmode=disable")
	conn.SetMaxIdleConns(1)
	conn.SetMaxOpenConns(2)

	if err != nil {
		panic(err)
	}

	repository := postgres.NewTaskStorage(conn, "test")
	if err := repository.Init(); err != nil {
		panic(err)
	}

	/*data := &struct {
		OfferID uuid.UUID `json:"offer_id"`
	}{
		OfferID: uuid.NewV4(),
	}

	dispatcher := taskworker.NewDispatcher(repository)
	if err := dispatcher.Process("cmd", "123", data); err != nil {
		panic(err)
	}*/

	receiver := taskworker.NewReceiver(repository, "cmd",
		taskworker.WithWorkHandler(handler),
		taskworker.WithTaskAge(time.Duration(10*time.Second)),
		taskworker.WithLogger(app.Logger.Spawn(logger.WithFields(app.KV{"command": "cmd"}))),
	)

	pman.AddProcess("task-receiver", receiver)

	pman.Start()

	app.Logger.Info("service terminated, bye!")
}

func handler(task *taskworker.Task) error {
	fmt.Printf("olha passou aqui: %d:%s\n", task.ID, task.TaskID)

	return nil
}
