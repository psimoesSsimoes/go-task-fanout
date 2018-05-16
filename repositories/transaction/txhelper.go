package transaction

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

// Handler helper alias
type Handler func(context.Context, *sql.Tx) error

// InTransaction helper function to process functions inside a database transaction
func InTransaction(ctx context.Context, pool *sql.DB, handler Handler) (err error) {
	var tx *sql.Tx

	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%+v", p)
		}

		if tx != nil {
			if err != nil {
				if rErr := tx.Rollback(); rErr != nil {
					err = errors.Wrap(rErr, err.Error())
				}
			} else {
				err = tx.Commit()
			}
		}
	}()

	tx, err = pool.BeginTx(ctx, nil)
	if err != nil {
		return
	}

	if err = handler(ctx, tx); err != nil {
		return
	}

	return
}
