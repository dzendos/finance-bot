package callbacks

import (
	"context"
	"errors"
	"time"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/types"
)

const (
	// createExpenseKeyboard
	ChangeExpenseSum      string = "ChangeExpenseSum"
	ChangeExpenseCategory string = "ChangeExpenseCategory"
	ChangeExpenseDate     string = "ChangeExpenseDate"
	ChangeExpenseDone     string = "ChangeExpenseDone"
	ChangeExpenseCancel   string = "ChangeExpenseCancel"

	// getReportKeyboard
	GetWeekReport  string = "GetWeekReport"
	GetMonthReport string = "GetMonthReport"
	GetYearReport  string = "GetYearReport"

	// changeCurrencyKeyboard
	USD string = "USD"
	CNY string = "CNY"
	EUR string = "EUR"
	RUB string = "RUB"
)

type callbackHandler interface {
	SendMessage(text string, userID int64) error
	EditMessage(text string, userID int64, messageID int) error
	ShowAlert(text string, messageID string) error
	DoneMessage(userID int64, messageID int) error
	CancelMessage(userID int64, messageID int) error
}

type expensesDB interface {
	WriteExpense(ctx context.Context, fromID int64, expense *types.Expense) error
	DeleteExpense(ctx context.Context, userID int64, expenseID int) error
	EditNewExpense(ctx context.Context, userID int64, expenseID int, expense *types.Expense) error
	GetReport(ctx context.Context, fromID int64, dateBegin time.Time, dateEnd time.Time) (map[string]int, error)
}

type usersDB interface {
	SetCurrentState(ctx context.Context, userID int64, state types.CurrentState) error
	GetCurrentState(ctx context.Context, userID int64) (*types.UserStateType, bool)
	ToWaitState(ctx context.Context, userID int64) error
	SetUserCurrency(ctx context.Context, userID int64, currency types.Currency) error
	GetUserCurrency(ctx context.Context, userID int64) (types.Currency, error)
}

type ratesDB interface {
	GetCurrencyRate(ctx context.Context, currency types.Currency, date time.Time) (int, error)
}

type Model struct {
	tgClient   callbackHandler
	expensesDB expensesDB
	usersDB    usersDB
	ratesDB    ratesDB
}

func New(tgClient callbackHandler, expensesDB expensesDB, usersDB usersDB, ratesDB ratesDB) *Model {
	return &Model{
		tgClient:   tgClient,
		expensesDB: expensesDB,
		usersDB:    usersDB,
		ratesDB:    ratesDB,
	}
}

type CallbackData struct {
	FromID     int64
	MessageID  int
	Data       string
	CallbackID string
}

func (s *Model) IncomingCallback(ctx context.Context, data *CallbackData) error {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"IncomingCallback",
	)
	span.SetTag("callback", data.Data)
	defer span.Finish()

	switch data.Data {
	case ChangeExpenseSum:
		return s.toWriteSumState(ctx, data)

	case ChangeExpenseCategory:
		return s.toWriteCategoryState(ctx, data)

	case ChangeExpenseDate:
		return s.toWriteDateState(ctx, data)

	case ChangeExpenseDone:
		return s.saveExpense(data)

	case ChangeExpenseCancel:
		return s.cancelExpense(ctx, data)

	case GetWeekReport:
		return s.getWeekReport(ctx, data)

	case GetMonthReport:
		return s.getMonthReport(ctx, data)

	case GetYearReport:
		return s.getYearReport(ctx, data)

	case USD, CNY, EUR, RUB:
		return s.changeCurrentCurrency(ctx, data)
	}

	return errors.New("Callback handler for data '" + data.Data + "' was not found.")
}
