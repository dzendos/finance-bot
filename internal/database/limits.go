package database

import (
	"database/sql"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type LimitsDB struct {
	db *sql.DB
}

func NewLimitsDB(db *sql.DB) *LimitsDB {
	return &LimitsDB{
		db: db,
	}
}

func (db *LimitsDB) GetLimit(ctx context.Context, userID int64, monthNo int) (int, bool, error) {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"GetLimit",
	)
	defer span.Finish()

	const query = `
		SELECT
			user_limit
		FROM 
			limits
		WHERE
			tg_user_id = $1 AND
			month_no = $2
	`

	var limit int
	err := db.db.QueryRowContext(ctx, query,
		userID,
		monthNo,
	).Scan(&limit)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}

		return 0, false, errors.Wrap(err, "cannot Scan")
	}

	return limit, true, nil
}

func (db *LimitsDB) SetLimit(ctx context.Context, userID int64, monthNo, limit int) error {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"SetLimit",
	)
	defer span.Finish()

	_, ok, err := db.GetLimit(ctx, userID, monthNo)

	if err != nil {
		return errors.Wrap(err, "cannot GetLimit")
	}

	if ok {
		const query = `
			UPDATE  limits
			SET
				user_limit = $2
			WHERE
				tg_user_id = $1,
				month_no = $3
		`

		_, err := db.db.ExecContext(ctx, query,
			userID,
			limit,
			monthNo,
		)

		if err != nil {
			return errors.Wrap(err, "cannot ExecContent")
		}

		return nil
	}

	const query = `
		INSERT INTO limits(
			tg_user_id,
			user_limit,
			month_no
		) values (
			$1, $2, $3
		)
	`

	_, err = db.db.ExecContext(ctx, query,
		userID,
		limit,
		monthNo,
	)

	if err != nil {
		return errors.Wrap(err, "cannot ExecContent")
	}

	return nil
}
