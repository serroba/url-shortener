package handlers

import (
	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(api huma.API, h *URLHandler) {
	huma.Post(api, "/shorten", h.CreateShortURL)
	huma.Get(api, "/{code}", h.RedirectToURL)
}
