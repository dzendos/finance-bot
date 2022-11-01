package tg

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

type tokenGetter interface {
	Token() string
}

type Client struct {
	client *tgbotapi.BotAPI
}

func New(tokenGetter tokenGetter) (*Client, error) {
	client, err := tgbotapi.NewBotAPI(tokenGetter.Token())
	if err != nil {
		return nil, errors.Wrap(err, "cannot NewBotAPI")
	}

	return &Client{
		client: client,
	}, nil
}

func (c *Client) SendMessage(text string, userID int64) error {
	_, err := c.client.Send(tgbotapi.NewMessage(userID, text))
	if err != nil {
		return errors.Wrap(err, "cannot Send")
	}
	return nil
}

func (c *Client) CreateExpense(text string, userID int64) error {
	msg := tgbotapi.NewMessage(userID, text)

	msg.ReplyMarkup = createExpenseKeyboard

	_, err := c.client.Send(msg)

	if err != nil {
		return errors.Wrap(err, "cannot Send")
	}

	return nil
}

func (c *Client) ShowAlert(text string, messageID string) error {
	alert := tgbotapi.NewCallback(messageID, text)
	_, err := c.client.Request(alert)

	if err != nil {
		return errors.Wrap(err, "cannot Request")
	}

	return nil
}

func (c *Client) EditExpenseMessage(text string, userID int64, messageID int) error {
	editMessage := tgbotapi.NewEditMessageTextAndMarkup(userID, messageID, text, createExpenseKeyboard)
	_, err := c.client.Send(editMessage)

	if err != nil {
		return errors.Wrap(err, "cannot Send")
	}

	return nil
}

func (c *Client) EditMessage(text string, userID int64, messageID int) error {
	editMessage := tgbotapi.NewEditMessageText(userID, messageID, text)
	_, err := c.client.Send(editMessage)

	if err != nil {
		return errors.Wrap(err, "cannot Send")
	}

	return nil
}

func (c *Client) DeleteMessage(userID int64, messageID int) error {
	deleteMessage := tgbotapi.NewDeleteMessage(userID, messageID)
	_, err := c.client.Request(deleteMessage)

	if err != nil {
		return errors.Wrap(err, "cannot Request")
	}

	return nil
}

func (c *Client) DoneMessage(userID int64, messageID int) error {
	editMessage := tgbotapi.NewEditMessageText(userID, messageID, "Сохранено")
	_, err := c.client.Send(editMessage)

	if err != nil {
		return errors.Wrap(err, "cannot Send")
	}

	return nil
}

func (c *Client) CancelMessage(userID int64, messageID int) error {
	editMessage := tgbotapi.NewEditMessageText(userID, messageID, "Отменено")
	_, err := c.client.Send(editMessage)

	if err != nil {
		return errors.Wrap(err, "cannot Send")
	}

	return nil
}

func (c *Client) GetReport(text string, userID int64) error {
	msg := tgbotapi.NewMessage(userID, text)

	msg.ReplyMarkup = getReportKeyboard
	_, err := c.client.Send(msg)

	if err != nil {
		return errors.Wrap(err, "cannot Send")
	}

	return nil
}

func (c *Client) ChangeCurrency(text string, userID int64) error {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = changeCurrencyKeyboard
	_, err := c.client.Send(msg)

	if err == nil {
		return errors.Wrap(err, "cannot Send")
	}

	return nil
}

func (c *Client) Start() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return c.client.GetUpdatesChan(u)
}

func (c *Client) Stop() {
	log.Println("Stop receiving updates")
	c.client.StopReceivingUpdates()
}

func (c *Client) Request(callback tgbotapi.CallbackConfig) error {
	_, err := c.client.Request(callback)
	return errors.Wrap(err, "cannot Request")
}
