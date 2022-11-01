package messages

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/types"
)

func (s *Model) editExpenseAfterEditing(ctx context.Context, expense *types.Expense, userID int64, expenseID int) error {
	userCurrency, err := s.usersDB.GetUserCurrency(ctx, userID)

	if err != nil {
		return errors.Wrap(err, "cannot GetUserCurrency")
	}

	rate, err := s.getCurrentCurrencyRate(ctx, userCurrency)

	if err != nil {
		return errors.Wrap(err, "cannot getCurrentCurrencyRate")
	}

	message := expense.ToString(&types.UserModel{
		Currency:     string(userCurrency),
		CurrencyRate: rate,
	})
	return s.tgClient.EditExpenseMessage(message, userID, expenseID)
}

func (s *Model) sumEntered(ctx context.Context, msg *Message, userState types.CurrentState) error {
	// Try to convert to a number.
	sum, err := strconv.ParseFloat(msg.Text, 64)
	if err != nil {
		return errors.Wrap(err, "cannot ParseFloat")
	}

	expense, err := s.expensesDB.GetExpense(ctx, msg.UserID, userState.ExpenseID)

	if err != nil {
		return errors.Wrap(err, "cannot GetExpense")
	}

	if expense == nil {
		// Then we work with this expense for the first time.
		expense = types.NewExpense()
		expense.ExpenseID = userState.ExpenseID
		err := s.initializeExpense(ctx, msg, expense)
		if err != nil {
			return errors.Wrap(err, "cannot InitializeExpense")
		}
	}

	currency, err := s.getUserCurrency(ctx, msg.UserID)

	if err != nil {
		return errors.Wrap(err, "cannot GetUserCurrency")
	}

	rate, err := s.getCurrentCurrencyRate(ctx, currency)

	if err != nil {
		return errors.Wrap(err, "cannot getCurrentCurrencyRate")
	}

	// Change value of the expense.
	int_sum := int(sum * float64(rate))
	expense.Sum = int_sum
	err = s.expensesDB.WriteSum(ctx, int_sum, msg.UserID, userState.ExpenseID)

	if err != nil {
		return errors.Wrap(err, "cannot WriteSum")
	}

	err = s.checkCategorySum(ctx, msg.UserID, expense)
	if err != nil {
		return errors.Wrap(err, "cannot checkCategorySum")
	}

	err = s.tgClient.DeleteMessage(msg.UserID, msg.MessageID)
	if err != nil {
		return errors.Wrap(err, "cannot DeleteMessage")
	}

	// Changing state to the waiting one.
	err = s.usersDB.ToWaitState(ctx, msg.UserID)
	if err != nil {
		return errors.Wrap(err, "cannot ToWaitState")
	}

	// Edit message.
	return s.editExpenseAfterEditing(ctx, expense, msg.UserID, userState.ExpenseID)
}

func (s *Model) checkCategorySum(ctx context.Context, userID int64, expense *types.Expense) error {
	limit, ok, err := s.limitsDB.GetLimit(ctx, userID, int(expense.Date.Month()))

	if err != nil {
		return errors.Wrap(err, "cannot GetLimit")
	}

	if !ok {
		limit = defaultLimit
		err := s.limitsDB.SetLimit(ctx, userID, int(expense.Date.Month()), limit)
		if err != nil {
			return errors.Wrap(err, "cannot SetLimit")
		}
	}

	currentMonthExpenses, err := s.expensesDB.GetMonthReport(ctx, userID, expense.Date)

	if err != nil {
		return errors.Wrap(err, "cannot GetMonthReport")
	}

	if currentMonthExpenses >= limit {
		return s.tgClient.SendMessage(limitExceededMsg, userID)
	}

	return nil
}

func (s *Model) categoryEntered(ctx context.Context, msg *Message, userState types.CurrentState) error {
	// Write to expense.
	expense, err := s.expensesDB.GetExpense(ctx, msg.UserID, userState.ExpenseID)

	if err != nil {
		return errors.Wrap(err, "cannot GetExpense")
	}

	if expense == nil {
		// Then we work with this expense for the first time.
		expense = types.NewExpense()
		expense.ExpenseID = userState.ExpenseID
		err := s.initializeExpense(ctx, msg, expense)
		if err != nil {
			return errors.Wrap(err, "cannot initializeExpense")
		}
	}

	expense.Category = msg.Text
	err = s.expensesDB.WriteCategory(ctx, msg.Text, msg.UserID, userState.ExpenseID)

	if err != nil {
		return errors.Wrap(err, "cannot WriteCategory")
	}

	err = s.tgClient.DeleteMessage(msg.UserID, msg.MessageID)

	if err != nil {
		return errors.Wrap(err, "cannot DeleteMessage")
	}

	err = s.usersDB.ToWaitState(ctx, msg.UserID)
	if err != nil {
		return errors.Wrap(err, "cannot ToWaitState")
	}
	// Edit message.
	return s.editExpenseAfterEditing(ctx, expense, msg.UserID, userState.ExpenseID)
}

