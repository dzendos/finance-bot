package types

type CurrentExpense struct {
	ExpenseID int
	Expense   Expense
}

type Currency string

const (
	USD Currency = "USD"
	CNY Currency = "CNY"
	EUR Currency = "EUR"
	RUB Currency = "RUB"
)

type UserStateType struct {
	CurrentState CurrentState // Contains the expense we are modifying now, and what we are modifying.
	Currency     Currency     // With which currency the user is working now.
}
