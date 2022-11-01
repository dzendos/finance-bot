package worker

import (
	"context"
	"log"
	"time"
)

type updater interface {
	UpdateCurrencyRate(ctx context.Context) error
	Close()
}

type CurrencyRateWorker struct {
	updater updater
}

func NewCurrencyRateWorker(updater updater) *CurrencyRateWorker {
	return &CurrencyRateWorker{
		updater: updater,
	}
}

func (w *CurrencyRateWorker) Run(ctx context.Context, updateFrequency time.Duration) {
	go func() {
		ticker := time.NewTicker(updateFrequency)
		err := w.updater.UpdateCurrencyRate(ctx)
		if err != nil {
			log.Println(err)
		}

		for {
			select {
			case <-ctx.Done():
				w.updater.Close()
				return
			case <-ticker.C:
				select {
				case <-ctx.Done():
					w.updater.Close()
					return
				default:
					err := w.updater.UpdateCurrencyRate(ctx)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}()
}
