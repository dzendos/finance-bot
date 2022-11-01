package main

import (
	"context"
	"os"
	"os/signal"

	"gitlab.ozon.dev/e.gerasimov/telegram-bot/cmd/logging"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/cmd/metrics"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/cmd/tracing"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/clients/tg"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/config"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/currency"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/database"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/model/callbacks"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/model/messages"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/redis"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/worker"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	logger := logging.InitLogger()
	tracing.InitTracing("actions_handler", logger)

	logger.Info("initializing config")
	config, err := config.New()
	if err != nil {
		logger.Fatal("config init failed:", zap.Error(err))
	}

	logger.Info("initializing database")
	db, err := database.New(config)
	if err != nil {
		logger.Fatal("database init failed", zap.Error(err))
	}

	logger.Info("initializing cache")
	cache, err := redis.New(config)
	if err != nil {
		logger.Fatal("cache init failed", zap.Error(err))
	}

	expensesDB := database.NewExpensesDB(db, cache)
	usersDB := database.NewUsersDB(db)
	ratesDB := database.NewRatesDB(db)
	limitsDB := database.NewLimitsDB(db)

	logger.Info("initializing telegram client")
	tgClient, err := tg.New(config)
	if err != nil {
		logger.Fatal("tg client init failed:", zap.Error(err))
	}

	currencyUpdateModel := currency.NewCbrCurrencyUpdater(config, ratesDB)

	msgModel := messages.New(tgClient, expensesDB, usersDB, ratesDB, limitsDB, currencyUpdateModel)
	callbackModel := callbacks.New(tgClient, expensesDB, usersDB, ratesDB)

	currencyRateWorker := worker.NewCurrencyRateWorker(currencyUpdateModel)
	updateListenerWorker := worker.NewUpdateListenerWorker(tgClient, msgModel, callbackModel, cache)

	metrics.CollectMetrics(logger)
	currencyRateWorker.Run(ctx, config.GetUpdateRate())
	updateListenerWorker.Run(ctx)
}
