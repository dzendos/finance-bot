package types

// State tells what we are modifying un the expense.
type State int

const (
	EditingSum State = iota + 1
	EditingCategory
	EditingDate
	EditingLimit
	WaitState
)

// CurrentState contains id on the expense we are modifying now, and what we are modifying.
type CurrentState struct {
	ExpenseID int
	State     State
}
