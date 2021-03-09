package progressbar

import (
	"fmt"
	"strings"
)

type Bar struct {
	total int    // total value for progress
	per   int    // current in percentage
	bar   string // the actual progress bar to be printed
}

func New(total int) *Bar {
	return &Bar{
		total: total,
		bar:   strings.Repeat("-", 100),
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
		fmt.Println("\nDone!")
	}
}
