package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Thivyesh/accident-prediction-api/go-services/camera-service/api"
	"github.com/Thivyesh/accident-prediction-api/go-services/camera-service/service"
	"github.com/gorilla/mux"
)

func main() {
	// Create a new camera service
	cameraService := service.NewCameraService()

	// Create a http handler
	handler := api.NewHandler()

	// Create router and register routes
	router := mux.NewRouter()

	// API routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/sources", handler.HandleAddSource).Methods("POST")
	apiRouter.HandleFunc("/sources", handler.HandleListSources).Methods("GET")
	apiRouter.HandleFunc("/sources/{id}", handler.HandleRemoveSource).Methods("DELETE")
	apiRouter.HandleFunc("sources/{id}/stream", handler.HandleStreamFrames)

	// Create a HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
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
