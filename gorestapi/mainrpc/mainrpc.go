package mainrpc

import (
	"github.com/go-chi/chi/v5"
	"github.com/jqrd/gorestapi-mongo/gorestapi"
	"go.uber.org/zap"
)

// Server is the API web server
type Server struct {
	logger *zap.SugaredLogger
	router chi.Router
	store  gorestapi.DataStore
}

// Set up the API listener
func Setup(router chi.Router, store gorestapi.DataStore) error {

	s := &Server{
		logger: zap.S().With("package", "thingrpc"),
		router: router,
		store:  store,
	}

	things := s.NewThingsAPI()

	// Base Functions
	s.router.Route("/api", func(r chi.Router) {
		r.Post("/things", things.Create())
		r.Get("/things/{id}", things.GetByID())
		r.Delete("/things/{id}", things.DeleteByID())
		r.Get("/things", things.Find())

		// TODO widgets
	})

	return nil

}
