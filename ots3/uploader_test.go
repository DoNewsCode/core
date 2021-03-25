// +build integration

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
		WithHttpClient(http.DefaultClient),
		WithLocationFunc(func(location string) (url string) {
			return ""
		}),
		WithKeyer(key.New()),
	))
}

func TestMain(m *testing.M) {
	manager := NewManager("minioadmin", "minioadmin", "http://localhost:9000", "asia", "mybucket")
	_ = manager.CreateBucket(context.Background(), "foo")
	os.Exit(m.Run())
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
	tracer := mocktracer.New()
	m := setupManagerWithTracer(tracer)
	_ = m.CreateBucket(context.Background(), "mybucket")
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
		"minioadmin",
		"minioadmin",
		"http://localhost:9000",
		"asia",
		"mybucket",
		WithTracer(tracer),
	)
	return m
}
