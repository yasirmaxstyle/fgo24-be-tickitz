// utils/file_upload.go
package utils

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func SaveUploadedFile(ctx *gin.Context, formField string, destDir string) (*string, error) {
	file, err := ctx.FormFile(formField)
	if err != nil {
		return nil, nil
	}

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
	path := filepath.Join(destDir, filename)

	if err := ctx.SaveUploadedFile(file, path); err != nil {
		return nil, fmt.Errorf("failed to save uploaded file: %w", err)
	}

	return &path, nil
}
