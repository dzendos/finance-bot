package database

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/types"
	"golang.org/x/net/context"
)

type CacheModel interface {
	Ping() error
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Exists(key string) (bool, error)
	Delete(key string) error
}

type expensesDB struct {
	db    *sql.DB
	cache CacheModel
}

func NewExpensesDB(db *sql.DB, cache CacheModel) *expensesDB {
	return &expensesDB{
		db:    db,
		cache: cache,
	}
}

func (db *expensesDB) WriteExpense(ctx context.Context, fromID int64, expense *types.Expense) error {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"WriteExpense",
	)
	defer span.Finish()

	const query = `
		INSERT INTO expenses(
			tg_user_id,
			expense_id,
			expense_sum,
			category,
			created_at
		) values (
			$1, $2, $3, $4, $5
		);
	`

	_, err := db.db.ExecContext(ctx, query,
		fromID,
		expense.ExpenseID,
		expense.Sum,
		expense.Category,
		expense.Date,
	)

	if err != nil {
		return errors.Wrap(err, "cannot ExecContent")
	}

	return nil
}

func (db *expensesDB) GetExpense(ctx context.Context, userID int64, expenseID int) (*types.Expense, error) {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"GetExpense",
	)
	defer span.Finish()

	const query = `
		SELECT 
			expense_sum,
			category,
			created_at
		FROM expenses 
		WHERE 
			tg_user_id = $1 AND expense_id = $2
	`

	expense := types.NewExpense()

	err := db.db.QueryRowContext(ctx, query,
		userID,
		expenseID,
	).Scan(&expense.Sum, &expense.Category, &expense.Date)

	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "cannot Scan")
		}

		// In another case we haven't just found the needed row.
		return nil, nil
	}

	return expense, nil
}

func (db *expensesDB) DeleteExpense(ctx context.Context, userID int64, expenseID int) error {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"DeleteExpense",
	)
	defer span.Finish()

	const query = `
		DELETE FROM
			expenses
		WHERE
			tg_user_id = $1 AND
			expense_id = $2
	`
	_, err := db.db.ExecContext(ctx, query,
		userID,
		expenseID,
	)

	if err != nil {
		return errors.Wrap(err, "cannot ExecContent")
	}

	return nil
}

func (db *expensesDB) EditNewExpense(ctx context.Context, userID int64, expenseID int, expense *types.Expense) error {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"EditNewExpense",
	)
	defer span.Finish()

	const query = `
		UPDATE 
			expenses
		SET
			expense_sum = $1,
			category = $2,
			created_at = $3
		WHERE
			tg_user_id = $4 AND
			expense_id = $5
	`

	_, err := db.db.ExecContext(ctx, query,
		expense.Sum,
		expense.Category,
		expense.Date,
		userID,
		expenseID,
	)

	if err != nil {
		return errors.Wrap(err, "cannot ExecContent")
	}

	return nil
}

func (db *expensesDB) GetReport(ctx context.Context, fromID int64, dateBegin time.Time, dateEnd time.Time) (map[string]int, error) {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"GetReport",
	)
	defer span.Finish()

	report, err := db.getCachedReport(fromID, dateBegin, dateEnd)
	if err != nil {
		return nil, errors.Wrap(err, "cannot db.getCachedExpense")
	}

	if report != nil {
		return report, nil
	}

	const query = `
		SELECT 
			SUM(expense_sum),
			category
		FROM expenses
		WHERE 
			tg_user_id = $1 AND
			(created_at BETWEEN $2 AND $3)
		GROUP BY
			category
	`

	rows, err := db.db.QueryContext(ctx, query,
		fromID,
		dateBegin,
		dateEnd,
	)

	if err != nil {
		return nil, errors.Wrap(err, "cannot QueryContext")
	}
	defer rows.Close()

	report = make(map[string]int)

	for rows.Next() {
		var name string
		var sum int

		if err := rows.Scan(&sum, &name); err != nil {
			return nil, errors.Wrap(err, "cannot Scan")
		}

		report[name] = sum
	}

	err = db.cacheReport(fromID, dateBegin, dateEnd, report)
	if err != nil {
		return nil, errors.Wrap(err, "cannot cacheReport")
	}

	return report, nil
}

