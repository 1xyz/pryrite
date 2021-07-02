package markdown

import "fmt"

type Cursor struct {
	Start int
	Stop  int
}

func (c *Cursor) Contains(pos int) bool {
	return c.Start <= pos && pos <= c.Stop
}

func (c *Cursor) String() string {
	return fmt.Sprint(c.Start, ":", c.Stop)
}
