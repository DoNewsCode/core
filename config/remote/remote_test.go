package remote

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
)

func TestRemote(t *testing.T) {
	cfg := &clientv3.Config{
		Endpoints:   []string{"192.168.82.3:2379"},
		DialTimeout: 2 * time.Second,
	}

	r := Provider("config.yaml", cfg)

	var testVal = "name: app"
	// PREPARE TEST DATA
	if err := put(r, testVal); err != nil {
		t.Fatal(err)
	}

	_, err := r.Read()
	assert.Error(t, err)

	bytes, err := r.ReadBytes()
	assert.NoError(t, err)
	assert.Equal(t, testVal, string(bytes))

	var ch = make(chan string)
	go r.Watch(context.Background(), func() error {
		bytes, err := r.ReadBytes()
		if err != nil {
			ch <- ""
			return err
		}
		ch <- string(bytes)
		return nil
	})

	time.Sleep(1 * time.Second)

	if err := put(r, testVal); err != nil {
		t.Fatal(err)
	}

	newVal := <-ch
	assert.Equal(t, testVal, newVal)
}

func TestError(t *testing.T) {
	var (
		r   *Remote
		err error
	)

	cfg := &clientv3.Config{
		Endpoints: []string{},
	}

	r = Provider("config.yaml", cfg)
	err = put(r, "test")
	assert.Error(t, err)

	_, err = r.ReadBytes()
	assert.Error(t, err)

	err = r.Watch(context.Background(), func() error {
		return nil
	})
	assert.Error(t, err)

	cfg = &clientv3.Config{
		Endpoints:   []string{"192.168.82.3:2379"},
		DialTimeout: 2 * time.Second,
	}
	r = Provider("config-test1", cfg)
	_, err = r.ReadBytes()
	assert.Error(t, err)

	r = Provider("config-test2", cfg)

	// Confirm that the two coroutines are finished
	g := sync.WaitGroup{}
	g.Add(2)
	go func() {
		err := r.Watch(context.Background(), func() error {
			return fmt.Errorf("for test")
		})
		assert.Error(t, err)
		g.Done()
	}()

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := r.Watch(ctx, func() error {
			return fmt.Errorf("for test")
		})
		assert.Error(t, err)
		g.Done()
	}()

	time.Sleep(1 * time.Second)
	if err := put(r, "name: test"); err != nil {
		t.Fatal(err)
	}
	g.Wait()
}

func put(r *Remote, val string) error {
	client, err := clientv3.New(*r.clientConfig)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.Put(context.Background(), r.key, val)
	if err != nil {
		return err
	}
	return nil
}
