package messages

import (
	"context"
	"log"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/types"
)

type messageSender interface {
	SendMessage(text string, userID int64) error
	CreateExpense(text string, userID int64) error
	EditExpenseMessage(text string, userID int64, messageID int) error
	DeleteMessage(userID int64, messageID int) error
	GetReport(text string, userID int64) error
	ChangeCurrency(text string, userID int64) error
}

type expensesDB interface {
	WriteExpense(ctx context.Context, fromID int64, expense *types.Expense) error
	GetExpense(ctx context.Context, userID int64, expenseID int) (*types.Expense, error)
	WriteSum(ctx context.Context, sum int, userID int64, expenseID int) error
	WriteCategory(ctx context.Context, category string, userID int64, expenseID int) error
	WriteDate(ctx context.Context, date time.Time, userID int64, expenseID int) error
	GetMonthReport(ctx context.Context, userID int64, date time.Time) (int, error)
}

type usersDB interface {
	ToWaitState(ctx context.Context, userID int64) error
	SetUserCurrency(ctx context.Context, userID int64, currency types.Currency) error
	GetUserCurrency(ctx context.Context, userID int64) (types.Currency, error)
	SetCurrentState(ctx context.Context, userID int64, state types.CurrentState) error
	GetCurrentState(ctx context.Context, userID int64) (*types.UserStateType, bool)
}

type ratesDB interface {
	GetCurrencyRate(ctx context.Context, currency types.Currency, date time.Time) (int, error)
}

type limitsDB interface {
	GetLimit(ctx context.Context, userID int64, monthNo int) (int, bool, error)
	SetLimit(ctx context.Context, userID int64, monthNo, limit int) error
}

type currencyUpdater interface {
	UpdateCurrencyRate(ctx context.Context) error
}

type Model struct {
	tgClient        messageSender
	expensesDB      expensesDB
	usersDB         usersDB
	ratesDB         ratesDB
	limitsDB        limitsDB
	currencyUpdater currencyUpdater
}

func New(tgClient messageSender, expensesDB expensesDB, usersDB usersDB, ratesDB ratesDB, limitsDB limitsDB, updater currencyUpdater) *Model {
	return &Model{
		tgClient:        tgClient,
		expensesDB:      expensesDB,
		usersDB:         usersDB,
		ratesDB:         ratesDB,
		limitsDB:        limitsDB,
		currencyUpdater: updater,
	}
}

type Message struct {
	Text      string
	UserID    int64
	MessageID int
}

const (
	defaultLimit = 1000000 // in kopecks

	getReportMsg      = "Запросить отчет за:"
	changeCurrencyMsg = "Выберите валюту"
	setLimitMsg       = "Введите два числа через пробел - номер месяца (от 1 до 12) и новый лимит на данный месяц"
	incorrectLimitMsg = "Ошибка при обновлении лимита. Проверьте корректность введенных данных"
	limitExceededMsg  = "Внимание, лимит трат в этом месяце исчерпан!"
)

func (s *Model) newExpenseMsg(ctx context.Context, userID int64) string {
	currency, err := s.getUserCurrency(ctx, userID)
	if err != nil {
		log.Println(err)
	}

	rate, err := s.getCurrentCurrencyRate(ctx, currency)
	log.Println("rate ", rate)
	if err != nil {
		log.Println(err)
	}

	newExpenseMsg := types.Expense{
		Sum:      0,
		Category: "Новая категория",
		Date:     time.Now(),
	}

	return newExpenseMsg.ToString(&types.UserModel{
		Currency:     string(currency),
		CurrencyRate: rate,
	})
}

func (s *Model) IncomingMessage(ctx context.Context, msg *Message) error {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"IncomingMessage",
	)
	span.SetTag("message", msg.Text)
	defer span.Finish()

	// Trying to recognize the command.
	switch msg.Text {
	case "/start":
		return s.tgClient.SendMessage("hello", msg.UserID)
	case "/new_expense":
		return s.tgClient.CreateExpense(s.newExpenseMsg(ctx, msg.UserID), msg.UserID)
	case "/change_currency":
		return s.tgClient.ChangeCurrency(changeCurrencyMsg, msg.UserID)
	case "/get_report":
		return s.tgClient.GetReport(getReportMsg, msg.UserID)
	case "/set_limit":
		return s.setLimit(ctx, msg)
	}

	// It is not a known command - maybe it is message to change the state.
	if userState, ok := s.usersDB.GetCurrentState(ctx, msg.UserID); ok {
		switch userState.CurrentState.State {
		case types.EditingSum:
			return s.sumEntered(ctx, msg, userState.CurrentState)
		case types.EditingCategory:
			return s.categoryEntered(ctx, msg, userState.CurrentState)
		case types.EditingDate:
			return s.dateEntered(ctx, msg, userState.CurrentState)
		case types.EditingLimit:
			err := s.limitEntered(ctx, msg)

			if err != nil {
				err = s.tgClient.SendMessage(incorrectLimitMsg, msg.UserID)
				if err != nil {
					return errors.Wrap(err, "cannot SendMessage")
				}
			}

			return errors.Wrap(err, "cannot limitEntered")
		case types.WaitState:
		}
	}

	return s.tgClient.SendMessage("не знаю эту команду", msg.UserID)
}
