
<a name="v0.4.0.alpha.2"></a>
## [v0.4.0.alpha.2](https://github.com/DoNewsCode/core/compare/v0.4.0.alpha.1...v0.4.0.alpha.2) (2021-03-13)


<a name="v0.4.0.alpha.1"></a>
## [v0.4.0.alpha.1](https://github.com/DoNewsCode/core/compare/v0.3.0...v0.4.0.alpha.1) (2021-03-13)

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

