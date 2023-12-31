package router

import (
	"github.com/Seymour-creates/budget-server/internal/handlers"
	"github.com/Seymour-creates/budget-server/internal/utils"
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
	// Create file server and serve static files
	fs := http.FileServer(http.Dir("./internal/assets"))
	s.mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	s.mux.HandleFunc("/get_summary", utils.ErrorHandler(handlers.GetExpensesSummary))
	s.mux.HandleFunc("/get_compare", utils.ErrorHandler(handlers.GetCompare))
	s.mux.HandleFunc("/post_expense", utils.ErrorHandler(handlers.PostExpense))
	s.mux.HandleFunc("/post_forecast", utils.ErrorHandler(handlers.PostForecast))
	s.mux.HandleFunc("/link_cheddar", utils.ErrorHandler(handlers.LinkBank))
	s.mux.HandleFunc("/create_item", utils.ErrorHandler(handlers.CreateItem))
	s.mux.HandleFunc("/refresh_expense_data", utils.ErrorHandler(handlers.UpdateExpenseData))
	// ... other routes
}

func (s *Server) Run(port string) error {
	return http.ListenAndServe(port, s.mux)
}
