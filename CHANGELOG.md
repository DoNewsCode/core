
<a name="v0.3.0"></a>
## [v0.3.0](https://github.com/DoNewsCode/core/compare/v0.2.0...v0.3.0) (2021-03-10)

### ♻️ Code Refactoring

* **otes:** optimize logger ([#68](https://github.com/DoNewsCode/core/issues/68))

### ✨ Features

* Saga ([#63](https://github.com/DoNewsCode/core/issues/63))
* **es:** Add otes package ([#61](https://github.com/DoNewsCode/core/issues/61))
* **kitmw:** limit maximum concurrency ([#67](https://github.com/DoNewsCode/core/issues/67))

### 🐛 Bug Fixes

* **ots3:** investigate race condition ([#62](https://github.com/DoNewsCode/core/issues/62))
* **ots3:** missing trace in ots3

### Pull Requests

* Merge pull request [#58](https://github.com/DoNewsCode/core/issues/58) from DoNewsCode/Reasno-patch-1


<a name="v0.2.0"></a>
## [v0.2.0](https://github.com/DoNewsCode/core/compare/v0.1.1...v0.2.0) (2021-03-02)

### ✨ Features

* **leader:** add leader election package. ([#56](https://github.com/DoNewsCode/core/issues/56))


<a name="v0.1.1"></a>
## [v0.1.1](https://github.com/DoNewsCode/core/compare/v0.1.0...v0.1.1) (2021-02-24)

### 🐛 Bug Fixes

* **c.go:** fix LoggerProvider note
* **config:** add synchronization ([#52](https://github.com/DoNewsCode/core/issues/52))
* **core:** mysql creates trouble at bootup. Use sqlite as the default database. Only connect to mysql when opt in.
* **core:** unstable test
* **deps:** update module github.com/aws/aws-sdk-go to v1.37.16
* **watcher:** remove default signal watch, use ctx ([#49](https://github.com/DoNewsCode/core/issues/49))

### Pull Requests

* Merge pull request [#39](https://github.com/DoNewsCode/core/issues/39) from GGXXLL/master
* Merge pull request [#38](https://github.com/DoNewsCode/core/issues/38) from DoNewsCode/renovate/github.com-aws-aws-sdk-go-1.x


<a name="v0.1.0"></a>
## v0.1.0 (2021-02-22)

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

