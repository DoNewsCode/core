
<a name="v0.4.0.alpha1"></a>
## [v0.4.0.alpha1](https://github.com/DoNewsCode/core/compare/v0.3.0...v0.4.0.alpha1) (2021-03-13)

### ♻️ Code Refactoring

* config/env refactor ([#81](https://github.com/DoNewsCode/core/issues/81))
* move go kit and gin related package to seperate repo ([#74](https://github.com/DoNewsCode/core/issues/74))

### ✨ Features

* replace redis logger with Kitlog [#64](https://github.com/DoNewsCode/core/issues/64) ([#73](https://github.com/DoNewsCode/core/issues/73))

### 🐛 Bug Fixes

* don't panic when the database connection cannot be established at start up. ([#77](https://github.com/DoNewsCode/core/issues/77))
* fix example misspell ([#72](https://github.com/DoNewsCode/core/issues/72))
* **ginmw:** use c.FullPath() to calculate route matched ([#70](https://github.com/DoNewsCode/core/issues/70))


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

