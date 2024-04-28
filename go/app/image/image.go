package image

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
)

const (
	ImgDir = "../images"
)

// saveImage は画像を保存し、ハッシュ化した名前を返す関数
func SaveImage(file *multipart.FileHeader) (string, error) {
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
