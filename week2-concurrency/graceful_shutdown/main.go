package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func server(ctx context.Context) error {
	srv := &http.Server{Addr: ":8000"}

	// handle shutdown signal
	go func() {
		<-ctx.Done()
		fmt.Println("Shutting Down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("Shutdown error: %v\n", err)
		}
	}()

	fmt.Println("Server starting...")
	return srv.ListenAndServe()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go server(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("Interrupt recieved, shutting down...")
	cancel()

	time.Sleep(6 * time.Second)
}
