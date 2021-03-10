module github.com/DoNewsCode/core

go 1.14

replace (
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
	go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20200520232829-54ba9589114f
	google.golang.org/grpc => google.golang.org/grpc v1.27.0
)

require (
	github.com/HdrHistogram/hdrhistogram-go v1.0.1 // indirect
	github.com/Reasno/ifilter v0.1.2
	github.com/aws/aws-sdk-go v1.37.16
	github.com/cockroachdb/datadriven v0.0.0-20200714090401-bf6692d28da5 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gabriel-vasile/mimetype v1.1.2
	github.com/ghodss/yaml v1.0.0
	github.com/gin-gonic/gin v1.6.3
	github.com/go-gormigrate/gormigrate/v2 v2.0.0
	github.com/go-kit/kit v0.10.0
	github.com/go-redis/redis/v8 v8.6.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.4.3
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.6 // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/knadh/koanf v0.15.0
	github.com/kr/pretty v0.2.1 // indirect
	github.com/mitchellh/mapstructure v1.4.1
	github.com/oklog/run v1.1.0
	github.com/opentracing-contrib/go-gin v0.0.0-20201220185307-1dd2273433a4
	github.com/opentracing-contrib/go-grpc v0.0.0-20210225150812-73cb765af46e
	github.com/opentracing-contrib/go-stdlib v1.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/xid v1.2.1
	github.com/segmentio/kafka-go v0.4.10
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/soheilhy/cmux v0.1.5-0.20210205191134-5ec6847320e5 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20200427203606-3cfed13b9966 // indirect
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.0+incompatible
	go.etcd.io/bbolt v1.3.5 // indirect
	go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
	go.mongodb.org/mongo-driver v1.4.6
	go.uber.org/atomic v1.7.0
	go.uber.org/dig v1.10.0
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/grpc v1.35.0
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gorm.io/driver/mysql v1.0.4
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.20.12
	sigs.k8s.io/yaml v1.2.0 // indirect
)
