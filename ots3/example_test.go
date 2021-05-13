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
	uploader := NewManager(envDefaultS3AccessKey, envDefaultS3AccessSecret, envDefaultS3Endpoint, envDefaultS3Region, envDefaultS3Bucket)
	_ = uploader.CreateBucket(context.Background(), envDefaultS3Bucket)
	url, _ := uploader.Upload(context.Background(), "foo", file)
	url = strings.Replace(url, envDefaultS3Endpoint, "http://example.org", 1)
	fmt.Println(url)

	// Output:
	// http://example.org/mybucket/foo.png
}
