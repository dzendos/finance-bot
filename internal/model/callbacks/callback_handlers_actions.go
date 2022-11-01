package callbacks

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/types"
)

func (s *Model) toWriteSumState(ctx context.Context, data *CallbackData) error {
	// Change state of the user - he is now entering sum for this user and this messageID.
	err := s.usersDB.SetCurrentState(ctx, data.FromID, types.CurrentState{
		ExpenseID: data.MessageID,
		State:     types.EditingSum,
	})

	if err != nil {
		return errors.Wrap(err, "cannot SetCurrentState")
	}

	// Show notification about the action.
	return s.tgClient.ShowAlert("Введите сумму", data.CallbackID)
}

func (s *Model) toWriteCategoryState(ctx context.Context, data *CallbackData) error {
	// Change state of the user - he is now entering sum for this user and this messageID.

	err := s.usersDB.SetCurrentState(ctx, data.FromID, types.CurrentState{
		ExpenseID: data.MessageID,
		State:     types.EditingCategory,
	})

	if err != nil {
		return errors.Wrap(err, "cannot SetCurrentState")
	}

	// Show notification about the action.
	return s.tgClient.ShowAlert("Введите категорию", data.CallbackID)
}

func (s *Model) toWriteDateState(ctx context.Context, data *CallbackData) error {
	// Change state of the user - he is now entering sum for this user and this messageID.
	err := s.usersDB.SetCurrentState(ctx, data.FromID, types.CurrentState{
		ExpenseID: data.MessageID,
		State:     types.EditingDate,
	})

	if err != nil {
		return errors.Wrap(err, "cannot SetCurrentState")
	}

	// Show notification about the action.
	return s.tgClient.ShowAlert("Введите дату в формате YYYY-MM-DD", data.CallbackID)
}

func (s *Model) saveExpense(data *CallbackData) error {
	// Go to cancel expose to delete it from current expenses.
	return s.tgClient.DoneMessage(data.FromID, data.MessageID)
}

func (s *Model) cancelExpense(ctx context.Context, data *CallbackData) error {
	err := s.expensesDB.DeleteExpense(ctx, data.FromID, data.MessageID)
	if err != nil {
		return errors.Wrap(err, "cannot DeleteExpense")
	}

	return s.tgClient.CancelMessage(data.FromID, data.MessageID)
}

func (s *Model) getWeekReport(ctx context.Context, data *CallbackData) error {
	dateBegin := getDayBegin(time.Now().AddDate(0, 0, -7))
	dateEnd := getDayEnd(time.Now())
	return s.getReport(ctx, data, dateBegin, dateEnd)
}

func (s *Model) getMonthReport(ctx context.Context, data *CallbackData) error {
	dateBegin := getDayBegin(time.Now().AddDate(0, -1, 0))
	dateEnd := getDayEnd(time.Now())
	return s.getReport(ctx, data, dateBegin, dateEnd)
}

func (s *Model) getYearReport(ctx context.Context, data *CallbackData) error {
	dateBegin := getDayBegin(time.Now().AddDate(-1, 0, 0))
	dateEnd := getDayEnd(time.Now())
	return s.getReport(ctx, data, dateBegin, dateEnd)
}

func (s *Model) getReport(ctx context.Context, data *CallbackData, dateBegin, dateEnd time.Time) error {
	report, err := s.expensesDB.GetReport(ctx, data.FromID, dateBegin, dateEnd)

	if err != nil {
		return errors.Wrap(err, "cannot GetReport")
	}

	reportMessage, err := s.reportMessage(ctx, report, dateBegin, dateEnd, data.FromID)

	if err != nil {
		return errors.Wrap(err, "cannot reportMessage")
	}

	return s.tgClient.SendMessage(reportMessage, data.FromID)
}

func (s *Model) reportMessage(ctx context.Context, report map[string]int, dateBegin, dateEnd time.Time, userID int64) (string, error) {
	result := fmt.Sprintf("Отчет в период с %s по %s\n\n",
		dateBegin.Format("2006-01-02"),
		dateEnd.Format("2006-01-02"))

	st, ok := s.usersDB.GetCurrentState(ctx, userID)
	currentCurrency := types.RUB
	if ok && st.Currency != "" {
		currentCurrency = st.Currency
	}

	currencyRate, err := s.ratesDB.GetCurrencyRate(ctx, currentCurrency, time.Now())

	if err != nil {
		return "", errors.Wrap(err, "cannot GetCurrencyRate")
	}

	for category, sum := range report {
		result += fmt.Sprintf("%s: %.2f\n", category, float64(sum)/float64(currencyRate))
	}

	return result, nil
}

func getDayBegin(date time.Time) time.Time {
	date.Add(-time.Duration(date.Hour()) * time.Hour)
	date.Add(-time.Duration(date.Minute()) * time.Minute)
	date.Add(-time.Duration(date.Second()) * time.Second)

	return date
}

func getDayEnd(date time.Time) time.Time {
	date.Add(time.Duration(23-date.Hour()) * time.Hour)
	date.Add(time.Duration(59-date.Minute()) * time.Minute)
	date.Add(time.Duration(59-date.Second()) * time.Second)

	return date
}

func (s *Model) changeCurrentCurrency(ctx context.Context, data *CallbackData) error {
	err := s.usersDB.SetUserCurrency(ctx, data.FromID, types.Currency(data.Data))

	if err != nil {
		return errors.Wrap(err, "cannot SetUserCurrency")
	}

	currency, err := s.usersDB.GetUserCurrency(ctx, data.FromID)

	if err != nil {
		return errors.Wrap(err, "cannot GetUserCurrency")
	}

	message := "Текущая валюта: " + string(currency)
	err = s.tgClient.EditMessage(message, data.FromID, data.MessageID)

	if err != nil {
		return errors.Wrap(err, "cannot EditMessage")
	}

	return nil
}
