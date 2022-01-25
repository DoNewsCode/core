package eventsv2

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestEvent(t *testing.T) {
	type TestResult struct {
		Count *int
	}

	var successListener = func(ctx context.Context, event TestResult) error {
		*event.Count++
		return nil
	}

	var failListener = func(ctx context.Context, event TestResult) error {
		*event.Count++
		return errors.New("failure")
	}

	for _, c := range []struct {
		name     string
		action   func(*Event[TestResult])
		expected int
	}{
		{
			"empty",
			func(*Event[TestResult]) {},
			0,
		},
		{
			"one success listener",
			func(e *Event[TestResult]) {
				e.Subscribe(successListener)
			},
			1,
		},
		{
			"two success listener",
			func(e *Event[TestResult]) {
				e.Subscribe(successListener)
				e.Subscribe(successListener)
			},
			2,
		},
		{
			"one fail listener & one success listener",
			func(e *Event[TestResult]) {
				e.Subscribe(failListener)
				e.Subscribe(successListener)
			},
			1,
		},
		{
			"prepend fail listener",
			func(e *Event[TestResult]) {
				e.Prepend(successListener)
				e.Prepend(failListener)
			},
			1,
		},
		{
			"one fail listener",
			func(e *Event[TestResult]) {
				e.Subscribe(failListener)
			},
			1,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var count int
			event := &Event[TestResult]{}
			c.action(event)
			event.Dispatch(context.Background(), TestResult{Count: &count})
			if count != c.expected {
				t.Errorf("expected %d listeners, got %d", c.expected, count)
			}
		})
	}
}

func TestEvent_DoubleFiring(t *testing.T) {
	type TestResult struct {
		Count *int
	}

	var successListener = func(ctx context.Context, event TestResult) error {
		*event.Count++
		return nil
	}

	for _, c := range []struct {
		name     string
		action   func(*Event[TestResult])
		expected int
	}{
		{
			"empty",
			func(*Event[TestResult]) {},
			0,
		},
		{
			"one success listener",
			func(e *Event[TestResult]) {
				e.Subscribe(successListener)
			},
			2,
		},
		{
			"prependOnce",
			func(e *Event[TestResult]) {
				e.PrependOnce(successListener)
			},
			1,
		},
		{
			"subscribeOnce",
			func(e *Event[TestResult]) {
				e.SubscribeOnce(successListener)
			},
			1,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var count int
			event := &Event[TestResult]{}
			c.action(event)
			event.Dispatch(context.Background(), TestResult{Count: &count})
			event.Dispatch(context.Background(), TestResult{Count: &count})
			if count != c.expected {
				t.Errorf("expected %d listeners, got %d", c.expected, count)
			}
		})
	}
}

func TestEvent_Races(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	event := &Event[struct{}]{}
	for i := 0; i < 50; i++ {
		wg.Add(4)
		go func() {
			defer wg.Done()
			event.Dispatch(ctx, struct{}{})
		}()
		go func() {
			defer wg.Done()
			event.SubscribeOnce(func(ctx context.Context, event struct{}) error {
				return nil
			})
		}()
		go func() {
			defer wg.Done()
			event.PrependOnce(func(ctx context.Context, event struct{}) error {
				return nil
			})
		}()
		go func() {
			defer wg.Done()
			event.ListenerCount()
		}()
	}
	wg.Wait()
}
