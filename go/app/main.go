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

var Items struct {
	Items []Item `json:"items"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")

	// items.jsonのデータを構造体にデコード
	if err := decodeItems(); err != nil {
		res := Response{Message: "Error decoding JSON"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	// 新しい商品情報を追加
	newItem := Item{Name: name, Category: category}
	Items.Items = append(Items.Items, newItem)

	// 更新した内容をエンコードしてファイルに書き込む
	if err := encodeItems(); err != nil {
		res := Response{Message: "Error encoding JSON"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	
	// ログとJSONレスポンスの作成
	c.Logger().Infof("Received item: %s, Category: %s", name, category)
	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)

}

func getItems(c echo.Context) error {
	// items.jsonのデータを構造体にデコード
	if err := decodeItems(); err != nil {
		res := Response{Message: "Error decoding JSON"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	
	// ログとJSONレスポンスの作成
	c.Logger().Info("Retrieved items")
	return c.JSON(http.StatusOK, Items)
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

// decodeItems は items.json ファイルをデコードする関数
func decodeItems() error {
	file, err := os.OpenFile(ItemsFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&Items); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// encodeItems は items.json ファイルにエンコードする関数
func encodeItems() error {
	file, err := os.OpenFile(ItemsFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	file.Seek(0, 0)
	file.Truncate(0)

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(Items); err != nil {
		return err
	}

	return nil
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
	e.GET("/items", getItems)
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
