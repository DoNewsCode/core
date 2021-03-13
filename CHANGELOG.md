
<a name="v0.3.0"></a>
## [v0.3.0](https://github.com/DoNewsCode/core/compare/v0.2.0...v0.3.0) (2021-03-10)

### ♻️ Code Refactoring

* **otes:** optimize logger ([#68](https://github.com/DoNewsCode/core/issues/68))

### ✨ Features

* Saga ([#63](https://github.com/DoNewsCode/core/issues/63))
* **es:** Add otes package ([#61](https://github.com/DoNewsCode/core/issues/61))
* **kitmw:** limit maximum concurrency ([#67](https://github.com/DoNewsCode/core/issues/67))

### 🏗 Chore

* add core-starter link ([#66](https://github.com/DoNewsCode/core/issues/66))

### 🐛 Bug Fixes

* **ots3:** investigate race condition ([#62](https://github.com/DoNewsCode/core/issues/62))
* **ots3:** missing trace in ots3

### 📚 Documentation

* fix grammar
* fix grammar
* add package level doc for leaderetcd and leaderredis

### 🚦 Test

* turn off flaky test

### Pull Requests

* Merge pull request [#58](https://github.com/DoNewsCode/core/issues/58) from DoNewsCode/Reasno-patch-1


<a name="v0.2.0"></a>
## [v0.2.0](https://github.com/DoNewsCode/core/compare/v0.1.1...v0.2.0) (2021-03-02)

### ✨ Features

* **leader:** add leader election package. ([#56](https://github.com/DoNewsCode/core/issues/56))

### 📚 Documentation

* fix cases
* add tutorial.md chapter 3 and 4
* add tutorial.md chapter 3 and 4
* add tutorial.md chapter 2
* add tutorial.md

### 🚦 Test

* **kitkafka:** improve test coverage in kitkafka package ([#54](https://github.com/DoNewsCode/core/issues/54))


<a name="v0.1.1"></a>
## [v0.1.1](https://github.com/DoNewsCode/core/compare/v0.1.0...v0.1.1) (2021-02-24)

### 🐛 Bug Fixes

* **c.go:** fix LoggerProvider note
* **config:** add synchronization ([#52](https://github.com/DoNewsCode/core/issues/52))
* **core:** mysql creates trouble at bootup. Use sqlite as the default database. Only connect to mysql when opt in.
* **core:** unstable test
* **deps:** update module github.com/aws/aws-sdk-go to v1.37.16
* **watcher:** remove default signal watch, use ctx ([#49](https://github.com/DoNewsCode/core/issues/49))

### 📚 Documentation

* add CONTRIBUTING.md ([#47](https://github.com/DoNewsCode/core/issues/47))
* **readme:** clarify module usage ([#53](https://github.com/DoNewsCode/core/issues/53))

### 🚦 Test

* move all integration tests under -integration tag ([#46](https://github.com/DoNewsCode/core/issues/46))
* **config:** try another approach to make tests less flaky
* **config:** improve test coverage in config package
* **config:** investigate config test failures
* **config:** investigate config test failures
* **config:** mark config tests as parallel
* **config:** fix flaky test
* **queue:** fix invalid tests
* **queue:** improve test stability ([#48](https://github.com/DoNewsCode/core/issues/48))

### Pull Requests

* Merge pull request [#39](https://github.com/DoNewsCode/core/issues/39) from GGXXLL/master
* Merge pull request [#38](https://github.com/DoNewsCode/core/issues/38) from DoNewsCode/renovate/github.com-aws-aws-sdk-go-1.x


<a name="v0.1.0"></a>
## v0.1.0 (2021-02-22)

### Ci

* add codecov.yml
* upload codecoverage
* fix kafka port
* fix kafka port
* fix kafka port
* fix kafka port
* fix kafka port
* fix kafka port
* fix kafka port
* add kafka test
* add kafka test
* add kafka test
* add redis service

### Refactory

* change kafka factory to kafka reader factory and kafka writer factory

### ♻️ Code Refactoring

* change server side encoding logic
* rename HealthCheck to HealthCheckModule for consistency.
* normalize naming convention
* change module name to core
* optimize queue info
* support HttpStatusCodeFunc
* change config mechanism
* remove unused function
* change interfaces
* change interfaces
* add kafka examples
* **core:** change names
* **core:** change names
* **core:** change README.md
* **core:** change names

### ✨ Features

* finish package log
* optimize queue
* add spanlogger
* add redis provider
* add redis provider
* add observability package
* add event system

### 🏗 Chore

* run go mod tidy
* **deps:** update module aws/aws-sdk-go to v1.37.3
* **deps:** update golang.org/x/sync commit hash to 09787c9
* **deps:** update module aws/aws-sdk-go to v1.37.12
* **deps:** update module gogo/protobuf to v1.3.2
* **deps:** update module go-redis/redis/v8 to v8.5.0
* **deps:** update module reasno/ifilter to v0.1.2
* **deps:** update module spf13/cobra to v1
* **deps:** update module aws/aws-sdk-go to v1.37.13
* **deps:** update module aws/aws-sdk-go to v1.37.2
* **deps:** update module aws/aws-sdk-go to v1.37.1
* **deps:** update module aws/aws-sdk-go to v1.37.0
* **deps:** update module opentracing/opentracing-go to v1.2.0
* **deps:** update module knadh/koanf to v0.15.0
* **deps:** update module gogo/protobuf to v1.3.2
* **deps:** update module prometheus/client_golang to v1.9.0

### 🐛 Bug Fixes

* test
* change default to sqlite
* rollback to mysql
* make ineffassign happy
* grammar errors
* proper cancel test
* export observability config
* export observability config
* make ineffassign happy
* test
* test
* test
* generic fixes heading to v0.1.0 release.
* remove duplicate logs
* **deps:** update module github.com/segmentio/kafka-go to v0.4.10
* **deps:** update module github.com/spf13/cobra to v1.1.3
* **deps:** update golang.org/x/sync commit hash to 036812b ([#36](https://github.com/DoNewsCode/core/issues/36))
* **deps:** update module github.com/go-redis/redis/v8 to v8.6.0 ([#33](https://github.com/DoNewsCode/core/issues/33))

### 📚 Documentation

* add link to skeleton
* link to go doc
* link to go doc
* fix badge links
* add badge
* add badge
* add badge
* fix wrong context in README.md
* document package event
* document package ginmw
* document package ginmw
* document package key
* remove redundant doc files
* write kafka doc
* document ots3 package
* add lots of doc in queue package
* document unierr

### 🚦 Test

* test custom printer
* improve test coverage
* fix module test
* fix module test
* fix module test
* cover large request
* cover large request
* cover large request
* cover large request
* with level
* with level
* with level
* add clihttp test
* fix go tests
* finish otgorm tests
* reduce flake
* reduce flake
* fix module test
* fix module test
* remove test metrix for now
* remove test metrix for now
* fix broken build
* fix broken build
* add mongodb
* add mongodb
* add test to s3 module.go
* fix module test
* fix module test
* fix module test
* **ginmw:** test metrics middleware
* **ginmw:** test metrics middleware
* **ginmw:** test metrics middleware
* **ginmw:** test log middleware

### Pull Requests

* Merge pull request [#34](https://github.com/DoNewsCode/core/issues/34) from DoNewsCode/renovate/github.com-spf13-cobra-1.x
* Merge pull request [#35](https://github.com/DoNewsCode/core/issues/35) from DoNewsCode/renovate/github.com-segmentio-kafka-go-0.x
* Merge pull request [#28](https://github.com/DoNewsCode/core/issues/28) from DoNewsCode/renovate/aws-aws-sdk-go-1.x
* Merge pull request [#23](https://github.com/DoNewsCode/core/issues/23) from DoNewsCode/renovate/aws-aws-sdk-go-1.x
* Merge pull request [#25](https://github.com/DoNewsCode/core/issues/25) from DoNewsCode/renovate/gogo-protobuf-1.x
* Merge pull request [#26](https://github.com/DoNewsCode/core/issues/26) from DoNewsCode/renovate/golang.org-x-sync-digest
* Merge pull request [#24](https://github.com/DoNewsCode/core/issues/24) from DoNewsCode/renovate/go-redis-redis-v8-8.x
* Merge pull request [#21](https://github.com/DoNewsCode/core/issues/21) from DoNewsCode/renovate/github.com-reasno-ifilter-0.x
* Merge pull request [#18](https://github.com/DoNewsCode/core/issues/18) from DoNewsCode/renovate/github.com-spf13-cobra-1.x
* Merge pull request [#16](https://github.com/DoNewsCode/core/issues/16) from DoNewsCode/renovate/github.com-aws-aws-sdk-go-1.x
* Merge pull request [#15](https://github.com/DoNewsCode/core/issues/15) from DoNewsCode/renovate/github.com-aws-aws-sdk-go-1.x
* Merge pull request [#14](https://github.com/DoNewsCode/core/issues/14) from DoNewsCode/renovate/github.com-aws-aws-sdk-go-1.x
* Merge pull request [#13](https://github.com/DoNewsCode/core/issues/13) from DoNewsCode/renovate/github.com-gogo-protobuf-1.x
* Merge pull request [#6](https://github.com/DoNewsCode/core/issues/6) from DoNewsCode/renovate/github.com-knadh-koanf-0.x
* Merge pull request [#8](https://github.com/DoNewsCode/core/issues/8) from DoNewsCode/renovate/github.com-opentracing-opentracing-go-1.x
* Merge pull request [#11](https://github.com/DoNewsCode/core/issues/11) from DoNewsCode/renovate/github.com-aws-aws-sdk-go-1.x
* Merge pull request [#9](https://github.com/DoNewsCode/core/issues/9) from DoNewsCode/renovate/github.com-prometheus-client_golang-1.x
* Merge pull request [#12](https://github.com/DoNewsCode/core/issues/12) from DoNewsCode/renovate/gorm.io-driver-mysql-1.x
* Merge pull request [#5](https://github.com/DoNewsCode/core/issues/5) from DoNewsCode/renovate/gorm.io-gorm-1.x
* Merge pull request [#2](https://github.com/DoNewsCode/core/issues/2) from DoNewsCode/renovate/github.com-aws-aws-sdk-go-1.x
* Merge pull request [#3](https://github.com/DoNewsCode/core/issues/3) from DoNewsCode/renovate/google.golang.org-grpc-1.x
* Merge pull request [#1](https://github.com/DoNewsCode/core/issues/1) from DoNewsCode/renovate/configure

