package ots3

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"net/http"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/key"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gabriel-vasile/mimetype"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

// Manager manages S3 uploads.
type Manager struct {
	bucket       string
	sess         *session.Session
	tracer       opentracing.Tracer
	doer         contract.HttpDoer
	pathPrefix   string
	keyer        contract.Keyer
	locationFunc func(location string) (url string)
}

// Config contains a various of configurations for Manager. It is mean to be modified by Option.
type Config struct {
	accessKey    string
	accessSecret string
	region       string
	bucket       string
	sess         *session.Session
	tracer       opentracing.Tracer
	doer         contract.HttpDoer
	keyer        contract.Keyer
	pathPrefix   string
	locationFunc func(location string) (url string)
}

// Option is the type of functional options to alter Config.
type Option func(*Config)

// WithTracer is an option that add opentracing.Tracer via the hook of S3 client.
func WithTracer(tracer opentracing.Tracer) Option {
	return func(c *Config) {
		c.tracer = tracer
	}
}

// WithPathPrefix is an option that changes the path prefix of uploaded file.
func WithPathPrefix(pathPrefix string) Option {
	return func(c *Config) {
		c.pathPrefix = pathPrefix
	}
}

// WithKeyer is an option that changes the path of the uploaded file.
func WithKeyer(keyer contract.Keyer) Option {
	return func(c *Config) {
		c.keyer = keyer
	}
}

// WithHttpClient is an option that replaces the default http client. Useful for interceptors like tracing and metrics.
func WithHttpClient(client contract.HttpDoer) Option {
	return func(c *Config) {
		c.doer = client
	}
}

// WithLocationFunc is an option that decides the how url is mapped to S3 bucket and path.
// Useful when not serving file directly from S3, but from a CDN.
func WithLocationFunc(f func(location string) (url string)) Option {
	return func(c *Config) {
		c.locationFunc = f
	}
}

// NewManager creates a new S3 manager
func NewManager(accessKey, accessSecret, endpoint, region, bucket string, opts ...Option) *Manager {
	c := &Config{
		doer:  http.DefaultClient,
		keyer: key.New(),
		locationFunc: func(location string) (url string) {
			return location
		},
	}
	for _, f := range opts {
		f(c)
	}

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, accessSecret, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
	sess := session.Must(session.NewSession(s3Config))
	c.keyer.Key("/")
	m := &Manager{
		bucket:       bucket,
		sess:         sess,
		tracer:       c.tracer,
		doer:         c.doer,
		pathPrefix:   c.pathPrefix,
		keyer:        c.keyer,
		locationFunc: c.locationFunc,
	}

	// add opentracing capabilities if opt in
	if c.tracer != nil {
		sess.Handlers.Build.PushFront(m.otHandler())
	}
	return m
}

// Upload uploads an io.reader to the S3 server, and returns the url on S3. The extension of the uploaded file
// is auto detected.
func (m *Manager) Upload(ctx context.Context, name string, reader io.Reader) (newUrl string, err error) {

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(m.sess)
	var extension = ""
	var buf = bytes.NewBuffer(nil)
	var tee = io.TeeReader(reader, buf)
	mi, err := mimetype.DetectReader(tee)
	if err == nil {
		extension = mi.Extension()
	}

	k := key.KeepOdd(m.keyer).Key("/", name+extension)

	// Efficiently use the buf for mime type reading and continue from the rest of the body
	result, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(m.pathPrefix + k),
		Body:   io.MultiReader(buf, reader),
	})

	if err != nil {
		return "", errors.Wrap(err, "unable to upload from io reader")
	}

	return m.locationFunc(result.Location), nil
}

// UploadFromUrl fetches a file from an external url, copy them to the S3 server, and generate a new, local url.
// It uses streams to relay files (instead of buffering the entire file in memory).
// it gives the file a random name using the global seed.
func (m *Manager) UploadFromUrl(ctx context.Context, url string) (newUrl string, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", errors.Wrap(err, "cannot build request")
	}
	resp, err := m.doer.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "cannot fetch image")
	}
	body := resp.Body
	defer body.Close()
	return m.Upload(ctx, randString(16), body)
}

// CreateBucket create a buckets in s3 server.
// TODO: handle acl
func (m *Manager) CreateBucket(ctx context.Context, name string) error {
	_, err := s3.New(m.sess).CreateBucket(&s3.CreateBucketInput{
		Bucket:    aws.String(name),
		GrantRead: aws.String("GrantRead"),
	})
	return err
}

func (m *Manager) otHandler() func(*request.Request) {
	tracer := m.tracer

	return func(r *request.Request) {
		var sp opentracing.Span

		ctx := r.Context()
		if ctx == nil || opentracing.IsGlobalTracerRegistered() {
			sp = tracer.StartSpan(r.Operation.Name)
		} else {
			sp, ctx = opentracing.StartSpanFromContextWithTracer(ctx, m.tracer, r.Operation.Name)
			r.SetContext(ctx)
		}
		ext.SpanKindRPCClient.Set(sp)
		ext.Component.Set(sp, "go-aws")
		ext.HTTPMethod.Set(sp, r.Operation.HTTPMethod)
		ext.HTTPUrl.Set(sp, r.HTTPRequest.URL.String())
		ext.PeerService.Set(sp, r.ClientInfo.ServiceName)

		_ = inject(tracer, sp, r.HTTPRequest.Header)

		r.Handlers.Complete.PushBack(func(req *request.Request) {
			if req.HTTPResponse != nil {
				ext.HTTPStatusCode.Set(sp, uint16(req.HTTPResponse.StatusCode))
			} else {
				ext.Error.Set(sp, true)
			}
			sp.Finish()
		})

		r.Handlers.Retry.PushBack(func(req *request.Request) {
			sp.LogFields(log.String("event", "retry"))
		})
	}
}

func inject(tracer opentracing.Tracer, span opentracing.Span, header http.Header) error {
	return tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(header))
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
