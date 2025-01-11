package openapiv3

import "github.com/google/gnostic/cmd/protoc-gen-openapi/generator"

type Option func(*options)

type options struct {
	conf func(c *generator.Configuration)
}

func WithConfig(f func(c *generator.Configuration)) Option {
	return func(opt *options) {
		opt.conf = f
	}
}
