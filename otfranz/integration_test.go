package otfranz

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	knoaf_json "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/oklog/run"
	"github.com/stretchr/testify/assert"
	"github.com/twmb/franz-go/pkg/kgo"
)

type channelWatcher struct {
	ch          chan struct{}
	afterReload chan struct{}
}

func (c *channelWatcher) Watch(ctx context.Context, reload func() error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.ch:
			reload()
			c.afterReload <- struct{}{}
		}
	}
}

func TestModule_hotReload(t *testing.T) {
	if os.Getenv("KAFKA_ADDR") == "" {
		t.Skip("set KAFKA_ADDR to run TestModule_ProvideRunGroup")
		return
	}
	addrs := strings.Split(os.Getenv("KAFKA_ADDR"), ",")
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cw := &channelWatcher{}
	cw.ch = make(chan struct{})
	cw.afterReload = make(chan struct{})

	conf := map[string]interface{}{
		"http": map[string]bool{
			"disable": true,
		},
		"grpc": map[string]bool{
			"disable": true,
		},
		"cron": map[string]bool{
			"disable": true,
		},
		"kafka": map[string]interface{}{
			"default": map[string]interface{}{
				"seed_brokers":          addrs,
				"default_produce_topic": "franz-foo",
			},
		},
		"log": map[string]string{
			"level": "none",
		},
	}
	path := createFile(conf)
	c := core.Default(core.WithConfigStack(file.Provider(path), knoaf_json.Parser()), core.WithConfigWatcher(cw))
	defer c.Shutdown()
	c.Provide(Providers(
		WithReload(true),
		WithInterceptor(func(name string, config *Config) {
			config.MaxBytes = 100
		}),
	))
	c.AddModuleFunc(config.New)

	var group run.Group
	c.ApplyRunGroup(&group)
	group.Add(func() error {
		<-ctx.Done()
		return ctx.Err()
	}, func(err error) {
		cancel()
	})
	go group.Run()

	c.Invoke(func(f Factory) {
		cli, err := f.Make("default")
		assert.NoError(t, err)
		record := &kgo.Record{Value: []byte("bar")}
		cli.Produce(ctx, record, func(r *kgo.Record, err error) {
			assert.Equal(t, "franz-foo", r.Topic)
		})
	})

	// Reload config
	conf["kafka"].(map[string]interface{})["default"].(map[string]interface{})["default_produce_topic"] = "franz-bar"
	overwriteFile(path, conf)
	cw.ch <- struct{}{}
	<-cw.afterReload

	// Test reloaded values
	c.Invoke(func(f Factory) {
		cli, err := f.Make("default")
		assert.NoError(t, err)
		record := &kgo.Record{Value: []byte("bar")}
		cli.Produce(ctx, record, func(r *kgo.Record, err error) {
			assert.Equal(t, "franz-bar", r.Topic)
		})
	})
}

func createFile(content map[string]interface{}) string {
	f, _ := ioutil.TempFile("", "*")
	data, _ := json.Marshal(content)
	ioutil.WriteFile(f.Name(), data, os.ModePerm)
	return f.Name()
}

func overwriteFile(path string, content map[string]interface{}) {
	data, _ := json.Marshal(content)
	ioutil.WriteFile(path, data, os.ModePerm)
}
