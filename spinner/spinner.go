package spinner

import "fmt"

var ticks = [...]rune{'/', '-', '\\', '|'}

type Spinner struct {
	i int
}

func (s *Spinner) Increment() {
	out := fmt.Sprintf("[%s]", string(ticks[s.i%len(ticks)]))
	if s.i != 0 {
		out = "\b\b\b" + out
	}
	fmt.Print(out)
	s.i = s.i + 1
}
