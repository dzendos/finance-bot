package database

import (
	"context"
	"database/sql"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/types"
)

type usersDB struct {
	db *sql.DB
}

func NewUsersDB(db *sql.DB) *usersDB {
	return &usersDB{
		db: db,
	}
}

// Actions with users table.

func (db *usersDB) SetCurrentState(ctx context.Context, userID int64, state types.CurrentState) error {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"SetCurrentState",
	)
	defer span.Finish()

	const query = `
		INSERT INTO users(
			tg_user_id,
			expense_id,
			current_state
		) VALUES (
			$1, $2, $3
		)
		ON CONFLICT (tg_user_id) DO UPDATE
		SET 
			expense_id = $2,
			current_state = $3
	`

	_, err := db.db.ExecContext(ctx, query,
		userID,
		state.ExpenseID,
		state.State,
	)

	if err != nil {
		return errors.Wrap(err, "cannot ExecContent")
	}

	return nil
}

func (db *usersDB) GetCurrentState(ctx context.Context, userID int64) (*types.UserStateType, bool) {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"GetCurrentState",
	)
	defer span.Finish()

	const query = `
		SELECT
			expense_id,
			current_state,
			current_currency
		FROM
			users
		WHERE
			tg_user_id = $1
	`

	var userState types.UserStateType

	err := db.db.QueryRowContext(ctx, query,
		userID,
	).Scan(&userState.CurrentState.ExpenseID, &userState.CurrentState.State, &userState.Currency)

	if err != nil {
		return nil, false
	}

	return &userState, true
}

func (db *usersDB) SetUserCurrency(ctx context.Context, userID int64, currency types.Currency) error {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"SetUserCurrency",
	)
	defer span.Finish()

	const query = `
	INSERT INTO users(
		tg_user_id,
		current_currency
	) VALUES (
		$1, $2
	)
	ON CONFLICT(tg_user_id)
	DO UPDATE 
		SET
			current_currency = $2
	`

	_, err := db.db.ExecContext(ctx, query,
		userID,
		currency,
	)

	if err != nil {
		return errors.Wrap(err, "cannot ExecContent")
	}

	return nil
}

func (db *usersDB) GetUserCurrency(ctx context.Context, userID int64) (types.Currency, error) {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"GetUserCurrency",
	)
	defer span.Finish()

	const query = `
		SELECT 
			current_currency
		FROM
			users
		WHERE
			tg_user_id = $1
	`

	var currency types.Currency

	err := db.db.QueryRowContext(ctx, query,
		userID,
	).Scan(&currency)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", types.ErrNoCurrency
		}

		return "", errors.Wrap(err, "cannot QueryRowContext")
	}

	return currency, nil
}

func (db *usersDB) ToWaitState(ctx context.Context, userID int64) error {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"ToWaitState",
	)
	defer span.Finish()

	const query = `
		INSERT INTO users(
			tg_user_id,
			current_state
		) VALUES (
			$1, $2
		)
		ON CONFLICT(tg_user_id)
		DO UPDATE 
			SET
			current_state = $2
	`

	_, err := db.db.ExecContext(ctx, query,
		userID,
		types.WaitState,
	)

	if err != nil {
		return errors.Wrap(err, "cannot ExecContent")
	}

	return nil
}
