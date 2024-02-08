package router

import (
	"database/sql"
	"github.com/Seymour-creates/budget-server/internal/db"
	"github.com/Seymour-creates/budget-server/internal/plaidCtl"
	"github.com/plaid/plaid-go/plaid"
	"log"
	"net/http"
	"os"

	"github.com/Seymour-creates/budget-server/internal/handlers"
	"github.com/Seymour-creates/budget-server/internal/utils"
)

type Server struct {
	mux     *http.ServeMux
	handler *handlers.Handler
}

func createNewPlaidClient() *plaid.APIClient {
	clientOptions := plaid.NewConfiguration()
	clientOptions.AddDefaultHeader("PLAID-CLIENT-ID", os.Getenv("PLAID_CLIENT_ID"))
	clientOptions.AddDefaultHeader("PLAID-SECRET", os.Getenv("PLAID_SECRET"))

	// Use plaidCtl.Development or plaidCtl.Production depending on your environment
	clientOptions.UseEnvironment(plaid.Sandbox)
	return plaid.NewAPIClient(clientOptions)
}

func ConfigServer() *Server {
	mysqlConn, err := sql.Open("mysql", "yourdatasource")
	if err != nil {
		log.Printf("error connecting to db: %v", err)
	}
	DBManager := db.NewDBManager(mysqlConn)
	plaidClient := plaidCtl.NewService(createNewPlaidClient())
	handler := handlers.MakeNewHttpHandler(plaidClient, DBManager)
	server := &Server{
		mux:     http.NewServeMux(),
		handler: handler,
	}
	server.registerRoutes()
	return server
}

func (s *Server) registerRoutes() {
	// Create file server and serve static files
	fs := http.FileServer(http.Dir("./internal/assets"))
	s.mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	s.mux.HandleFunc("/get_summary", utils.ErrorHandler(s.handler.GetExpensesSummary))
	s.mux.HandleFunc("/get_compare", utils.ErrorHandler(s.handler.GetCompare))
	s.mux.HandleFunc("/post_expense", utils.ErrorHandler(s.handler.PostExpense))
	s.mux.HandleFunc("/post_forecast", utils.ErrorHandler(s.handler.PostForecast))
	s.mux.HandleFunc("/link_cheddar", utils.ErrorHandler(s.handler.LinkBank))
	s.mux.HandleFunc("/main", utils.ErrorHandler(s.handler.GetRight))
	s.mux.HandleFunc("/create_item", utils.ErrorHandler(s.handler.CreateItem))
	s.mux.HandleFunc("/refresh_expense_data", utils.ErrorHandler(s.handler.UpdateExpenseData))
	// ... other routes
}

func (s *Server) Run(port string) error {
	log.Printf("trying to start svr in router.Run()")
	return http.ListenAndServe(port, s.mux)
}
