package infrastructure

import (
	"database/sql"
	"fmt"

	"gitlab.com/worten/dom/dom-integration-inbound/repositories/mysql"

	"time"
	//driver
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/psimoesSsimoes/go-task-fanout/interactors"
	"gitlab.com/worten/dom/dom-integration-inbound/configurations"
)

// DefaultInfrastructure holder type
type DefaultInfrastructure struct {
	pool *sql.DB
}

// NewDefaultInfrastructure factory method to create the default infrastructure method
func NewDefaultInfrastructure(c configurations.Config) (Infrastructure, error) {
	i := DefaultInfrastructure{}

	pool, err := initPool(c.GetDBConfigs())
	if err != nil {
		return nil, err
	}

	i.pool = pool

	return &i, nil
}

// TaskDispatcherRepository implements infrastucture
func (i *DefaultInfrastructure) TaskDispatcherRepository() interactors.TaskDispatcherRepository {
	r := mysql.NewTaskDispatcher(i.pool)

	return &r
}

// TaskRegisterRepository implements infrastucture
func (i *DefaultInfrastructure) TaskRegisterRepository() interactors.TaskRegisterRepository {
	r := mysql.NewTaskRegisterUpdater(i.pool)

	return &r
}

func initPool(dsn configurations.DSN) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?parseTime=true&timeout=5s",
		dsn.User, dsn.Pass, dsn.Host, dsn.Name,
	)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, errors.Wrap(err, "sql.Open error at initPool")
	}

	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(1 * time.Hour)

	return db, nil
}
