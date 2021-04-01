package remote

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
)

func TestRemote(t *testing.T) {
	cfg := &clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 2 * time.Second,
	}

	r := Provider("config.yaml", cfg)

	var testVal = "name: app"
	// PREPARE TEST DATA
	if err := r.Put(testVal); err != nil {
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

	if err := r.Put(testVal); err != nil {
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
	err = r.Put("test")
	assert.Error(t, err)

	_, err = r.ReadBytes()
	assert.Error(t, err)

	err = r.Watch(context.Background(), func() error {
		return nil
	})
	assert.Error(t, err)

	cfg = &clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 2 * time.Second,
	}
	r = Provider("config-test1", cfg)
	_, err = r.ReadBytes()
	assert.Error(t, err)

	r = Provider("config-test2", cfg)
	go func() {
		err := r.Watch(context.Background(), func() error {
			return fmt.Errorf("for test")
		})
		assert.Error(t, err)
	}()

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := r.Watch(ctx, func() error {
			return fmt.Errorf("for test")
		})
		assert.Error(t, err)
	}()

	time.Sleep(1 * time.Second)
	if err := r.Put("name: test"); err != nil {
		t.Fatal(err)
	}
}
