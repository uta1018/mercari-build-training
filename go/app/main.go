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
	"crypto/sha256"
	"mime/multipart"
	"strconv"
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
	ImageName string `json:"image-name"`
}

type Items struct {
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

	// 画像ファイルの保存
	imageFile, err := c.FormFile("image")
	if err != nil {
			res := Response{Message: "Error uploading image"}
			return c.JSON(http.StatusInternalServerError, res)
	}
	imageName, err := saveImage(imageFile)
	if err != nil {
			res := Response{Message: "Error saving image"}
			return c.JSON(http.StatusInternalServerError, res)
	}

	// items.jsonのデータを構造体にデコード
	items, err := decodeItems()
	if err != nil {
		res := Response{Message: "Error decoding JSON"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	// 新しい商品情報を追加
	newItem := Item{Name: name, Category: category, ImageName: imageName}
	items.Items = append(items.Items, newItem)

	// 更新した内容をエンコードしてファイルに書き込む
	if err := encodeItems(items); err != nil {
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
	items, err := decodeItems()
	if err != nil {
		res := Response{Message: "Error decoding JSON"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	
	// ログとJSONレスポンスの作成
	c.Logger().Info("Retrieved items")
	return c.JSON(http.StatusOK, items)
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

// saveImage は画像を保存し、ハッシュ化した名前を返す関数
func saveImage(file *multipart.FileHeader) (string, error) {
	// 元ファイルを開く
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// 元ファイルのハッシュを計算
	hash := sha256.New()
	if _, err := io.Copy(hash, src); err != nil {
		return "", err
	}

	// ハッシュを16進数文字列に変換
	hashString := hash.Sum(nil)

	// 行先ファイルを作成
	hashedImageName := fmt.Sprintf("%x.jpeg", hashString)
	dst, err := os.Create(path.Join(ImgDir, hashedImageName))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// 元ファイルを読み込み、行先ファイルに保存
	src.Seek(0, 0) 
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return hashedImageName, nil
}

func getItemById(c echo.Context) error {
	id := c.Param("id")
	itemID, err := strconv.Atoi(id)
	if err != nil {
		res := Response{Message: "Invalid item ID"}
		return c.JSON(http.StatusBadRequest, res)
	}

	items, err := decodeItems()
	if err != nil {
		res := Response{Message: "Error decoding JSON"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	
	if itemID < 1 || itemID > len(items.Items) {
		res := Response{Message: "Item not found"}
		return c.JSON(http.StatusNotFound, res)
	}
	item := items.Items[itemID-1]
	return c.JSON(http.StatusOK, item)
}

// decodeItems は items.json ファイルをデコードする関数
func decodeItems() (*Items, error) {
	file, err := os.OpenFile(ItemsFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var items Items
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&items); err != nil && err != io.EOF {
		return nil, err
	}

	return &items, nil
}

// encodeItems は items.json ファイルにエンコードする関数
func encodeItems(items *Items) error {
	file, err := os.OpenFile(ItemsFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// ポインタを先頭に戻す
	file.Seek(0, 0)
	file.Truncate(0)

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(items); err != nil {
		return err
	}

	return nil
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

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/items", getItems)
	e.GET("/items/:id", getItemById)
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
