package rosetta

import (
	"errors"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/pg"
	"github.com/p2p-org/mbelt-filecoin-streamer/rosetta/services"
	streamerServices "github.com/p2p-org/mbelt-filecoin-streamer/services"
	"log"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

const (
	ReadTimeout  time.Duration = 30 * time.Second
	WriteTimeout time.Duration = 30 * time.Second
	IdleTimeout  time.Duration = 60 * time.Second
)

var listener net.Listener

// StartServers starts the rosetta http server
// TODO (dm): optimize rosetta to use single flight & use extra caching type DB to avoid re-processing data
func StartServers(config *config.Config, endpoint string) error {
	err := streamerServices.InitServices(config)
	if err != nil {
		log.Println("[App][Debug]", "Cannot init services:", err)
		return err
	}

	serverAsserter, err := asserter.NewServer(
		streamerServices.MethodsList(),
		true,
		[]*types.NetworkIdentifier{{Blockchain: "Filecoin", Network: streamerServices.App().StateService().NetworkName()}},
		nil,
		false,
	)
	if err != nil {
		return err
	}

	router := recoverMiddleware(server.CorsMiddleware(loggerMiddleware(getRouter(serverAsserter, streamerServices.App().PgDatastore()))))
	log.Print("Starting rosetta server...")
	if listener, err = net.Listen("tcp", endpoint); err != nil {
		return err
	}
	go newHTTPServer(router).Serve(listener)
	fmt.Printf("Started Rosetta server at: %v\n", endpoint)
	return nil
}

// StopServers stops the rosetta http server
func StopServers() error {
	if listener == nil {
		return nil
	}
	if err := listener.Close(); err != nil {
		return err
	}
	return nil
}

func newHTTPServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:      handler,
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
		IdleTimeout:  IdleTimeout,
	}
}

func getRouter(asserter *asserter.Asserter, pg *pg.PgDatastore) http.Handler {
	return server.NewRouter(
		server.NewAccountAPIController(services.NewAccountAPI(pg), asserter),
		server.NewBlockAPIController(services.NewBlockAPI(pg), asserter),
		server.NewMempoolAPIController(services.NewMempoolAPI(), asserter),
		server.NewNetworkAPIController(services.NewNetworkAPI(), asserter),
		server.NewConstructionAPIController(services.NewConstructionAPI(), asserter),
		server.NewEventsAPIController(services.NewEventsAPI(pg), asserter),
		server.NewSearchAPIController(services.NewSearchAPI(pg), asserter),
	)
}

func recoverMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("unknown error")
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println("Rosetta Error")
				// Print to stderr for quick check of rosetta activity
				debug.PrintStack()
				_, _ = fmt.Fprintf(
					os.Stderr, "%s PANIC: %s\n", time.Now().Format("2006-01-02 15:04:05"), err.Error(),
				)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func loggerMiddleware(router http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		router.ServeHTTP(w, r)
		msg := fmt.Sprintf(
			"Rosetta: %s %s %s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
		log.Println(msg)
		// Print to stdout for quick check of rosetta activity
		log.Printf("%s %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
	})
}