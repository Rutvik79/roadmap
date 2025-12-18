package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"week3_services_testing/api/internal/s3helper"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileHandler struct {
	s3Client *s3helper.S3Client
}

func NewFileHandler(s3Client *s3helper.S3Client) *FileHandler {
	return &FileHandler{
		s3Client: s3Client,
	}
}

type UploadResponse struct {
	Message  string `json:"message"`
	FileKey  string `json:"file_key"`
	FileName string `json:"file_name"`
	Size     int64  `json:"size"`
}

// UploadFile handles file uploads to S3
func (h *FileHandler) UploadFile(c *gin.Context) {
	// Get file from request
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file provided",
		})
		return
	}
	defer file.Close()

	// Validate file size (10MB max)
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if header.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File too large. Maximum size is 10MB",
		})
		return
	}

	// Validate file type (optional)
	contentType := header.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"application/pdf": true,
		"text/plain":      true,
	}

	if !allowedTypes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("File type not allowed: %s", contentType),
		})
		return
	}

	// Generate unique file key
	ext := filepath.Ext(header.Filename)
	fileKey := fmt.Sprintf("uploads/%s/%s%s",
		time.Now().Format("2006/01/02"),
		uuid.New().String(),
		ext,
	)

	// Upload to S3
	fmt.Println("File key", fileKey, "file, contentType", contentType)
	err = h.s3Client.UploadFile(c.Request.Context(), fileKey, file, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload file",
		})
		return
	}

	c.JSON(http.StatusOK, UploadResponse{
		Message:  "File uploaded successfully",
		FileKey:  fileKey,
		FileName: header.Filename,
		Size:     header.Size,
	})
}

type FileInfo struct {
	Key          string    `json:"key"`
	FileName     string    `json:"file_name"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"content_type"`
	LastModified time.Time `json:"last_modified"`
	URL          string    `json:"url"`
}

// DownloadFile generates a pre-signed URL for downloading
func (h *FileHandler) DownloadFile(c *gin.Context) {
	fileKey := c.Param("key")
	if fileKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File key is required",
		})
		return
	}

	// Remove leading slash if present
	fileKey = strings.TrimPrefix(fileKey, "/")

	// Get file metadata
	metadata, err := h.s3Client.GetFileMetadata(c.Request.Context(), fileKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}

	// Generate pre-signed URL (valid for 1 hour)
	url, err := h.s3Client.GeneratePresignedURL(
		c.Request.Context(),
		fileKey,
		1*time.Hour,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate download URL",
		})
		return
	}

	c.JSON(http.StatusOK, FileInfo{
		Key:          fileKey,
		FileName:     filepath.Base(fileKey),
		Size:         *metadata.ContentLength,
		ContentType:  *metadata.ContentType,
		LastModified: *metadata.LastModified,
		URL:          url,
	})
}

// ListFiles lists all uploaded files
func (h *FileHandler) ListFiles(c *gin.Context) {
	files, err := h.s3Client.ListFiles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list files",
		})
		return
	}

	// Get metadata for each file
	var fileInfos []FileInfo
	for _, fileKey := range files {
		metadata, err := h.s3Client.GetFileMetadata(c.Request.Context(), fileKey)
		if err != nil {
			continue // Skip files we can't access
		}

		// Generate pre-signed URL
		url, err := h.s3Client.GeneratePresignedURL(
			c.Request.Context(),
			fileKey,
			1*time.Hour,
		)
		if err != nil {
			url = "" // URL generation failed, continue without it
		}

		fileInfos = append(fileInfos, FileInfo{
			Key:          fileKey,
			FileName:     filepath.Base(fileKey),
			Size:         *metadata.ContentLength,
			ContentType:  *metadata.ContentType,
			LastModified: *metadata.LastModified,
			URL:          url,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(fileInfos),
		"files": fileInfos,
	})
}

// DeleteFile deletes a file from S3
func (h *FileHandler) DeleteFile(c *gin.Context) {
	fileKey := c.Param("key")
	if fileKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File key is required",
		})
		return
	}

	// Remove leading slash
	fileKey = strings.TrimPrefix(fileKey, "/")

	// Delete from S3
	err := h.s3Client.DeleteFile(c.Request.Context(), fileKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File deleted successfully",
		"file_key": fileKey,
	})
}
