package progressbar

import (
	"fmt"
	"strings"
	"time"
)

type Bar struct {
	total int    // total value for progress
	per   int    // current in percentage
	bar   string // the actual progress bar to be printed
	start time.Time
}

func New(total int) *Bar {
	start := time.Now()
	return &Bar{
		total: total,
		bar:   strings.Repeat("-", 100),
		start: start,
	}
}

func (b *Bar) Update(newCur int) {
	newPer := (newCur * 100 / b.total) + 1
	if newPer == b.per {
		return
	}

	b.per = newPer
	b.bar = strings.Replace(b.bar, "-", "▊", 1)

	fmt.Printf("\r[%v/100] %v", b.per, b.bar)

	if b.per == 100 {
		elapsed := time.Since(b.start)
		fmt.Printf("\nDone after %s!\n", elapsed)
	}
}
