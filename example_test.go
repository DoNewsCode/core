package core_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/gorilla/mux"
)

func ExampleC_AddModuleFunc() {
	type Foo struct{}
	c := core.New()
	c.AddModuleFunc(func() (Foo, func(), error) {
		return Foo{}, func() {}, nil
	})
	fmt.Printf("%T\n", c.Modules()...)
	// Output:
	// core_test.Foo
}

func ExampleC_AddModule() {
	type Foo struct{}
	c := core.New()
	c.AddModule(Foo{})
	fmt.Printf("%T\n", c.Modules()...)
	// Output:
	// core_test.Foo
}

func ExampleC_Provide() {
	type Foo struct {
		Value string
	}
	type Bar struct {
		foo Foo
	}
	c := core.New()
	c.Provide(func() (foo Foo, cleanup func(), err error) {
		return Foo{
			Value: "test",
		}, func() {}, nil
	})
	c.Provide(func(foo Foo) Bar {
		return Bar{foo: foo}
	})
	c.Invoke(func(bar Bar) {
		fmt.Println(bar.foo.Value)
	})
	// Output:
	// test
}

func Example_minimal() {

	// Spin up a real server
	c := core.Default(core.WithInline("log.level", "none"))
	c.AddModule(core.HttpFunc(func(router *mux.Router) {
		router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			writer.Write([]byte("hello world"))
		})
	}))
	ctx, cancel := context.WithCancel(context.Background())
	go c.Serve(ctx)

	// Giver server sometime to be ready.
	time.Sleep(time.Second)

	// Let's try if the server works.
	resp, _ := http.Get("http://localhost:8080/")
	bytes, _ := ioutil.ReadAll(resp.Body)
	cancel()

	fmt.Println(string(bytes))
	// Output:
	// hello world
}
