package ots3

import (
	"context"
	"fmt"
	"os"
	"strings"
)

func Example() {
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
