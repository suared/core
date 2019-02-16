package api

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/suared/core/api/middleware"

	//Core Config setup
	_ "github.com/suared/core/infra"

	"github.com/gorilla/mux"
)

//Config - Interface to setup API specific routes
type Config interface {
	SetupRoutes(router *mux.Router)
	StartServer() bool
}

//StartHTTPListener - Starts the Http server.  Uses the mux provided example starter and adds common middleware + router setup interface
//TODO: Change additional hard codes to  env variables
func StartHTTPListener(config Config) {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	r := mux.NewRouter()

	//Add any middleware to be used
	middleware.SetUpMiddleware(r)

	// Setup core architecture handlers and routes
	r.HandleFunc("/health", HealthCheckHandler)

	// Create custom routes last - enables greatest customizaiton flex as well as enables non-standard listener
	config.SetupRoutes(r)

	// Only continue with web api standard setup if not being overridden by Lambda or other local handler method
	if config.StartServer() {
		srv := &http.Server{
			Addr: os.Getenv("PROCESS_LISTEN_ADDR"),
			// Good practice to set timeouts to avoid Slowloris attacks.
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      r, // Pass our instance of gorilla/mux in.
		}

		log.Printf("API Listener started at: %v", os.Getenv("PROCESS_LISTEN_ADDR"))
		// Run our server in a goroutine so that it doesn't block.
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				log.Println(err)
			}
		}()

		c := make(chan os.Signal, 1)
		// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
		// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
		signal.Notify(c, os.Interrupt)

		// Block until we receive our signal.
		<-c

		// Create a deadline to wait for.
		ctx, cancel := context.WithTimeout(context.Background(), wait)
		defer cancel()
		// Doesn't block if no connections, but will otherwise wait
		// until the timeout deadline.
		srv.Shutdown(ctx)
		// Optionally, you could run srv.Shutdown in a goroutine and block on
		// <-ctx.Done() if your application should wait for other services
		// to finalize based on context cancellation.
		log.Println("shutting down")
		os.Exit(0)
	}
}
