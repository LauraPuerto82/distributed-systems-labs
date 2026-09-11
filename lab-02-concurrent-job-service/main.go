package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	server := NewServer(100)

	server.startWorkers(5)

	http.HandleFunc("/jobs", server.handleJobs)

	httpServer := &http.Server{
		Addr: ":8080",
	}

	go func() {
		log.Println("server listening on :8080")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	close(server.jobs)

	server.workers.Wait()

	log.Println("server stopped")
}
