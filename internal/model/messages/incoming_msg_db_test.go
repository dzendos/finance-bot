//go:build integration
// +build integration

package messages

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/currency"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/database"
	mocks "gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/mocks/messages"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/types"
)

func newConfig() *config.Service {
	return &config.Service{
		Config: config.Config{
			CbrServiceUrl:               "https://cbr.ru/scripts/XML_daily.asp?date_req=",
			FrequencyCurrencyRateUpdate: 86400,
			Host:                        "localhost",
			Port:                        5432,
			User:                        "postgres",
			Password:                    "pass",
			TestDB:                      "test",
			SslMode:                     "disable",
		},
	}
}

func Test_OnSumState_ShouldEditExpenseMessage(t *testing.T) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	config := newConfig()

	ctrl := gomock.NewController(t)
	sender := mocks.NewMockMessageSender(ctrl)
	db, err := database.NewTestDB(config)
	assert.NoError(t, err)

	expensesDB := database.NewExpensesDB(db)
	usersDB := database.NewUsersDB(db)
	ratesDB := database.NewRatesDB(db)
	limitsDB := database.NewLimitsDB(db)
	updater := currency.NewCbrCurrencyUpdater(config, ratesDB)
	model := New(sender, expensesDB, usersDB, ratesDB, limitsDB, updater)

	for i := 0; i < 10; i++ {
		err := expensesDB.DeleteExpense(ctx, int64(i), 123)
		assert.NoError(t, err)

		err = usersDB.SetCurrentState(ctx, int64(i), types.CurrentState{
			ExpenseID: 123,
			State:     types.EditingSum,
		})
		assert.NoError(t, err)

		err = usersDB.SetUserCurrency(ctx, int64(i), types.RUB)
		assert.NoError(t, err)
	}

	for i := 0; i < 10; i++ {
		rand.Seed(time.Now().UnixNano())
		num1, num2 := rand.Int()%100, rand.Int()%100

		res := float64(num1) + float64(num2)/100.0

		resStr := fmt.Sprintf("%.3f", res)
		log.Println(res, res*100.0, int(res*100.0), resStr)
		exp := types.Expense{
			Sum:      int(res * 100.0),
			Category: "Новая категория",
			Date:     time.Now(),
		}

		message := exp.ToString(&types.UserModel{
			Currency:     string(types.RUB),
			CurrencyRate: 100,
		})
		sender.EXPECT().DeleteMessage(int64(i), 123)
		sender.EXPECT().EditExpenseMessage(message, int64(i), 123)

		err = model.IncomingMessage(ctx, &Message{
			Text:      resStr,
			UserID:    int64(i),
			MessageID: 123,
		})

		assert.NoError(t, err)
	}
}

func Test_OnSumState_DifferentCurrencies(t *testing.T) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	config := newConfig()

	ctrl := gomock.NewController(t)
	sender := mocks.NewMockMessageSender(ctrl)
	db, err := database.NewTestDB(config)
	assert.NoError(t, err)

	expensesDB := database.NewExpensesDB(db)
	usersDB := database.NewUsersDB(db)
	ratesDB := database.NewRatesDB(db)
	limitsDB := database.NewLimitsDB(db)
	updater := currency.NewCbrCurrencyUpdater(config, ratesDB)
	model := New(sender, expensesDB, usersDB, ratesDB, limitsDB, updater)

	for i := 0; i < 10; i++ {
		err = expensesDB.DeleteExpense(ctx, int64(i), 123)
		assert.NoError(t, err)

		err = usersDB.SetCurrentState(ctx, int64(i), types.CurrentState{
			ExpenseID: 123,
			State:     types.EditingSum,
		})
		assert.NoError(t, err)

		err = usersDB.SetUserCurrency(ctx, int64(i), types.RUB)
		assert.NoError(t, err)
	}

	err = usersDB.SetUserCurrency(ctx, int64(0), types.RUB)
	assert.NoError(t, err)
	err = usersDB.SetUserCurrency(ctx, int64(1), types.EUR)
	assert.NoError(t, err)
	err = usersDB.SetUserCurrency(ctx, int64(2), types.EUR)
	assert.NoError(t, err)
	err = usersDB.SetUserCurrency(ctx, int64(3), types.CNY)
	assert.NoError(t, err)
	err = usersDB.SetUserCurrency(ctx, int64(4), types.RUB)
	assert.NoError(t, err)
	err = usersDB.SetUserCurrency(ctx, int64(5), types.USD)
	assert.NoError(t, err)
	err = usersDB.SetUserCurrency(ctx, int64(6), types.RUB)
	assert.NoError(t, err)
	err = usersDB.SetUserCurrency(ctx, int64(7), types.USD)
	assert.NoError(t, err)
	err = usersDB.SetUserCurrency(ctx, int64(8), types.CNY)
	assert.NoError(t, err)
	err = usersDB.SetUserCurrency(ctx, int64(9), types.EUR)
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		rand.Seed(time.Now().UnixNano())
		num1, num2 := rand.Int()%100, rand.Int()%100

		currency, err := usersDB.GetUserCurrency(ctx, int64(i))
		assert.NoError(t, err)

		rate, err := model.getCurrentCurrencyRate(ctx, currency)
		assert.NoError(t, err)

		res := float64(num1) + float64(num2)/100.0

		resStr := fmt.Sprintf("%.3f", res)
		log.Println(res, res*100.0, int(res*100.0), resStr)
		exp := types.Expense{
			Sum:      int(res * float64(rate)),
			Category: "Новая категория",
			Date:     time.Now(),
		}

		message := exp.ToString(&types.UserModel{
			Currency:     string(currency),
			CurrencyRate: rate,
		})
		sender.EXPECT().DeleteMessage(int64(i), 123)
		sender.EXPECT().EditExpenseMessage(message, int64(i), 123)

		err = model.IncomingMessage(ctx, &Message{
			Text:      resStr,
			UserID:    int64(i),
			MessageID: 123,
		})

		assert.NoError(t, err)
	}
}

