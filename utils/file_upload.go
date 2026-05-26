// utils/file_upload.go
package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	storage_go "github.com/supabase-community/storage-go"
)

func SaveUploadedFile(ctx *gin.Context, formField string, destDir string) (*string, error) {
	fileHeader, err := ctx.FormFile(formField)
	if err != nil {
		return nil, nil
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	supabaseBucket := os.Getenv("SUPABASE_BUCKET")

	if supabaseBucket == "" {
		supabaseBucket = "noir-assets"
	}

	storageClient := storage_go.NewClient(supabaseURL, supabaseKey, nil)

	// Clean destDir to avoid leading slash issues
	cleanDestDir := strings.TrimPrefix(destDir, "uploads/")
	cleanDestDir = strings.TrimPrefix(cleanDestDir, "uploads\\")

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(fileHeader.Filename))
	path := filepath.ToSlash(filepath.Join(cleanDestDir, filename))

	_, err = storageClient.UploadFile(supabaseBucket, path, file)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to supabase: %w", err)
	}

	// Construct public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, supabaseBucket, path)

	return &publicURL, nil
}
