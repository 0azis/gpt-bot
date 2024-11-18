package utils

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path"

	"github.com/google/uuid"
)

func SaveImage(file *multipart.FileHeader, savePath string) (string, error) {
	if file == nil {
		return "", errors.New("bad image")
	}

	filename := generateUUID() + path.Ext(file.Filename)
	path := savePath + "/" + filename

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	return filename, nil
}

func generateUUID() string {
	uuidCode := uuid.New()
	return uuidCode.String()
}
