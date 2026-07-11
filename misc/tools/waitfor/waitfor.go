package waitfor

import (
	"os"
	"sync"
	"time"

	"github.com/getaudited/audited/internal/common/logs"
)

type WaitFor struct {
	logger *logs.Logger
	wg     *sync.WaitGroup
}

func NewWaitFor(logger *logs.Logger) *WaitFor {
	return &WaitFor{
		logger: logger,
		wg:     &sync.WaitGroup{},
	}
}

func (w *WaitFor) Do(handler func() error, svcName string, timeout time.Duration) {
	w.logger.Info("⏱️WaitFor started", "service", svcName)

	w.wg.Go(func() {
		until := time.NewTimer(timeout)

		for {
			select {
			case <-until.C:
				w.logger.Error("⛔Timeout reached", "service", svcName)

				// Exit with an error because the checks must be all or nothing.
				os.Exit(1)
			default:
				err := handler()
				if err != nil {
					w.logger.Warn("⚠️ Check failed, trying again...in 250ms", "service", svcName, "error", err)
					time.Sleep(time.Millisecond * 250)
					continue
				}

				if err == nil {
					w.logger.Info("✅Ready", "service", svcName)
					return
				}
			}
		}
	})
}

func (w *WaitFor) Wait() {
	w.wg.Wait()
}
