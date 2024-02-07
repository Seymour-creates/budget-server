package router

import (
	"database/sql"
	"github.com/Seymour-creates/budget-server/internal/db"
	"log"
	"net/http"

	"github.com/Seymour-creates/budget-server/internal/handlers"
	"github.com/Seymour-creates/budget-server/internal/utils"
)

type Server struct {
	mux   *http.ServeMux
	mysql *db.Manager
}

func NewServer() *Server {
	mysqlConn, err := sql.Open("mysql", "yourdatasource")
	if err != nil {
		log.Printf("error connecting to db: %v", err)
	}
	DBManager := db.NewDBManager(mysqlConn)
	server := &Server{
		mux:   http.NewServeMux(),
		mysql: DBManager,
	}
	server.registerRoutes()
	return server
}

func (s *Server) registerRoutes() {
	// Create file server and serve static files
	fs := http.FileServer(http.Dir("./internal/assets"))
	s.mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// s.mux.HandleFunc("/get_summary", utils.ErrorHandler(handlers.GetExpensesSummary))
	// s.mux.HandleFunc("/get_compare", utils.ErrorHandler(handlers.GetCompare))
	// s.mux.HandleFunc("/post_expense", utils.ErrorHandler(handlers.PostExpense))
	// s.mux.HandleFunc("/post_forecast", utils.ErrorHandler(handlers.PostForecast))
	// s.mux.HandleFunc("/link_cheddar", utils.ErrorHandler(handlers.LinkBank))
	s.mux.HandleFunc("/main", utils.ErrorHandler(handlers.GetRight))
	// s.mux.HandleFunc("/create_item", utils.ErrorHandler(handlers.CreateItem))
	// s.mux.HandleFunc("/refresh_expense_data", utils.ErrorHandler(handlers.UpdateExpenseData))
	// ... other routes
}

func (s *Server) Run(port string) error {
	log.Printf("trying to start svr in router.Run()")
	return http.ListenAndServe(port, s.mux)
}
