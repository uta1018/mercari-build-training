package model

import (
	"database/sql"
	"errors"
	"fmt"
	"mercari-build-training/app/constant"
	"mercari-build-training/app/image"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type ServerImpl struct {
	DB *sql.DB
}

func (s ServerImpl) AddItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")

	// 画像ファイルの保存
	imageFile, err := c.FormFile("image")
	if err != nil {
		c.Logger().Errorf("Error uploading image: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error uploading image")
	}
	imageName, err := image.SaveImage(imageFile)
	if err != nil {
		c.Logger().Errorf("Error saving image: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error saving image")
	}

	// トランザクション開始
	tx, err := s.DB.Begin()
	if err != nil {
		c.Logger().Errorf("Error starting database transactione: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error starting database transaction")
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			c.Logger().Errorf("Database transaction rollback failed: %v", err)
		}
	}()

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

func (s ServerImpl) GetItems(c echo.Context) error {
	// データの読み込み
	rows, err := s.DB.Query("SELECT items.id, items.name, categories.name as category, items.image_name FROM items join categories on items.category_id = categories.id;")
	if err != nil {
		c.Logger().Errorf("Error querying items from the database: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error querying items from the database")
	}
	defer rows.Close()

	var items Items

	for rows.Next() {
		var item Item
		err := rows.Scan(&item.Id, &item.Name, &item.Category, &item.ImageName)
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

func (s ServerImpl) SearchItems(c echo.Context) error {
	// クエリパラメータを受け取る
	keyword := c.QueryParam("keyword")

	// データの読み込み
	rows, err := s.DB.Query("SELECT items.id, items.name, categories.name as category, items.image_name FROM items join categories on items.category_id = categories.id WHERE items.name LIKE '%' || ? || '%'", keyword)
	if err != nil {
		c.Logger().Errorf("Error querying items from the database: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error querying items from the database")
	}
	defer rows.Close()

	var items Items

	for rows.Next() {
		var item Item
		err := rows.Scan(&item.Id, &item.Name, &item.Category, &item.ImageName)
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

func (s ServerImpl) GetImg(c echo.Context) error {
	// id+.jpgが渡ってくる
	imageIDWithExtension := c.Param("imageFilename")
	imgPath := path.Join(image.ImgDir, imageIDWithExtension)

	// 拡張子がjpgがチェック
	if !strings.HasSuffix(imgPath, ".jpg") {
		c.Logger().Errorf("Image path does not end with .jpg, got: %s", imgPath)
		return echo.NewHTTPError(http.StatusInternalServerError, "Image path does not end with .jpg")
	}

	// 拡張子を取り除く
	imageID := strings.TrimSuffix(imageIDWithExtension, ".jpg")

	// imageIDを使ってデータベースに問い合わせて、該当する画像のパスを取得する
	var imgPathById string
	err := s.DB.QueryRow("SELECT image_name FROM items WHERE id = ?", imageID).Scan(&imgPathById)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.Logger().Errorf("Image not found for ID: %s", imageID)
		} else {
			c.Logger().Errorf("Error querying image path from the database: %v", err)
		}
		imgPath = path.Join(image.ImgDir, "default.jpg")
	} else {
		imgPath = path.Join(image.ImgDir, imgPathById)
	}

	// ファイルが存在しないときはデフォルトを表示
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(image.ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func (s ServerImpl) GetItemById(c echo.Context) error {
	id := c.Param("id")
	itemID, err := strconv.Atoi(id)
	if err != nil {
		c.Logger().Errorf("Invalid item ID: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid item ID")
	}
	var item Item

	// データの読み込み
	row := s.DB.QueryRow("SELECT items.id, items.name, categories.name as category, items.image_name FROM items join categories on items.category_id = categories.id WHERE items.id = ?", itemID)
	err = row.Scan(&item.Id, &item.Name, &item.Category, &item.ImageName)
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

func (s ServerImpl) CreateTables() error {
	// ItemsSchema読み込み
	itemsSchema, err := os.ReadFile(constant.ItemsSchemaPath)
	if err != nil {
		return fmt.Errorf("Error reading Items schema file: %v", err)
	}
	// CategoriesSchema読み込み
	categoriesSchema, err := os.ReadFile(constant.CategoriesSchemaPath)
	if err != nil {
		return fmt.Errorf("Error reading Categories schema file: %v", err)
	}
	// Itemsテーブル作成
	if _, err := s.DB.Exec(string(itemsSchema)); err != nil {
		return fmt.Errorf("Error creating Items table: %v", err)
	}
	// Categoriesテーブル作成
	if _, err := s.DB.Exec(string(categoriesSchema)); err != nil {
		return fmt.Errorf("Error creating Categories table: %v", err)
	}

	return nil
}
