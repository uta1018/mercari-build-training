package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"encoding/json"
	"io"
)

const (
	ImgDir = "images"
	ItemsFilePath = "./items.json"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name string `json:"name"`
	Category string `json:"category"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	
	// items.jsonファイルを開くもしくは新規作成
	file, err := os.OpenFile(ItemsFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		res := Response{Message: "Error opening file"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	defer file.Close()

	// items.jsonのデータを構造体にデコード
	items := struct {
		Items []Item `json:"items"`
	}{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&items); err != nil && err != io.EOF {
		res := Response{Message: "Error decoding JSON"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	// 新しい商品情報を追加
	newItem := Item{Name: name, Category: category}
	items.Items = append(items.Items, newItem)

	// ファイルを先頭に戻し、内容をクリア
	file.Seek(0, 0)
	file.Truncate(0)

	// 更新した内容をエンコードしてファイルに書き込む
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(items); err != nil {
		res := Response{Message: "Error encoding JSON"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	
	// ログとJSONレスポンスの作成
	c.Logger().Infof("Received item: %s, Category: %s", name, category)
	message := fmt.Sprintf("Item received: %s, Category: %s", name, category)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)

	// message := fmt.Sprintf("item received: %s", name)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	frontURL := os.Getenv("FRONT_URL")
	if frontURL == "" {
		frontURL = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontURL},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
