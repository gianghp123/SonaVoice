package compensations

type Compensations struct {
	stack []func()
	done  bool
}

func New() *Compensations {
	return &Compensations{}
}

func (c *Compensations) Push(fn func()) {
	if c.done {
		return
	}
	c.stack = append(c.stack, fn)
}

func (c *Compensations) Pop() {
	if c.done || len(c.stack) == 0 {
		return
	}
	c.stack = c.stack[:len(c.stack)-1]
}

func (c *Compensations) Run() {
	if c.done {
		return
	}
	c.done = true
	for i := len(c.stack) - 1; i >= 0; i-- {
		c.stack[i]()
	}
	c.stack = nil
}
