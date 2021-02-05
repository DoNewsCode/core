package async

type Concurrent struct {
	cap chan struct{}
}

func NewConcurrent(max int) *Concurrent {
	caps := make(chan struct{}, max)
	for i := 0; i < max; i++ {
		caps <- struct{}{}
	}
	return &Concurrent{
		cap: caps,
	}
}

func (c *Concurrent) Go(callback func()) {
	<-c.cap
	go func() {
		defer func() {
			c.cap <- struct{}{}
		}()
		callback()
	}()
}
