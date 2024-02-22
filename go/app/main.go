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

	"io"
	"crypto/sha256"
	"mime/multipart"
	// "strconv"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir = "images"
	ItemsFilePath = "./items.json"
	DbFilePath = "../db/mercari.sqlite3"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name string `json:"name"`
	Category string `json:"category"`
	ImageName string `json:"image_name"`
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

	// DBとの接続
	db, err := sql.Open("sqlite3", DbFilePath)
	if err != nil {
		res := Response{Message: "Error connecting to the database"}
    return c.JSON(http.StatusInternalServerError, res)
	}
	defer db.Close()

	// カテゴリが存在するか調べる
	var categoryID int64
	err = db.QueryRow("SELECT id FROM categories WHERE name = ?", category).Scan(&categoryID)
	// カテゴリが存在しない場合、新しいカテゴリを追加
	if err == sql.ErrNoRows {
			result, err := db.Exec("INSERT INTO categories (name) VALUES (?)", category)
			if err != nil {
				res := Response{Message: "Error adding new category to the database"}
				return c.JSON(http.StatusInternalServerError, res)
			}
			categoryID, _ = result.LastInsertId()
	} else if err != nil {
		res := Response{Message: "Error querying categories from the database"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	
	// dbに保存
	stmt, err := db.Prepare("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)")
	if err != nil {
			res := Response{Message: "Error preparing statement for database insertion"}
			return c.JSON(http.StatusInternalServerError, res)
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, categoryID, imageName)
	if err != nil {
			res := Response{Message: "Error saving item to database"}
			return c.JSON(http.StatusInternalServerError, res)
	}
	
	// ログとJSONレスポンスの作成
	c.Logger().Infof("Received item: %s, Category: %s", name, category)
	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}


func getItems(c echo.Context) error {
	// DBとの接続
	db, err := sql.Open("sqlite3", DbFilePath)
	if err != nil {
		res := Response{Message: "Error connecting to the database"}
    return c.JSON(http.StatusInternalServerError, res)
	}
	defer db.Close()

	// データの読み込み
	rows, err := db.Query("SELECT items.name, categories.name as category, items.image_name FROM items join categories on items.category_id = categories.id;")
	if err != nil {
		res := Response{Message: "Error querying items from the database"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	defer rows.Close()

	var items Items

	for rows.Next() {
		var item Item
		err := rows.Scan(&item.Name, &item.Category, &item.ImageName)
		if err != nil {
			res := Response{Message: "Error scanning rows"}
			return c.JSON(http.StatusInternalServerError, res)
		}
		items.Items = append(items.Items, item)
	}

	// ログとJSONレスポンスの作成
	c.Logger().Info("Retrieved items")
	return c.JSON(http.StatusOK, items)
}

func searchItems(c echo.Context) error {
	// DBとの接続
	db, err := sql.Open("sqlite3", DbFilePath)
	if err != nil {
		res := Response{Message: "Error connecting to the database"}
    return c.JSON(http.StatusInternalServerError, res)
	}
	defer db.Close()

	// クエリパラメータを受け取る
	keyword := c.QueryParam("keyword")

	// データの読み込み
	rows, err := db.Query("SELECT items.name, categories.name as category, items.image_name FROM items join categories on items.category_id = categories.id WHERE items.name LIKE '%' || ? || '%'", keyword)
	if err != nil {
		res := Response{Message: "Error querying items from the database"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	defer rows.Close()

	var items Items

	for rows.Next() {
		var item Item
		err := rows.Scan(&item.Name, &item.Category, &item.ImageName)
		if err != nil {
			res := Response{Message: "Error scanning rows"}
			return c.JSON(http.StatusInternalServerError, res)
		}
		items.Items = append(items.Items, item)
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
	hashedImageName := fmt.Sprintf("%x.jpg", hashString)
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


// func getItemById(c echo.Context) error {
// 	id := c.Param("id")
// 	itemID, err := strconv.Atoi(id)
// 	if err != nil {
// 		res := Response{Message: "Invalid item ID"}
// 		return c.JSON(http.StatusBadRequest, res)
// 	}

// 	items, err := decodeItems()
// 	if err != nil {
// 		res := Response{Message: "Error decoding JSON"}
// 		return c.JSON(http.StatusInternalServerError, res)
// 	}
	
// 	if itemID < 1 || itemID > len(items.Items) {
// 		res := Response{Message: "Item not found"}
// 		return c.JSON(http.StatusNotFound, res)
// 	}
// 	item := items.Items[itemID-1]
// 	return c.JSON(http.StatusOK, item)
// }


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
	e.GET("/search", searchItems)
	// e.GET("/items/:id", getItemById)
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
