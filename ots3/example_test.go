package ots3

import (
	"context"
	"fmt"
	"os"
	"strings"
)

func Example() {
	if os.Getenv("S3_ENDPOINT") == "" {
		fmt.Println("set S3_ENDPOINT to run this exmaple")
		return
	}
	if os.Getenv("S3_ACCESSKEY") == "" {
		fmt.Println("set S3_ACCESSKEY to run this exmaple")
		return
	}
	if os.Getenv("S3_ACCESSSECRET") == "" {
		fmt.Println("set S3_ACCESSSECRET to run this exmaple")
		return
	}
	if os.Getenv("S3_BUCKET") == "" {
		fmt.Println("set S3_BUCKET to run this exmaple")
		return
	}
	if os.Getenv("S3_REGION") == "" {
		fmt.Println("set S3_REGION to run this exmaple")
		return
	}
	file, err := os.Open("./testdata/basn0g01-30.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	uploader := NewManager(os.Getenv("S3_ACCESSKEY"), os.Getenv("S3_ACCESSSECRET"), os.Getenv("S3_ENDPOINT"), os.Getenv("S3_REGION"), os.Getenv("S3_BUCKET"))
	_ = uploader.CreateBucket(context.Background(), os.Getenv("S3_BUCKET"))
	url, _ := uploader.Upload(context.Background(), "foo", file)
	url = strings.Replace(url, os.Getenv("S3_ENDPOINT"), "http://example.org", 1)
	fmt.Println(url)

	// Output:
	// http://example.org/mybucket/foo.png
}
