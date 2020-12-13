// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

import (
	"context"
	"time"
)

func Go(ctx context.Context, work func(ctx context.Context)) {
	go work(ctx)
}

func DoTimerSchedule(ctx context.Context, work func(), delay time.Duration, supplier func() time.Duration) {
	go func() {
		timer := time.NewTimer(delay)

		for {
			select {
			case <-ctx.Done():
				timer.Stop()
			case <-timer.C:
				work()
				timer.Reset(supplier())
			}
		}

	}()
}

func DoTickerSchedule(ctx context.Context, work func(), delay time.Duration) {
	go func() {
		ticker := time.NewTicker(delay)

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
			case <-ticker.C:
				work()
			}
		}

	}()

}

type HashTimeWheel struct {
	tickDuration  time.Duration
	ticksPerWheel int32
}

func (htw *HashTimeWheel) Start()  {
	
}

type Timeout interface {
	Run()
}
