package types

import (
	"fmt"
	"log"
	"time"
)

type Expense struct {
	ExpenseID int
	Sum       int
	Category  string
	Date      time.Time
}

func NewExpense() *Expense {
	return &Expense{
		Sum:      0,
		Category: "Новая категория",
		Date:     time.Now(),
	}
}

type UserModel struct {
	Currency     string
	CurrencyRate int
}

func (e *Expense) ToString(model *UserModel) string {
	log.Println(model.CurrencyRate, float64(e.Sum)/float64(model.CurrencyRate))
	return fmt.Sprintf(
		"Используемая валюта: "+model.Currency+"\n\n"+
			"Сумма: %.2f\nКатегория: %s\nДата: %s",
		float64(e.Sum)/float64(model.CurrencyRate),
		e.Category,
		e.Date.Format("2006-01-02"),
	)
}
