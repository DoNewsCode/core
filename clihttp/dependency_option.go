package clihttp

import "github.com/DoNewsCode/core/contract"

type providersOption struct {
	clientConstructor func(args ClientArgs) (contract.HttpDoer, error)
	clientOptions     []Option
}

type ClientArgs struct {
	Populator contract.DIPopulator
}

// ProvidersOptionFunc is the type of functional providersOption for Providers. Use this type to change how Providers work.
type ProvidersOptionFunc func(options *providersOption)

// WithDriverConstructor instructs the Providers to accept an alternative constructor for election driver.
func WithClientConstructor(f func(args ClientArgs) (contract.HttpDoer, error)) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.clientConstructor = f
	}
}

// WithClientOption instructs the Providers to accept additional options for the NewClient call, such as WithRequestLogThreshold.
func WithClientOption(options ...Option) ProvidersOptionFunc {
	return func(providerOption *providersOption) {
		providerOption.clientOptions = options
	}
}
