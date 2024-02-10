package router

import (
	"context"
	"database/sql"
	"github.com/Seymour-creates/budget-server/internal/db"
	"github.com/Seymour-creates/budget-server/internal/plaidCtl"
	"github.com/plaid/plaid-go/plaid"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
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

	s.mux.HandleFunc("/get_expenses", utils.ErrorHandler(s.handler.GetExpensesSummary))
	s.mux.HandleFunc("/get_compare", utils.ErrorHandler(s.handler.GetForecastAndExpenses))
	s.mux.HandleFunc("/post_expense", utils.ErrorHandler(s.handler.PostExpense))
	s.mux.HandleFunc("/post_forecast", utils.ErrorHandler(s.handler.PostForecast))
	s.mux.HandleFunc("/link_user_account", utils.ErrorHandler(s.handler.LinkBank))
	s.mux.HandleFunc("/main", utils.ErrorHandler(s.handler.GetRight))
	s.mux.HandleFunc("/create_plaid_item", utils.ErrorHandler(s.handler.CreatePlaidBankItem))
	s.mux.HandleFunc("/oauth_after", utils.ErrorHandler(s.handler.OauthRedirect))
	s.mux.HandleFunc("/refresh_expenses_via_plaid", utils.ErrorHandler(s.handler.UpdateExpenseData))
	// ... other routes
}

func (s *Server) Run(port string) error {
	ctx := context.Background()

	// Set up ngrok configuration and start a tunnel
	ngrokConfig := config.HTTPEndpoint(
		config.WithDomain(os.Getenv("DOMAIN")),
	)

	// Authenticate with ngrok using your auth token stored in an environment variable
	listener, err := ngrok.Listen(ctx, ngrokConfig, ngrok.WithAuthtoken(os.Getenv("NGROK_AUTH_TOKEN")))
	if err != nil {
		log.Fatalf("Failed to start ngrok tunnel: %v", err)
		return err
	}

	// Log the public ngrok URL
	log.Printf("ngrok tunnel started: %s", listener.URL())
	_ = os.Setenv("LOCAL_URL", listener.URL())

	return http.Serve(listener, s.mux)
	//return http.ListenAndServe(port, s.mux)
}
