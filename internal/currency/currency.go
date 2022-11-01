package currency

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/types"
	"golang.org/x/net/html/charset"
)

const kopecksInRouble float64 = 100.0

type config interface {
	GetUrl() string
	GetUpdateRate() time.Duration
}

type ratesDB interface {
	GetCurrencyRate(ctx context.Context, currency types.Currency, date time.Time) (int, error)
	SetCurrencyRate(ctx context.Context, currency types.CurrencyRate) error
}

type CbrCurrencyUpdater struct {
	CurrencyMtx sync.Mutex
	config      config
	ratesDB     ratesDB
}

func NewCbrCurrencyUpdater(config config, ratesDB ratesDB) *CbrCurrencyUpdater {
	return &CbrCurrencyUpdater{
		config:  config,
		ratesDB: ratesDB,
	}
}

func (c *CbrCurrencyUpdater) UpdateCurrencyRate(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	log.Println("Updating currency rate", time.Now())
	c.CurrencyMtx.Lock()
	defer c.CurrencyMtx.Unlock()

	date := time.Now().Format("02/01/2006")
	url := c.config.GetUrl() + date
	log.Println(url)
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return errors.Wrap(err, "cannot NewRequestWithContext")
	}

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		return errors.Wrap(err, "cannot Do")
	}
	defer response.Body.Close()

	err = c.encode(ctx, response, string(types.RUB), time.Now())

	if err != nil {
		return errors.Wrap(err, "cannot encode")
	}

	log.Println("Finished updating currency rate", time.Now())

	return nil
}

func (c *CbrCurrencyUpdater) Close() {
	log.Println("Closing Cbr currency rate updater")
}

func (c *CbrCurrencyUpdater) encode(ctx context.Context, response *http.Response, base string, date time.Time) error {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.Wrap(err, "cannot ReadAll")
	}

	rates := types.CurrenciesRate{}

	reader := bytes.NewReader(body)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&rates)

	if err != nil {
		return errors.Wrap(err, "cannot Decode")
	}

	for _, rate := range rates.EncodedCurrencies {
		switch rate.CharCode {
		case string(types.USD), string(types.CNY), string(types.EUR):
			value, err := getIntValueFromCommaFloat(rate.Value)

			if err != nil {
				return errors.Wrap(err, "cannot getIntValueFromCommaFloat")
			}

			currency := types.CurrencyRate{
				CharCode:     rate.CharCode,
				Rate:         value,
				BaseCurrency: base,
				Date:         date,
			}

			err = c.ratesDB.SetCurrencyRate(ctx, currency)
			if err != nil {
				return errors.Wrap(err, "cannot SetCurrencyState")
			}
		}
	}

	if err != nil {
		return errors.Wrap(err, "currency.encode")
	}

	err = c.ratesDB.SetCurrencyRate(ctx, types.CurrencyRate{
		CharCode:     string(types.RUB),
		Rate:         int(kopecksInRouble),
		BaseCurrency: string(types.RUB),
		Date:         date,
	})

	if err != nil {
		return errors.Wrap(err, "cannot SetCurrencyState")
	}

	return nil
}

func getIntValueFromCommaFloat(value string) (int, error) {
	separated := strings.Split(value, ",")

	var result int64
	var err error

	result, err = strconv.ParseInt(separated[0]+separated[1][:2], 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "cannot ParseInt")
	}

	return int(result), nil
}