func (s *Model) dateEntered(ctx context.Context, msg *Message, userState types.CurrentState) error {
	// Try to convert to a date.
	date, err := time.Parse("2006-01-02", msg.Text)
	if err != nil {
		return errors.Wrap(err, "cannot Parse")
	}

	// Write to expense.
	expense, err := s.expensesDB.GetExpense(ctx, msg.UserID, userState.ExpenseID)

	if err != nil {
		return errors.Wrap(err, "cannot GetExpense")
	}

	if expense == nil {
		// Then we work with this expense for the first time.
		expense = types.NewExpense()
		expense.ExpenseID = userState.ExpenseID
		err := s.initializeExpense(ctx, msg, expense)
		if err != nil {
			return errors.Wrap(err, "cannot initializeExpense")
		}
	}

	expense.Date = date
	err = s.expensesDB.WriteDate(ctx, date, msg.UserID, userState.ExpenseID)

	if err != nil {
		return errors.Wrap(err, "cannot WriteDate")
	}

	err = s.tgClient.DeleteMessage(msg.UserID, msg.MessageID)

	if err != nil {
		return errors.Wrap(err, "cannot DeleteMessage")
	}

	err = s.usersDB.ToWaitState(ctx, msg.UserID)
	if err != nil {
		return errors.Wrap(err, "cannot ToWaitState")
	}
	// Edit message.
	return s.editExpenseAfterEditing(ctx, expense, msg.UserID, userState.ExpenseID)
}

func (s *Model) limitEntered(ctx context.Context, msg *Message) error {
	numbers := strings.Split(msg.Text, " ")

	if len(numbers) != 2 {
		return errors.New("incorrect input")
	}

	month, err := strconv.ParseInt(numbers[0], 10, 64)

	if err != nil {
		return errors.Wrap(err, "cannot ParseInt")
	}

	limit, err := strconv.ParseInt(numbers[1], 10, 64)

	if err != nil {
		return errors.Wrap(err, "cannot ParseInt")
	}

	if month < 1 || month > 12 {
		return errors.New("month is incorrect")
	}

	currency, err := s.getUserCurrency(ctx, msg.UserID)

	if err != nil {
		return errors.Wrap(err, "cannot getUserCurrency")
	}

	rate, err := s.getCurrentCurrencyRate(ctx, currency)

	if err != nil {
		return errors.Wrap(err, "cannot getCurrentCurrencyRate")
	}

	err = s.limitsDB.SetLimit(ctx, msg.UserID, int(month), int(limit)*rate)

	return errors.Wrap(err, "cannot SetLimit")
}

func (s *Model) initializeExpense(ctx context.Context, msg *Message, expense *types.Expense) error {
	err := s.expensesDB.WriteExpense(ctx, msg.UserID, expense)
	if err != nil {
		return errors.Wrap(err, "cannot WriteExpense")
	}

	return nil
}

func (s *Model) getCurrentCurrencyRate(ctx context.Context, c types.Currency) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rate, err := s.ratesDB.GetCurrencyRate(ctx, c, time.Now())

	if err != nil {
		if err == types.ErrNoCurrencyRate {
			err := s.currencyUpdater.UpdateCurrencyRate(ctx)

			if err != nil {
				return 0, errors.Wrap(err, "cannot UpdateCurrencyRate")
			}

			return s.getCurrentCurrencyRate(ctx, c)
		} else {
			return 0, errors.Wrap(err, "cannot GetCurrencyRate")
		}
	}

	return rate, err
}

func (s *Model) getUserCurrency(ctx context.Context, userID int64) (types.Currency, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	currency, err := s.usersDB.GetUserCurrency(ctx, userID)

	if err != nil {
		if err == types.ErrNoCurrency {
			err := s.usersDB.SetUserCurrency(ctx, userID, types.RUB)

			if err != nil {
				return "", errors.Wrap(err, "cannot SetUserCurrency")
			}

			_, err = s.getUserCurrency(ctx, userID)
			if err != nil {
				return "", errors.Wrap(err, "cannot getUserCurrency")
			}
		} else {
			return "", errors.Wrap(err, "cannot GetUserCurrency")
		}
	}

	return currency, nil
}

func (s *Model) setLimit(ctx context.Context, msg *Message) error {
	err := s.usersDB.SetCurrentState(ctx, msg.UserID, types.CurrentState{
		ExpenseID: msg.MessageID,
		State:     types.EditingLimit,
	})

	if err != nil {
		return errors.Wrap(err, "cannot SetCurrentState")
	}

	return s.tgClient.SendMessage(setLimitMsg, msg.UserID)
}