func (db *expensesDB) WriteSum(ctx context.Context, sum int, userID int64, expenseID int) error {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"WriteSum",
	)
	defer span.Finish()

	const query = `
		UPDATE 
			expenses
		SET
			expense_sum = $1
		WHERE
			tg_user_id = $2 AND
			expense_id = $3
	`

	_, err := db.db.ExecContext(ctx, query,
		sum,
		userID,
		expenseID,
	)

	if err != nil {
		return errors.Wrap(err, "cannot ExecContent")
	}

	return nil
}

func (db *expensesDB) WriteCategory(ctx context.Context, category string, userID int64, expenseID int) error {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"WriteCategory",
	)
	defer span.Finish()

	const query = `
		UPDATE 
			expenses
		SET
			category = $1
		WHERE
			tg_user_id = $2 AND
			expense_id = $3
	`

	_, err := db.db.ExecContext(ctx, query,
		category,
		userID,
		expenseID,
	)

	if err != nil {
		return errors.Wrap(err, "cannot ExecContent")
	}

	return nil
}

func (db *expensesDB) WriteDate(ctx context.Context, date time.Time, userID int64, expenseID int) error {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"WriteDate",
	)
	defer span.Finish()

	const query = `
		UPDATE 
			expenses
		SET
			created_at = $1
		WHERE
			tg_user_id = $2 AND
			expense_id = $3
	`

	_, err := db.db.ExecContext(ctx, query,
		date,
		userID,
		expenseID,
	)

	if err != nil {
		return errors.Wrap(err, "cannot ExecContent")
	}

	return nil
}

func (db *expensesDB) GetMonthReport(ctx context.Context, userID int64, date time.Time) (int, error) {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"GetMonthReport",
	)
	defer span.Finish()

	const query = `
		SELECT 
			SUM(expense_sum)
		FROM expenses
		WHERE 
			tg_user_id = $1 AND
			EXTRACT(YEAR FROM created_at) = EXTRACT(YEAR FROM DATE($2)) AND
			EXTRACT(MONTH FROM created_at) = EXTRACT(MONTH FROM DATE($2))			
	`

	var sum int
	err := db.db.QueryRowContext(ctx, query,
		userID,
		date,
	).Scan(&sum)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}

		return 0, errors.Wrap(err, "cannot QueryRowContext")
	}

	return sum, nil
}

func (db *expensesDB) cacheReport(userID int64, dateBegin time.Time, dateEnd time.Time, report map[string]int) error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(report)

	if err != nil {
		return errors.Wrap(err, "cannot Encode")
	}

	err = db.cache.Set(getReportKeyString(userID, dateBegin, dateEnd), buffer.Bytes())

	if err != nil {
		return errors.Wrap(err, "cannot cache.Set")
	}

	return nil
}

func (db *expensesDB) getCachedReport(userID int64, dateBegin time.Time, dateEnd time.Time) (map[string]int, error) {
	ok, err := db.cache.Exists(getReportKeyString(userID, dateBegin, dateEnd))
	if err != nil {
		return nil, errors.Wrap(err, "cannot cache.Exists")
	}

	if ok {
		obj, err := db.cache.Get(getReportKeyString(userID, dateBegin, dateEnd))
		if err != nil {
			return nil, errors.Wrap(err, "cannot cache.Get")
		}

		var buffer bytes.Buffer
		_, err = buffer.Write(obj)
		if err != nil {
			return nil, errors.Wrap(err, "cannot buffer.Write")
		}

		dec := gob.NewDecoder(&buffer)

		var report map[string]int
		err = dec.Decode(&report)

		if err != nil {
			return nil, errors.Wrap(err, "cannot dec.Decode")
		}

		return report, nil
	}

	return nil, nil
}

func getReportKeyString(userID int64, dateBegin time.Time, dateEnd time.Time) string {
	return fmt.Sprintf("%d_{%s}_{%s}", userID, dateBegin.Format("2006-01-02"), dateEnd.Format("2006-01-02"))
}
