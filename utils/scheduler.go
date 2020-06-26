// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

import (
	"context"
	"time"
)

func DoTimerSchedule(work func(), delay time.Duration, supplier func() time.Duration, ctx context.Context) {
	go func() {
		timer := time.NewTimer(delay)

		for  {
			select {
			case <- ctx.Done():
				timer.Stop()
			case <- timer.C:
				work()
				timer.Reset(supplier())
			}
		}

	}()

}

func DoTickerSchedule(work func(), delay time.Duration, ctx context.Context) {
	go func() {
		ticker := time.NewTicker(delay)

		for  {
			select {
			case <- ctx.Done():
				ticker.Stop()
			case <- ticker.C:
				work()
			}
		}

	}()

}
