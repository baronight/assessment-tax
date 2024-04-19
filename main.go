package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"

	_ "github.com/baronight/assessment-tax/docs"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title			K-Tax API
// @version		1.0
// @description	K-Tax Calculate API
// @host			localhost:8080
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "1323"
	}
	e := echo.New()
	// setup swagger document
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})

	// make graceful shutdown
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	e.POST("/quit", func(c echo.Context) error {
		stop()
		return c.String(http.StatusOK, "OK")
	})

	// Start server
	go func() {
		if err := e.Start(fmt.Sprintf(":%s", port)); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-shutdownCtx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	} else {
		fmt.Println("shutting down the server")
	}
}
