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
	"strconv"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"errors"
)

const (
	ImgDir = "images"
	DbFilePath = "./db/mercari.sqlite3"
	ItemsSchemaPath = "./db/items.db"
	CategoriesSchemaPath = "./db/categories.db"
)

type Response struct {
	Message string `json:"message"`
}

type ServerImpl struct {
	db *sql.DB
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
	return echo.NewHTTPError(http.StatusOK, "Hello, world!")
}

func (s ServerImpl) addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")


	// 画像ファイルの保存
	imageFile, err := c.FormFile("image")
	if err != nil {
			c.Logger().Errorf("Error uploading image: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error uploading image")
	}
	imageName, err := saveImage(imageFile)
	if err != nil {
			c.Logger().Errorf("Error saving image: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error saving image")
	}


	// トランザクション開始
	tx, err := s.db.Begin()
	if err != nil {
		c.Logger().Errorf("Error starting database transactione: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error starting database transaction")
	}
	defer tx.Rollback()
	

	// カテゴリが存在するか調べる
	var categoryID int64
	err = tx.QueryRow("SELECT id FROM categories WHERE name = ?", category).Scan(&categoryID)
	// カテゴリが存在しない場合、新しいカテゴリを追加
	if errors.Is(err, sql.ErrNoRows) {
			result, err := tx.Exec("INSERT INTO categories (name) VALUES (?)", category)
			if err != nil {
				c.Logger().Errorf("Error adding new category to the database: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Error adding new category to the database")
			}
			categoryID, _ = result.LastInsertId()
	} else if err != nil {
		c.Logger().Errorf("Error querying categories from the database: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error querying categories from the database")
	}
	

	// dbに保存
	stmt, err := tx.Prepare("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)")
	if err != nil {
		c.Logger().Errorf("Error preparing statement for database insertion: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error preparing statement for database insertion")
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, categoryID, imageName)
	if err != nil {
		c.Logger().Errorf("Error saving item to database: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error saving item to database")
	}


	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		c.Logger().Errorf("Error committing database transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error committing database transaction")
	}
	

	// ログとJSONレスポンスの作成
	c.Logger().Infof("Received item: %s, Category: %s", name, category)
	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func  (s ServerImpl) getItems(c echo.Context) error {
	// データの読み込み
	rows, err := s.db.Query("SELECT items.name, categories.name as category, items.image_name FROM items join categories on items.category_id = categories.id;")
	if err != nil {
		c.Logger().Errorf("Error querying items from the database: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error querying items from the database")
	}
	defer rows.Close()

	var items Items

	for rows.Next() {
		var item Item
		err := rows.Scan(&item.Name, &item.Category, &item.ImageName)
		if err != nil {
			c.Logger().Errorf("Error scanning rows: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error scanning rows")
		}
		items.Items = append(items.Items, item)
	}

	// ログとJSONレスポンスの作成
	c.Logger().Info("Retrieved items")
	return c.JSON(http.StatusOK, items)
}

func  (s ServerImpl) searchItems(c echo.Context) error {
	// クエリパラメータを受け取る
	keyword := c.QueryParam("keyword")

	// データの読み込み
	rows, err := s.db.Query("SELECT items.name, categories.name as category, items.image_name FROM items join categories on items.category_id = categories.id WHERE items.name LIKE '%' || ? || '%'", keyword)
	if err != nil {
		c.Logger().Errorf("Error querying items from the database: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error querying items from the database")
	}
	defer rows.Close()

	var items Items

	for rows.Next() {
		var item Item
		err := rows.Scan(&item.Name, &item.Category, &item.ImageName)
		if err != nil {
			c.Logger().Errorf("Error scanning rows: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error scanning rows")
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
		c.Logger().Error("Image path does not end with .jpg")
		return echo.NewHTTPError(http.StatusInternalServerError, "Image path does not end with .jpg")
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

func  (s ServerImpl) getItemById(c echo.Context) error {
  id := c.Param("id")
	itemID, err := strconv.Atoi(id)
	if err != nil {
		c.Logger().Errorf("Invalid item ID: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid item ID")
	}
	var item Item

	// データの読み込み
	row := s.db.QueryRow("SELECT items.name, categories.name as category, items.image_name FROM items join categories on items.category_id = categories.id WHERE items.id = ?", itemID)
	err = row.Scan(&item.Name, &item.Category, &item.ImageName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.Logger().Errorf("Item not found: %v", err)
			return echo.NewHTTPError(http.StatusNotFound, "Item not found")
		}
		c.Logger().Errorf("Error querying items from the database: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error querying items from the database")
	}

	// ログとJSONレスポンスの作成
	c.Logger().Info("Retrieved items")
	return c.JSON(http.StatusOK, item)
}

func  (s ServerImpl) createTables() error {
	// ItemsSchema読み込み
	itemsSchema, err := os.ReadFile(ItemsSchemaPath)
	if err != nil {
		return fmt.Errorf("Error reading Items schema file: %v", err)
	}
	// CategoriesSchema読み込み
	categoriesSchema, err := os.ReadFile(CategoriesSchemaPath)
	if err != nil {
		return fmt.Errorf("Error reading Categories schema file: %v", err)
	}
	// Itemsテーブル作成
	if _, err := s.db.Exec(string(itemsSchema)); err != nil {
		return fmt.Errorf("Error creating Items table: %v", err)
	}
	// Categoriesテーブル作成
	if _, err := s.db.Exec(string(categoriesSchema)); err != nil {
		return fmt.Errorf("Error creating Categories table: %v", err)
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

	// DBとの接続
	// dbにはポインタが返される
	db, err := sql.Open("sqlite3", DbFilePath)
	if err != nil {
		e.Logger.Errorf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	serverImpl := ServerImpl{db: db}

	// テーブルの作成
	if err := serverImpl.createTables(); err != nil {
		e.Logger.Errorf("Failed to create tables: %v", err)
	}

	// Routes
	e.GET("/", root)
	e.POST("/items", serverImpl.addItem)
	e.GET("/items", serverImpl.getItems)
	e.GET("/search", serverImpl.searchItems)
	e.GET("/items/:id", serverImpl.getItemById)
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
