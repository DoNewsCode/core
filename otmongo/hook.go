package otmongo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.mongodb.org/mongo-driver/event"
)

type spanKey struct {
	ConnectionID string
	RequestID    int64
}

type monitor struct {
	sync.Mutex
	tracer opentracing.Tracer
	spans  map[spanKey]opentracing.Span
}

func (m *monitor) Started(ctx context.Context, evt *event.CommandStartedEvent) {
	hostname, port := peerInfo(evt)
	statement := evt.Command.String()

	span, _ := opentracing.StartSpanFromContextWithTracer(ctx, m.tracer, "mongodb.query")
	ext.DBType.Set(span, "mongo")
	ext.DBInstance.Set(span, evt.DatabaseName)
	ext.PeerHostname.Set(span, hostname)
	ext.PeerPort.Set(span, port)
	ext.DBStatement.Set(span, statement)
	ext.SpanKind.Set(span, ext.SpanKindEnum("client"))
	key := spanKey{
		ConnectionID: evt.ConnectionID,
		RequestID:    evt.RequestID,
	}
	m.Lock()
	m.spans[key] = span
	m.Unlock()
}

func (m *monitor) Succeeded(ctx context.Context, evt *event.CommandSucceededEvent) {
	m.Finished(&evt.CommandFinishedEvent, nil)
}

func (m *monitor) Failed(ctx context.Context, evt *event.CommandFailedEvent) {
	m.Finished(&evt.CommandFinishedEvent, fmt.Errorf("%s", evt.Failure))
}

func (m *monitor) Finished(evt *event.CommandFinishedEvent, err error) {
	key := spanKey{
		ConnectionID: evt.ConnectionID,
		RequestID:    evt.RequestID,
	}
	m.Lock()
	span, ok := m.spans[key]
	if ok {
		delete(m.spans, key)
	}
	m.Unlock()
	if !ok {
		return
	}
	if err != nil {
		ext.LogError(span, err)
	}
	span.Finish()
}

// NewMonitor creates a new mongodb event CommandMonitor.
func NewMonitor(tracer opentracing.Tracer) *event.CommandMonitor {
	m := &monitor{
		spans:  make(map[spanKey]opentracing.Span),
		tracer: tracer,
	}
	return &event.CommandMonitor{
		Started:   m.Started,
		Succeeded: m.Succeeded,
		Failed:    m.Failed,
	}
}

func peerInfo(evt *event.CommandStartedEvent) (hostname string, port uint16) {
	hostname = evt.ConnectionID
	strPort := "27017"
	if idx := strings.IndexByte(hostname, '['); idx >= 0 {
		hostname = hostname[:idx]
	}
	if idx := strings.IndexByte(hostname, ':'); idx >= 0 {
		strPort = hostname[idx+1:]
		hostname = hostname[:idx]
	}
	uintPort, _ := strconv.ParseUint(strPort, 10, 32)
	return hostname, uint16(uintPort)
}
