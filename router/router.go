package router

import (
	"github.com/Seymour-creates/budget-server/handlers"
	"github.com/Seymour-creates/budget-server/utils"
	"net/http"
)

type Server struct {
	mux *http.ServeMux
}

func NewServer() *Server {
	server := &Server{
		mux: http.NewServeMux(),
	}
	server.registerRoutes()
	return server
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/add_expense", utils.ErrorHandler(handlers.AddExpense))
	s.mux.HandleFunc("/compare", utils.ErrorHandler(handlers.Compare))
	// ... other routes
}

func (s *Server) Run(port string) error {
	return http.ListenAndServe(port, s.mux)
}
