
<a name="v0.4.2"></a>
## [v0.4.2](https://github.com/DoNewsCode/core/compare/v0.4.1...v0.4.2) (2021-04-01)

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

