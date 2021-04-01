package remote

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
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
