package progressbar

import "fmt"

type Bar struct {
	cur   int    // current progress
	total int    // total value for progress
	per   int    // current in percentage
	bar   string // the actual progress bar to be printed
}

func New(total int) *Bar {
	return &Bar{total: total}
}

func (b *Bar) Update(newCur int) {
	newPer := newCur * 100 / b.total
	if newPer == b.per {
		return
	}

	b.cur = newCur
	b.per = newPer
	b.bar += "#"

	fmt.Printf("\r%v", b.bar)
}