func Test_OnSumStateIncorrectInput_ShouldGiveError(t *testing.T) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	config := newConfig()

	ctrl := gomock.NewController(t)
	sender := mocks.NewMockMessageSender(ctrl)
	db, err := database.NewTestDB(config)
	assert.NoError(t, err)

	expensesDB := database.NewExpensesDB(db)
	usersDB := database.NewUsersDB(db)
	ratesDB := database.NewRatesDB(db)
	limitsDB := database.NewLimitsDB(db)
	updater := currency.NewCbrCurrencyUpdater(config, ratesDB)
	model := New(sender, expensesDB, usersDB, ratesDB, limitsDB, updater)

	err = expensesDB.DeleteExpense(ctx, int64(0), 123)
	assert.NoError(t, err)

	err = usersDB.SetCurrentState(ctx, int64(0), types.CurrentState{
		ExpenseID: 123,
		State:     types.EditingSum,
	})
	assert.NoError(t, err)

	err = usersDB.SetUserCurrency(ctx, int64(0), types.RUB)
	assert.NoError(t, err)

	err = model.IncomingMessage(ctx, &Message{
		Text:      "some non number text",
		UserID:    0,
		MessageID: 123,
	})
	assert.Error(t, err)
}

func Test_OnCategoryState_ShouldEditExpenseMessage(t *testing.T) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	config := newConfig()

	ctrl := gomock.NewController(t)
	sender := mocks.NewMockMessageSender(ctrl)
	db, err := database.NewTestDB(config)
	assert.NoError(t, err)

	expensesDB := database.NewExpensesDB(db)
	usersDB := database.NewUsersDB(db)
	ratesDB := database.NewRatesDB(db)
	limitsDB := database.NewLimitsDB(db)
	updater := currency.NewCbrCurrencyUpdater(config, ratesDB)
	model := New(sender, expensesDB, usersDB, ratesDB, limitsDB, updater)

	err = expensesDB.DeleteExpense(ctx, int64(0), 123)
	assert.NoError(t, err)

	err = usersDB.SetCurrentState(ctx, int64(0), types.CurrentState{
		ExpenseID: 123,
		State:     types.EditingCategory,
	})
	assert.NoError(t, err)

	err = usersDB.SetUserCurrency(ctx, int64(0), types.RUB)
	assert.NoError(t, err)

	exp := types.Expense{
		Sum:      0,
		Category: "some category",
		Date:     time.Now(),
	}

	currency, err := usersDB.GetUserCurrency(ctx, int64(0))
	assert.NoError(t, err)

	rate, err := model.getCurrentCurrencyRate(ctx, currency)
	assert.NoError(t, err)

	message := exp.ToString(&types.UserModel{
		Currency:     string(currency),
		CurrencyRate: rate,
	})
	sender.EXPECT().DeleteMessage(int64(0), 123)
	sender.EXPECT().EditExpenseMessage(message, int64(0), 123)

	err = model.IncomingMessage(ctx, &Message{
		Text:      "some category",
		UserID:    int64(0),
		MessageID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnDateState_ShouldEditExpenseMessage(t *testing.T) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	config := newConfig()

	ctrl := gomock.NewController(t)
	sender := mocks.NewMockMessageSender(ctrl)
	db, err := database.NewTestDB(config)
	assert.NoError(t, err)

	expensesDB := database.NewExpensesDB(db)
	usersDB := database.NewUsersDB(db)
	ratesDB := database.NewRatesDB(db)
	limitsDB := database.NewLimitsDB(db)
	updater := currency.NewCbrCurrencyUpdater(config, ratesDB)
	model := New(sender, expensesDB, usersDB, ratesDB, limitsDB, updater)

	for i := 0; i < 10; i++ {
		err = expensesDB.DeleteExpense(ctx, int64(i), 123)
		assert.NoError(t, err)

		err = usersDB.SetCurrentState(ctx, int64(i), types.CurrentState{
			ExpenseID: 123,
			State:     types.EditingDate,
		})
		assert.NoError(t, err)

		err = usersDB.SetUserCurrency(ctx, int64(i), types.RUB)
		assert.NoError(t, err)
	}

	for i := 0; i < 10; i++ {
		date := getRandomDate()
		exp := types.Expense{
			Sum:      0,
			Category: "Новая категория",
			Date:     date,
		}

		currency, err := usersDB.GetUserCurrency(ctx, int64(0))
		assert.NoError(t, err)

		rate, err := model.getCurrentCurrencyRate(ctx, currency)
		assert.NoError(t, err)

		message := exp.ToString(&types.UserModel{
			Currency:     string(currency),
			CurrencyRate: rate,
		})

		sender.EXPECT().DeleteMessage(int64(i), 123)
		sender.EXPECT().EditExpenseMessage(message, int64(i), 123)

		err = model.IncomingMessage(ctx, &Message{
			Text:      fmt.Sprintf("%.10s", date),
			UserID:    int64(i),
			MessageID: 123,
		})

		assert.NoError(t, err)
	}
}

func getRandomDate() time.Time {
	rand.Seed(time.Now().UnixNano())
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}
