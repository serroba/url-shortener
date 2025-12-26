package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/samber/do"
	"github.com/serroba/web-demo-go/internal/container"
)

func main() {
	cli := humacli.New(func(hooks humacli.Hooks, options *container.Options) {
		injector := do.New()
		do.ProvideValue(injector, options)
		container.RedisPackage(injector)
		container.RepositoryPackage(injector)
		container.RateLimitPackage(injector)
		container.HTTPPackage(injector)

		var server *http.Server

		hooks.OnStart(func() {
			router := do.MustInvoke[*chi.Mux](injector)

			// Invoke API to trigger route registration
			_ = do.MustInvoke[huma.API](injector)

			server = &http.Server{
				Addr:              fmt.Sprintf(":%d", options.Port),
				Handler:           router,
				ReadHeaderTimeout: 10 * time.Second,
			}
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				panic(err)
			}
		})

		hooks.OnStop(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if server != nil {
				_ = server.Shutdown(ctx)
			}

			_ = injector.Shutdown()
		})
	})

	cli.Run()
}
