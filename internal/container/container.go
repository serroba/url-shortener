package container

import (
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	_ "github.com/danielgtaylor/huma/v2/formats/cbor" // CBOR format support for huma
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/samber/do"
	"github.com/serroba/web-demo-go/internal/handlers"
	"github.com/serroba/web-demo-go/internal/store"
)

type Options struct {
	Port int `default:"8888" help:"Port to listen on" short:"p"`
}

func New(_ humacli.Hooks, options *Options) *do.Injector {
	injector := do.New()

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("URL Shortener", "1.0.0"))

	urlStore := store.NewMemoryStore()
	baseURL := fmt.Sprintf("http://localhost:%d", options.Port)
	urlHandler := handlers.NewURLHandler(urlStore, baseURL)

	do.ProvideValue(injector, router)
	do.ProvideValue(injector, api)
	do.ProvideValue(injector, options)

	handlers.RegisterRoutes(api, urlHandler)

	return injector
}
