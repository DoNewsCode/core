package ots3

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/DoNewsCode/core/key"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	t.Parallel()
	assert.NotNil(t, NewManager(
		"",
		"",
		"",
		"",
		"",
		WithTracer(opentracing.GlobalTracer()),
		WithPathPrefix(""),
		WithHTTPClient(http.DefaultClient),
		WithLocationFunc(func(location string) (url string) {
			return ""
		}),
		WithKeyer(key.New()),
		WithAutoExtension(true),
	))
}

func TestManager_CreateBucket(t *testing.T) {
	t.Parallel()
	m := setupManager()
	err := m.CreateBucket(context.Background(), "foo")
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				return
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				return
			default:
				t.Fail()
			}
		} else {
			t.Fail()
		}
		return
	}
}

func TestManager_UploadFromUrl(t *testing.T) {
	if os.Getenv("S3_ENDPOINT") == "" {
		t.Skip("set S3_ENDPOINT to run TestManager_UploadFromUrl")
		return
	}
	if os.Getenv("S3_ACCESSKEY") == "" {
		t.Skip("set S3_ACCESSKEY to run TestManager_UploadFromUrl")
		return
	}
	if os.Getenv("S3_ACCESSSECRET") == "" {
		t.Skip("set S3_ACCESSSECRET to run TestManager_UploadFromUrl")
		return
	}
	if os.Getenv("S3_BUCKET") == "" {
		t.Skip("set S3_BUCKET to run TestManager_UploadFromUrl")
		return
	}
	if os.Getenv("S3_REGION") == "" {
		t.Skip("set S3_REGION to run TestManager_UploadFromUrl")
		return
	}
	tracer := mocktracer.New()
	m := setupManagerWithTracer(tracer)
	_ = m.CreateBucket(context.Background(), os.Getenv("S3_BUCKET"))
	newURL, err := m.UploadFromUrl(context.Background(), "https://avatars.githubusercontent.com/u/43054062")
	assert.NoError(t, err)
	assert.NotEmpty(t, newURL)
	assert.Len(t, tracer.FinishedSpans(), 2)
}

func setupManager() *Manager {
	return setupManagerWithTracer(nil)
}

func setupManagerWithTracer(tracer opentracing.Tracer) *Manager {
	m := NewManager(
		os.Getenv("S3_ACCESSKEY"),
		os.Getenv("S3_ACCESSSECRET"),
		os.Getenv("S3_ENDPOINT"),
		os.Getenv("S3_REGION"),
		os.Getenv("S3_BUCKET"),
		WithTracer(tracer),
	)
	return m
}
