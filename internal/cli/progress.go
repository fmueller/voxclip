package cli

import (
	"os"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

type stopFunc func()

func startSpinner(enabled bool, description string) stopFunc {
	if !enabled {
		return func() {}
	}

	bar := progressbar.NewOptions(
		-1,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionThrottle(80*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionClearOnFinish(),
	)

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)
		ticker := time.NewTicker(120 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				_ = bar.Finish()
				return
			case <-ticker.C:
				_ = bar.Add(1)
			}
		}
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			close(stopCh)
			<-doneCh
		})
	}
}

func startDurationProgress(enabled bool, description string, duration time.Duration) stopFunc {
	if !enabled || duration <= 0 {
		return func() {}
	}

	total := int64(duration / time.Second)
	if total <= 0 {
		total = 1
	}

	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(20),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionClearOnFinish(),
	)

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				_ = bar.Finish()
				return
			case <-ticker.C:
				_ = bar.Add(1)
			}
		}
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			close(stopCh)
			<-doneCh
		})
	}
}
