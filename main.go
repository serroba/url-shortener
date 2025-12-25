package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/samber/do"
	"github.com/serroba/web-demo-go/internal/container"
)

func main() {
	cli := humacli.New(func(hooks humacli.Hooks, options *container.Options) {
		injector := container.New(hooks, options)

		hooks.OnStart(func() {
			router := do.MustInvoke[*chi.Mux](injector)
			opts := do.MustInvoke[*container.Options](injector)

			server := &http.Server{
				Addr:              fmt.Sprintf(":%d", opts.Port),
				Handler:           router,
				ReadHeaderTimeout: 10 * time.Second,
			}
			if err := server.ListenAndServe(); err != nil {
				panic(err)
			}
		})
	})

	cli.Run()
}
