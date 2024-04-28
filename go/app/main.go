package main

import (
	"mercari-build-training/app/constant"
	"mercari-build-training/app/model"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func root(c echo.Context) error {
	return echo.NewHTTPError(http.StatusOK, "Hello, world!")
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.DEBUG)

	frontURL := os.Getenv("FRONT_URL")
	if frontURL == "" {
		frontURL = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontURL},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// DBとの接続
	// dbにはポインタが返される
	db, err := sql.Open("sqlite3", constant.DbFilePath)
	if err != nil {
		e.Logger.Errorf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	serverImpl := model.ServerImpl{DB: db}

	// テーブルの作成
	if err := serverImpl.CreateTables(); err != nil {
		e.Logger.Errorf("Failed to create tables: %v", err)
	}

	// Routes
	e.GET("/", root)
	e.POST("/items", serverImpl.AddItem)
	e.GET("/items", serverImpl.GetItems)
	e.GET("/search", serverImpl.SearchItems)
	e.GET("/items/:id", serverImpl.GetItemById)
	e.GET("/image/:imageFilename", serverImpl.GetImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
