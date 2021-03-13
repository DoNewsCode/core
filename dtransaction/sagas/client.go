package sagas

import "context"

type Client struct {
	hasSeen map[string]bool
}

type Receiver interface {
	Do(ctx context.Context, correlationId string)
	UnDo(ctx context.Context, correlationId string)
}

func (c *Client) Do() func(ctx context.Context, correlationId string) {
	if c.hasSeen == nil {
		c.hasSeen = make(map[string]bool)
	}
	return func(ctx context.Context, correlationId string) {
		// Avoid hanging
		if _, ok := c.hasSeen[correlationId]; ok {
			return
		}
		c.hasSeen[correlationId] = true
		c.realDo()
	}
}

func (c *Client) realDo() {}

func (c *Client) Undo() func(ctx context.Context, correlationId string) {
	return func(ctx context.Context, correlationId string) {
		// Empty Rollback
		if _, ok := c.hasSeen[correlationId]; !ok {
			c.hasSeen[correlationId] = true
			return
		}
		c.realUndo()
	}
}

func (c *Client) realUndo() {}
