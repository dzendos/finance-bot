package types

import "time"

type CurrencyRate struct {
	CharCode     string
	BaseCurrency string
	Rate         int
	Date         time.Time
}

type EncodedCurrencies struct {
	CharCode string `xml:"CharCode"`
	Value    string `xml:"Value"`
}

type CurrenciesRate struct {
	EncodedCurrencies []EncodedCurrencies `xml:"Valute"`
}
