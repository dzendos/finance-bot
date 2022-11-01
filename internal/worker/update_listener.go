package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/model/callbacks"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/model/messages"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/redis"
)

var (
	SentMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ozon",
			Subsystem: "messages_handler",
			Name:      "request_total",
		},
		[]string{"user", "message"},
	)

	MessageResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ozon",
			Subsystem: "messages_handler",
			Name:      "response_time",
		},
		[]string{"status"},
	)

	CallbacksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ozon",
			Subsystem: "callback_handler",
			Name:      "request_total",
		},
		[]string{"user", "callback"},
	)

	CallbackResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ozon",
			Subsystem: "callback_handler",
			Name:      "response_time",
		},
		[]string{"status"},
	)
)

type updateFetcher interface {
	Start() tgbotapi.UpdatesChannel
	Request(callback tgbotapi.CallbackConfig) error
	Stop()
}

type MessageHandler interface {
	IncomingMessage(ctx context.Context, msg *messages.Message) error
}

type CallbackHandler interface {
	IncomingCallback(ctx context.Context, callback *callbacks.CallbackData) error
}

type updateListenerWorker struct {
	updateFetcher   updateFetcher
	messageHandler  MessageHandler
	callbackHandler CallbackHandler
	cache           *redis.Cache
}

func NewUpdateListenerWorker(updateFetcher updateFetcher,
	messageHandler MessageHandler, callbackHandler CallbackHandler, cache *redis.Cache) *updateListenerWorker {
	return &updateListenerWorker{
		updateFetcher:   updateFetcher,
		messageHandler:  messageHandler,
		callbackHandler: callbackHandler,
		cache:           cache,
	}
}

func (w *updateListenerWorker) Run(ctx context.Context) {
	updates := w.updateFetcher.Start()

	for {
		select {
		case <-ctx.Done():
			w.cache.Close()
			w.updateFetcher.Stop()
			return
		case update, ok := <-updates:
			if !ok {
				w.cache.Close()
				w.updateFetcher.Stop()
				return
			}
			err := w.HandleUpdate(ctx, update)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (w *updateListenerWorker) HandleUpdate(ctx context.Context, update tgbotapi.Update) error {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		"HandleUpdate",
	)

	defer span.Finish()

	if update.Message != nil {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		user := fmt.Sprintf("%d", update.SentFrom().ID)
		SentMessagesTotal.WithLabelValues(user, update.Message.Text).Inc()

		startTime := time.Now()
		err := w.messageHandler.IncomingMessage(ctx, &messages.Message{
			Text:      update.Message.Text,
			UserID:    update.Message.From.ID,
			MessageID: update.Message.MessageID,
		})

		duration := time.Since(startTime)

		if err == nil {
			MessageResponseTime.WithLabelValues("success").Observe(duration.Seconds())
		} else {
			MessageResponseTime.WithLabelValues("error").Observe(duration.Seconds())
			return errors.Wrap(err, "cannot IncomingMessage")
		}
	} else if update.CallbackQuery != nil {
		log.Printf("[%s] data: %s",
			update.CallbackQuery.Message.From.UserName,
			update.CallbackQuery.Data,
		)

		user := fmt.Sprintf("%d", update.SentFrom().ID)
		CallbacksTotal.WithLabelValues(user, update.CallbackData()).Inc()

		startTime := time.Now()
		err := w.callbackHandler.IncomingCallback(ctx, &callbacks.CallbackData{
			Data:       update.CallbackData(),
			FromID:     update.CallbackQuery.From.ID,
			MessageID:  update.CallbackQuery.Message.MessageID,
			CallbackID: update.CallbackQuery.ID,
		})
		duration := time.Since(startTime)

		if err == nil {
			CallbackResponseTime.WithLabelValues("success").Observe(duration.Seconds())
		} else {
			CallbackResponseTime.WithLabelValues("error").Observe(duration.Seconds())
			return errors.Wrap(err, "cannot IncomingCallback")
		}
	}

	return nil
}
