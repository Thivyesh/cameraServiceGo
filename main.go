// @title Camera Service API
// @version 1.0
// @description A real-time video streaming service with multiple camera sources
// @host localhost:8080
// @BasePath /api
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Thivyesh/cameraServiceGo/api"
	_ "github.com/Thivyesh/cameraServiceGo/docs"
	"github.com/Thivyesh/cameraServiceGo/service"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func main() {
	// Create a new camera service
	cameraService := service.NewCameraService()

	// Create a http handler
	handler := api.NewHandler(cameraService)

	// Create router and register routes
	router := mux.NewRouter()

	// Add Swagger documentation route
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"), // The URL pointing to API definition
	))

	// API routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/sources", handler.HandleListSources).Methods("GET")
	apiRouter.HandleFunc("/sources", handler.HandleAddSource).Methods("POST")
	apiRouter.HandleFunc("/sources/{id}", handler.HandleRemoveSource).Methods("DELETE")
	apiRouter.HandleFunc("/sources/{id}/stream", handler.HandleStreamFrames)

	// Create CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	// Create a HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: c.Handler(router),
	}

	// Channel to listen for interupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("Server stopped")
}
