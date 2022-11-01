package messages

import (
	"context"
	"os"
	"os/signal"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/mocks/messages"
)

func Test_OnStartCommand_ShouldAnswerWithIntroMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := mocks.NewMockmessageSender(ctrl)
	expensesDB := mocks.NewMockexpensesDB(ctrl)
	usersDB := mocks.NewMockusersDB(ctrl)
	ratesDB := mocks.NewMockratesDB(ctrl)
	limitsDB := mocks.NewMocklimitsDB(ctrl)
	updater := mocks.NewMockcurrencyUpdater(ctrl)
	model := New(sender, expensesDB, usersDB, ratesDB, limitsDB, updater)

	sender.EXPECT().SendMessage("hello", int64(123))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	err := model.IncomingMessage(ctx, &Message{
		Text:   "/start",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnNewExpenseCommand_ShouldCreateNewExpense(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := mocks.NewMockmessageSender(ctrl)
	expensesDB := mocks.NewMockexpensesDB(ctrl)
	usersDB := mocks.NewMockusersDB(ctrl)
	ratesDB := mocks.NewMockratesDB(ctrl)
	limitsDB := mocks.NewMocklimitsDB(ctrl)
	updater := mocks.NewMockcurrencyUpdater(ctrl)
	model := New(sender, expensesDB, usersDB, ratesDB, limitsDB, updater)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	usersDB.EXPECT().GetUserCurrency(gomock.Any(), int64(123))
	ratesDB.EXPECT().GetCurrencyRate(gomock.Any(), gomock.Any(), gomock.Any())
	sender.EXPECT().CreateExpense(gomock.Any(), int64(123))

	defer cancel()
	err := model.IncomingMessage(ctx, &Message{
		Text:   "/new_expense",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnGetReportCommand_ShouldCreateReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := mocks.NewMockmessageSender(ctrl)
	expensesDB := mocks.NewMockexpensesDB(ctrl)
	usersDB := mocks.NewMockusersDB(ctrl)
	ratesDB := mocks.NewMockratesDB(ctrl)
	limitsDB := mocks.NewMocklimitsDB(ctrl)
	updater := mocks.NewMockcurrencyUpdater(ctrl)
	model := New(sender, expensesDB, usersDB, ratesDB, limitsDB, updater)

	sender.EXPECT().GetReport("Запросить отчет за:", int64(123))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	err := model.IncomingMessage(ctx, &Message{
		Text:   "/get_report",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnUnknownCommand_ShouldAnswerWithHelpMessage(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := mocks.NewMockmessageSender(ctrl)
	expensesDB := mocks.NewMockexpensesDB(ctrl)
	usersDB := mocks.NewMockusersDB(ctrl)
	ratesDB := mocks.NewMockratesDB(ctrl)
	limitsDB := mocks.NewMocklimitsDB(ctrl)
	updater := mocks.NewMockcurrencyUpdater(ctrl)
	model := New(sender, expensesDB, usersDB, ratesDB, limitsDB, updater)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	usersDB.EXPECT().GetCurrentState(gomock.Any(), int64(123))
	sender.EXPECT().SendMessage("не знаю эту команду", int64(123))

	err := model.IncomingMessage(ctx, &Message{
		Text:   "some text",
		UserID: 123,
	})

	assert.NoError(t, err)
}
