package utils

import (
	"bytes"
	"io"
	"io/fs"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func UploadFile(fileHeader *multipart.FileHeader) (string, error) {
	uniqueID, err := uuid.NewUUID()
	if err != nil {
		return "", nil
	}
	ext := strings.Split(fileHeader.Filename, ".")
	nameFile := uniqueID.String() + "." + ext[len(ext)-1]
	// To buffer
	openFile, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, openFile); err != nil {
		return "", err
	}

	path := filepath.Join(settingsData.MEDIA_FOLDER, nameFile)
	err = os.WriteFile(path, buf.Bytes(), fs.FileMode(os.O_CREATE))
	if err != nil {
		return "", err
	}

	return nameFile, nil
}

func GetFile(nameFile string) ([]byte, error) {
	return os.ReadFile(filepath.Join(settingsData.MEDIA_FOLDER, nameFile))
}

func DeleteFile(nameFile string) error {
	return os.Remove(filepath.Join(settingsData.MEDIA_FOLDER, nameFile))
}
