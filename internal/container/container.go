package container

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	_ "github.com/danielgtaylor/huma/v2/formats/cbor" // CBOR format support for huma
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/samber/do"
)

type Options struct {
	Port int `default:"8888" help:"Port to listen on" short:"p"`
}

func New(_ humacli.Hooks, options *Options) *do.Injector {
	injector := do.New()

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("URL Shortener", "1.0.0"))

	do.ProvideValue(injector, router)
	do.ProvideValue(injector, api)
	do.ProvideValue(injector, options)

	return injector
}
