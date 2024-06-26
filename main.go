package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/baronight/assessment-tax/db"
	_ "github.com/baronight/assessment-tax/docs"
	"github.com/baronight/assessment-tax/handlers"
	"github.com/baronight/assessment-tax/middlewares"
	"github.com/baronight/assessment-tax/services"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @securityDefinitions.basic BasicAuth

// @title			K-Tax API
// @version		1.0
// @description	K-Tax Calculate API
// @host			localhost:8080
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "1323"
	}
	db, err := db.New()
	if err != nil {
		panic(err)
	}

	e := echo.New()
	// e.Validator = &models.CustomValidator{Validator: validator.New()}

	e.Use(middleware.Logger(), middleware.Recover(), middleware.CORS())
	// setup swagger document
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})

	taxService := services.NewTaxService(db)
	taxHandler := handlers.NewTaxHandlers(taxService)
	groupTax := e.Group("/tax")
	groupTax.POST("/calculations", taxHandler.TaxCalculateHandler)
	groupTax.POST("/calculations/upload-csv", taxHandler.TaxUploadCsvHandler)

	adminService := services.NewAdminService(db)
	adminHandler := handlers.NewAdminHandlers(adminService)
	groupAdmin := e.Group("/admin")
	groupAdmin.Use(middlewares.BasicAuthMiddleware())
	groupAdmin.POST("/deductions/personal", adminHandler.PersonalDeductionConfigHandler)
	groupAdmin.POST("/deductions/k-receipt", adminHandler.KReceiptDeductionConfigHandler)

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
	fmt.Println()
	// close db
	if err := db.Db.Close(); err != nil {
		e.Logger.Fatal(err)
	} else {
		fmt.Println("closing database connection")
	}
	// close server
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	} else {
		fmt.Println("shutting down the server")
	}
}
