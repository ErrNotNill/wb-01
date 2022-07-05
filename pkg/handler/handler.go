package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"github.com/patrickmn/go-cache"
	"wb-l0/pkg/repository"
)

type Handler struct {
	repo  *repository.Repository
	cache *cache.Cache
	nats  *nats.EncodedConn
}

func NewHandler(repo *repository.Repository, cache *cache.Cache, nats *nats.EncodedConn) *Handler {
	return &Handler{repo: repo, cache: cache, nats: nats}
}

func (h *Handler) InitRoutes() *chi.Mux {
	router := chi.NewRouter()

	h.natsHandler()

	router.Route("/", func(r chi.Router) {
		r.Get("/data", h.GetData)
		r.Post("/data_response", h.DataResponse)
	})

	return router
}
