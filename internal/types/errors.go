package types

import "errors"

var (
	ErrNoCurrency     = errors.New("user does not have any currency")
	ErrNoCurrencyRate = errors.New("currency rate for this date does not exist")
)
