package remote

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRemote_Read(t *testing.T) {
	if os.Getenv("ETCD_ADDR") == "" {
		t.Skip("set ETCD_ADDR to run TestRemote_Read")
		return
	}
	r := Provider(Config{
		Name:      "etcd",
		Endpoints: strings.Split(os.Getenv("ETCD_ADDR"), ","),
		Key:       "remote_read_test",
	})

	err := r.set("remote_read_test", []byte("test"))
	assert.NoError(t, err)

	_, err = r.Read()
	assert.Error(t, err)

	res, err := r.ReadBytes()
	assert.NoError(t, err)
	assert.Equal(t, []byte("test"), res)
}

func TestRemote_Watch(t *testing.T) {
	if os.Getenv("ETCD_ADDR") == "" {
		t.Skip("set ETCD_ADDR to run TestRemote_Watch")
		return
	}
	r := Provider(Config{
		Name:      "etcd",
		Endpoints: strings.Split(os.Getenv("ETCD_ADDR"), ","),
		Key:       "remote_watch_test",
	})

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

	err := r.set("remote_watch_test", []byte("test"))
	assert.NoError(t, err)

	newVal := <-ch
	assert.Equal(t, "test", newVal)

}
