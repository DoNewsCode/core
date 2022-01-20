
<a name="v0.12.0"></a>
## [v0.12.0](https://github.com/DoNewsCode/core/compare/v0.11.1...v0.12.0) (2022-01-20)

### ♻️ Code Refactoring

* **cron:** replace cron implementation. ([#226](https://github.com/DoNewsCode/core/issues/226)) (@[谷溪](guxi99@gmail.com))

### ✨ Features

* add module field injection ([#234](https://github.com/DoNewsCode/core/issues/234)) (@[谷溪](guxi99@gmail.com))
* **dag:** add dag package. ([#222](https://github.com/DoNewsCode/core/issues/222)) (@[谷溪](guxi99@gmail.com))
* **logging:** add WithBaggage ([#225](https://github.com/DoNewsCode/core/issues/225)) (@[谷溪](guxi99@gmail.com))
* **observability:** capture transport status ([#224](https://github.com/DoNewsCode/core/issues/224)) (@[谷溪](guxi99@gmail.com))

### 🐛 Bug Fixes

* avoid extra logger wraps  ([#232](https://github.com/DoNewsCode/core/issues/232)) (@[谷溪](guxi99@gmail.com))
* prepend dig.LocationForPC instead of append ([#231](https://github.com/DoNewsCode/core/issues/231)) (@[谷溪](guxi99@gmail.com))
* **dag:** validate edges in AddEdges before adding them. ([#223](https://github.com/DoNewsCode/core/issues/223)) (@[谷溪](guxi99@gmail.com))
* **deps:** upgrade github.com/go-redis/redis/v8 to v8.11.4 ([#221](https://github.com/DoNewsCode/core/issues/221)) (@[江湖大牛](nfangxu@gmail.com))
* **factory:** on reload, close the connection right away. ([#235](https://github.com/DoNewsCode/core/issues/235)) (@[谷溪](guxi99@gmail.com))
* **observability:** data races ([#227](https://github.com/DoNewsCode/core/issues/227)) (@[谷溪](guxi99@gmail.com))

### BREAKING CHANGE


most Observe() methods now take a time.Duration instead of float64.

* wip: new cron package

* refactor(cron): remove cronopts, add cron

This PR replaces the cron implementation.

the new cron package github/DoNewsCode/core/cron is not compatible with github.com/robfig/cron/v3. See examples for how to migrate.

* refactor(cron): deprecate cronopts, add cron

This PR replaces the cron implementation.

the new cron package github/DoNewsCode/core/cron is not compatible with github.com/robfig/cron/v3. See examples for how to migrate.

* fix: delayed time calculation

* refactor: change job middleware to job options

* fix: use time.Since

* fix: inconsistent labels

* fix: race

* fix: race

* fix: race

* fix: race

* fix: race

* fix: rename JobOptions to JobOption

* refactor: Reduce the API interface of Container

* refactor: Reduce the API interface of Container

* refactor: Reduce the API interface of Container

* refactor: Reduce the API interface of Container

* fix: minor adjustments of docs,imports

most Observe() methods now take a time.Duration instead of float64.


<a name="v0.11.1"></a>
## [v0.11.1](https://github.com/DoNewsCode/core/compare/v0.11.0...v0.11.1) (2022-01-06)

### ⚡️ Performance

* logger performance optimization ([#219](https://github.com/DoNewsCode/core/issues/219)) (@[谷溪](guxi99@gmail.com))

### ✨ Features

* add GetOrInjectBaggage ([#217](https://github.com/DoNewsCode/core/issues/217)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))


<a name="v0.11.0"></a>
## [v0.11.0](https://github.com/DoNewsCode/core/compare/v0.10.4...v0.11.0) (2021-12-27)

### ♻️ Code Refactoring

* **observability:** avoid panics caused by missing labels ([#213](https://github.com/DoNewsCode/core/issues/213)) (@[谷溪](guxi99@gmail.com))

### ✨ Features

* add cronjob metrics ([#215](https://github.com/DoNewsCode/core/issues/215)) (@[谷溪](guxi99@gmail.com))
* simple span error log and gofumpt ([#214](https://github.com/DoNewsCode/core/issues/214)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))


<a name="v0.10.4"></a>
## [v0.10.4](https://github.com/DoNewsCode/core/compare/v0.10.3...v0.10.4) (2021-11-18)

### 🐛 Bug Fixes

* **clihttp:** log errors ([#212](https://github.com/DoNewsCode/core/issues/212)) (@[谷溪](guxi99@gmail.com))
* **logging:** inconsistency between go kit Logger and spanLogger ([#211](https://github.com/DoNewsCode/core/issues/211)) (@[谷溪](guxi99@gmail.com))


<a name="v0.10.3"></a>
## [v0.10.3](https://github.com/DoNewsCode/core/compare/v0.10.2...v0.10.3) (2021-10-28)

### 🐛 Bug Fixes

* use "tag" as the key to identify logs (@[Reasno](guxi99@gmail.com))
* spanlogger should support log.Valuer ([#209](https://github.com/DoNewsCode/core/issues/209)) (@[谷溪](guxi99@gmail.com))
* **clihttp:** add the missing Providers function ([#210](https://github.com/DoNewsCode/core/issues/210)) (@[谷溪](guxi99@gmail.com))


<a name="v0.10.2"></a>
## [v0.10.2](https://github.com/DoNewsCode/core/compare/v0.10.1...v0.10.2) (2021-10-21)

### 🐛 Bug Fixes

* **logging:** nil pointer ([#205](https://github.com/DoNewsCode/core/issues/205)) (@[谷溪](guxi99@gmail.com))
* **serve:** signal group couldn't be canceled ([#208](https://github.com/DoNewsCode/core/issues/208)) (@[谷溪](guxi99@gmail.com))
* **srvhttp:** RequestDurationSeconds shouldn't panic when missing labels ([#207](https://github.com/DoNewsCode/core/issues/207)) (@[谷溪](guxi99@gmail.com))


<a name="v0.10.1"></a>
## [v0.10.1](https://github.com/DoNewsCode/core/compare/v0.10.0...v0.10.1) (2021-09-30)

### 🐛 Bug Fixes

* lint error ([#204](https://github.com/DoNewsCode/core/issues/204)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* metrics default interval ([#203](https://github.com/DoNewsCode/core/issues/203)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* **logging:** nil pointer (@[Reasno](guxi99@gmail.com))
* **logging:** nil pointer (@[Reasno](guxi99@gmail.com))


<a name="v0.10.0"></a>
## [v0.10.0](https://github.com/DoNewsCode/core/compare/v0.9.2...v0.10.0) (2021-09-27)

### ♻️ Code Refactoring

* remove DIContainer abstraction ([#196](https://github.com/DoNewsCode/core/issues/196)) (@[谷溪](guxi99@gmail.com))

### ✨ Features

* add WithoutCancel ([#200](https://github.com/DoNewsCode/core/issues/200)) (@[谷溪](guxi99@gmail.com))
* make reload opt-in ([#195](https://github.com/DoNewsCode/core/issues/195)) (@[谷溪](guxi99@gmail.com))

### 🐛 Bug Fixes

* **config:** watch etcd from last revision ([#199](https://github.com/DoNewsCode/core/issues/199)) (@[谷溪](guxi99@gmail.com))

### BREAKING CHANGE


reload used to be enabled by default

* refactor: add WithReload to otetcd

* refactor: add WithReload to otkafka

* refactor: add WithReload to ots3

* refactor: add WithReload to otgorm

* refactor: add WithReload to otkafka

* refactor: add WithReload to otmongo

* fix: tests in otkafka

* fix: tests in otmongo

* fix: tests in otredis

* fix: tests in otredis


<a name="v0.9.2"></a>
## [v0.9.2](https://github.com/DoNewsCode/core/compare/v0.9.1...v0.9.2) (2021-09-15)

### 🐛 Bug Fixes

* panic when signal sent twice ([#193](https://github.com/DoNewsCode/core/issues/193)) (@[谷溪](guxi99@gmail.com))
* golangci-lint complaint (@[Reasno](guxi99@gmail.com))
* don't close connection right away after reload (@[Reasno](guxi99@gmail.com))


<a name="v0.9.1"></a>
## [v0.9.1](https://github.com/DoNewsCode/core/compare/v0.9.0...v0.9.1) (2021-09-13)

### ♻️ Code Refactoring

* reduce clustering (@[Reasno](guxi99@gmail.com))

### 🐛 Bug Fixes

* prefer driver from DI (@[Reasno](guxi99@gmail.com))
* prefer driver from DI ([#192](https://github.com/DoNewsCode/core/issues/192)) (@[谷溪](guxi99@gmail.com))


<a name="v0.9.0"></a>
## [v0.9.0](https://github.com/DoNewsCode/core/compare/v0.8.0...v0.9.0) (2021-09-10)

### ♻️ Code Refactoring

* change collector mechanism ([#189](https://github.com/DoNewsCode/core/issues/189)) (@[谷溪](guxi99@gmail.com))
* remove dtx and rename to core-dtx ([#181](https://github.com/DoNewsCode/core/issues/181)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* make config and log easier to use and extend ([#177](https://github.com/DoNewsCode/core/issues/177)) (@[谷溪](guxi99@gmail.com))
* **container:** remove ifilter dependency. ([#183](https://github.com/DoNewsCode/core/issues/183)) (@[谷溪](guxi99@gmail.com))

### ✨ Features

* expand the API of SyncDispatcher ([#191](https://github.com/DoNewsCode/core/issues/191)) (@[谷溪](guxi99@gmail.com))
* introduce provider options ([#190](https://github.com/DoNewsCode/core/issues/190)) (@[谷溪](guxi99@gmail.com))
* add ctxmeta package ([#188](https://github.com/DoNewsCode/core/issues/188)) (@[谷溪](guxi99@gmail.com))
* make metrics struct type safe ([#185](https://github.com/DoNewsCode/core/issues/185)) (@[谷溪](guxi99@gmail.com))
* **clihttp:** limit max length a client can read from body ([#186](https://github.com/DoNewsCode/core/issues/186)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))

### 🐛 Bug Fixes

* remove unused factoryOut (@[Reasno](guxi99@gmail.com))
* rename ThreeStats to AggStats ([#184](https://github.com/DoNewsCode/core/issues/184)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* **container:** shutdown the modules in the reversed order of registr… ([#187](https://github.com/DoNewsCode/core/issues/187)) (@[谷溪](guxi99@gmail.com))


<a name="v0.8.0"></a>
## [v0.8.0](https://github.com/DoNewsCode/core/compare/v0.7.3...v0.8.0) (2021-08-17)

### ♻️ Code Refactoring

* rename package remote to package etcd ([#175](https://github.com/DoNewsCode/core/issues/175)) (@[谷溪](guxi99@gmail.com))
* move and rename ([#170](https://github.com/DoNewsCode/core/issues/170)) (@[谷溪](guxi99@gmail.com))
* change event system ([#165](https://github.com/DoNewsCode/core/issues/165)) (@[谷溪](guxi99@gmail.com))
* move otkafka/processor out of core, rename to core-processor as another package ([#156](https://github.com/DoNewsCode/core/issues/156)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))

### ✨ Features

* config.Duration implement TextMarshaller (close [#164](https://github.com/DoNewsCode/core/issues/164)) ([#166](https://github.com/DoNewsCode/core/issues/166)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* add codec package ([#161](https://github.com/DoNewsCode/core/issues/161)) (@[谷溪](guxi99@gmail.com))
* **config:** add Validators to ExportedConfigs and "config verify" command. ([#154](https://github.com/DoNewsCode/core/issues/154)) (@[谷溪](guxi99@gmail.com))

### 🐛 Bug Fixes

* clarify env usage ([#174](https://github.com/DoNewsCode/core/issues/174)) (@[谷溪](guxi99@gmail.com))
* optimize di output ([#171](https://github.com/DoNewsCode/core/issues/171)) (@[谷溪](guxi99@gmail.com))
* when elasticsearch server is not up, the elasticsearch client should be constructed normally. ([#167](https://github.com/DoNewsCode/core/issues/167)) (@[谷溪](guxi99@gmail.com))
* golangci-lint run (@[Reasno](guxi99@gmail.com))
* fix remote test (close [#160](https://github.com/DoNewsCode/core/issues/160)) ([#162](https://github.com/DoNewsCode/core/issues/162)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* **core:** remove WithRemote (@[reasno](guxi99@gmail.com))

### BREAKING CHANGE


package remote no longer exists.

WithRemote option removed.

the event interface is changed. queue package is removed.


<a name="v0.7.3"></a>
## [v0.7.3](https://github.com/DoNewsCode/core/compare/v0.7.2...v0.7.3) (2021-07-19)

### 🐛 Bug Fixes

* row close ([#152](https://github.com/DoNewsCode/core/issues/152)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))


<a name="v0.7.2"></a>
## [v0.7.2](https://github.com/DoNewsCode/core/compare/v0.7.1...v0.7.2) (2021-07-13)

### 🐛 Bug Fixes

* config.Duration support int for yaml ([#149](https://github.com/DoNewsCode/core/issues/149)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* generic metrics should not have a namespace ([#148](https://github.com/DoNewsCode/core/issues/148)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))


<a name="v0.7.1"></a>
## [v0.7.1](https://github.com/DoNewsCode/core/compare/v0.7.0...v0.7.1) (2021-07-02)

### 🐛 Bug Fixes

* provide the correct env value ([#146](https://github.com/DoNewsCode/core/issues/146)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* **clihttp:** fix context passing in http client ([#147](https://github.com/DoNewsCode/core/issues/147)) (@[谷溪](guxi99@gmail.com))


<a name="v0.7.0"></a>
## [v0.7.0](https://github.com/DoNewsCode/core/compare/v0.6.1...v0.7.0) (2021-06-11)

### ✨ Features

* Serve run ([#137](https://github.com/DoNewsCode/core/issues/137)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* Add more gorm-driver ([#138](https://github.com/DoNewsCode/core/issues/138)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* **config:** support configuration ReloadedEvent ([#142](https://github.com/DoNewsCode/core/issues/142)) (@[谷溪](guxi99@gmail.com))
* **di:** support config reloading at factory level ([#143](https://github.com/DoNewsCode/core/issues/143)) (@[谷溪](guxi99@gmail.com))

### 🐛 Bug Fixes

* Incorrect loading of the pprof.Index ([#145](https://github.com/DoNewsCode/core/issues/145)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* **tests:** allow test to pass when no env is provided. ([#141](https://github.com/DoNewsCode/core/issues/141)) (@[谷溪](guxi99@gmail.com))


<a name="v0.6.1"></a>
## [v0.6.1](https://github.com/DoNewsCode/core/compare/v0.6.0...v0.6.1) (2021-05-25)

### 🐛 Bug Fixes

* otkafka processor error (close [#135](https://github.com/DoNewsCode/core/issues/135)) ([#136](https://github.com/DoNewsCode/core/issues/136)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))


<a name="v0.6.0"></a>
## [v0.6.0](https://github.com/DoNewsCode/core/compare/v0.5.0...v0.6.0) (2021-05-19)

### ✨ Features

* add otkafka processor ([#134](https://github.com/DoNewsCode/core/issues/134)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* Add kafka metrics ([#130](https://github.com/DoNewsCode/core/issues/130)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* cluster address in env (close [#131](https://github.com/DoNewsCode/core/issues/131)) ([#133](https://github.com/DoNewsCode/core/issues/133)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))


<a name="v0.5.0"></a>
## [v0.5.0](https://github.com/DoNewsCode/core/compare/v0.4.2...v0.5.0) (2021-04-29)

### ✨ Features

* Add redis connection metrics (close [#127](https://github.com/DoNewsCode/core/issues/127)) ([#128](https://github.com/DoNewsCode/core/issues/128)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* add mysql connection metrics ([#118](https://github.com/DoNewsCode/core/issues/118)) (@[谷溪](guxi99@gmail.com))
* **unierr:** allow nil errors (close [#125](https://github.com/DoNewsCode/core/issues/125)) ([#126](https://github.com/DoNewsCode/core/issues/126)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))

### 🐛 Bug Fixes

* cases (@[谷溪](guxi99@gmail.com))


<a name="v0.4.2"></a>
## [v0.4.2](https://github.com/DoNewsCode/core/compare/v0.4.1...v0.4.2) (2021-04-06)

### ♻️ Code Refactoring

* Valid -> IsZero ([#119](https://github.com/DoNewsCode/core/issues/119)) (@[谷溪](guxi99@gmail.com))
* **sagas:** prettify sagas config ([#116](https://github.com/DoNewsCode/core/issues/116)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))

### 🐛 Bug Fixes

* config.Duration Unmarshal with koanf ([#114](https://github.com/DoNewsCode/core/issues/114)) ([#115](https://github.com/DoNewsCode/core/issues/115)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* **otetcd:** change configuration to use config.Duration type. ([#112](https://github.com/DoNewsCode/core/issues/112)) (@[谷溪](guxi99@gmail.com))
* **sagas:** change configuration to use config.Duration type. ([#111](https://github.com/DoNewsCode/core/issues/111)) (@[谷溪](guxi99@gmail.com))

### BREAKING CHANGE


the new sagas configuration is not backward compatible.

* doc: unified tag format

the new otetcd configuration is not backward compatible.


<a name="v0.4.1"></a>
## [v0.4.1](https://github.com/DoNewsCode/core/compare/v0.4.0...v0.4.1) (2021-03-25)

### 🐛 Bug Fixes

* sort otes configuration ([#108](https://github.com/DoNewsCode/core/issues/108)) (@[谷溪](guxi99@gmail.com))
* sort redis configuration ([#107](https://github.com/DoNewsCode/core/issues/107)) (@[谷溪](guxi99@gmail.com))


<a name="v0.4.0"></a>
## [v0.4.0](https://github.com/DoNewsCode/core/compare/v0.4.0-alpha.2...v0.4.0) (2021-03-18)

### ✨ Features

* **sagas:** add mysql store ([#100](https://github.com/DoNewsCode/core/issues/100)) (@[谷溪](guxi99@gmail.com))

### 🐛 Bug Fixes

* logging logfmt use sync-logger ([#102](https://github.com/DoNewsCode/core/issues/102)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))


<a name="v0.4.0-alpha.2"></a>
## [v0.4.0-alpha.2](https://github.com/DoNewsCode/core/compare/v0.4.0-alpha.1...v0.4.0-alpha.2) (2021-03-17)

### ✨ Features

* add CronLogAdapter [#88](https://github.com/DoNewsCode/core/issues/88) ([#96](https://github.com/DoNewsCode/core/issues/96)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* add a configuration entry to disable servers. ([#93](https://github.com/DoNewsCode/core/issues/93)) (@[谷溪](guxi99@gmail.com))
* add server events ([#86](https://github.com/DoNewsCode/core/issues/86)) (@[谷溪](guxi99@gmail.com))
* **otes:** allow users to specify extra options ([#97](https://github.com/DoNewsCode/core/issues/97)) (@[谷溪](guxi99@gmail.com))

### 🐛 Bug Fixes

* change default_config redis DB to db ([#95](https://github.com/DoNewsCode/core/issues/95)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* s3 config was not exported correctly ([#89](https://github.com/DoNewsCode/core/issues/89)) (@[谷溪](guxi99@gmail.com))
* correctly export CorrelationID field ([#87](https://github.com/DoNewsCode/core/issues/87)) (@[谷溪](guxi99@gmail.com))


<a name="v0.4.0-alpha.1"></a>
## [v0.4.0-alpha.1](https://github.com/DoNewsCode/core/compare/v0.3.0...v0.4.0-alpha.1) (2021-03-13)

### ♻️ Code Refactoring

* config/env refactor ([#81](https://github.com/DoNewsCode/core/issues/81)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))
* move go kit and gin related package to seperate repo ([#74](https://github.com/DoNewsCode/core/issues/74)) (@[谷溪](guxi99@gmail.com))

### ✨ Features

* replace redis logger with Kitlog [#64](https://github.com/DoNewsCode/core/issues/64) ([#73](https://github.com/DoNewsCode/core/issues/73)) (@[Trock](35254251+GGXXLL@users.noreply.github.com))

### 🐛 Bug Fixes

* don't panic when the database connection cannot be established at start up. ([#77](https://github.com/DoNewsCode/core/issues/77)) (@[谷溪](guxi99@gmail.com))
* fix example misspell ([#72](https://github.com/DoNewsCode/core/issues/72)) (@[另维64](lingwei0604@gmail.com))
* **ginmw:** use c.FullPath() to calculate route matched ([#70](https://github.com/DoNewsCode/core/issues/70)) (@[谷溪](guxi99@gmail.com))


<a name="v0.3.0"></a>
## [v0.3.0](https://github.com/DoNewsCode/core/compare/v0.2.0...v0.3.0) (2021-03-10)

### ♻️ Code Refactoring

* **otes:** optimize logger ([#68](https://github.com/DoNewsCode/core/issues/68)) (@[谷溪](guxi99@gmail.com))

### ✨ Features

* Saga ([#63](https://github.com/DoNewsCode/core/issues/63)) (@[谷溪](guxi99@gmail.com))
* **es:** Add otes package ([#61](https://github.com/DoNewsCode/core/issues/61)) (@[另维64](1142674342@qq.com))
* **kitmw:** limit maximum concurrency ([#67](https://github.com/DoNewsCode/core/issues/67)) (@[谷溪](guxi99@gmail.com))

### 🐛 Bug Fixes

* **ots3:** investigate race condition ([#62](https://github.com/DoNewsCode/core/issues/62)) (@[谷溪](guxi99@gmail.com))
* **ots3:** missing trace in ots3 (@[Reasno](guxi99@gmail.com))

### Pull Requests

* Merge pull request [#58](https://github.com/DoNewsCode/core/issues/58) from DoNewsCode/Reasno-patch-1


<a name="v0.2.0"></a>
## [v0.2.0](https://github.com/DoNewsCode/core/compare/v0.1.1...v0.2.0) (2021-03-02)

### ✨ Features

* **leader:** add leader election package. ([#56](https://github.com/DoNewsCode/core/issues/56)) (@[谷溪](guxi99@gmail.com))

