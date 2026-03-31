package upload

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type LocalStore struct {
	baseDir string
}

func NewLocalStore(baseDir string) *LocalStore {
	return &LocalStore{baseDir: baseDir}
}

func (s *LocalStore) SaveImage(category string, header *multipart.FileHeader) (*StoredFile, error) {
	return s.save(category, header, true)
}

func (s *LocalStore) SaveFile(category string, header *multipart.FileHeader) (*StoredFile, error) {
	return s.save(category, header, false)
}

func (s *LocalStore) save(category string, header *multipart.FileHeader, imagesOnly bool) (*StoredFile, error) {
	source, err := header.Open()
	if err != nil {
		return nil, err
	}
	defer source.Close()

	buffer := make([]byte, 512)
	readBytes, err := source.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if seeker, ok := source.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
	}

	mimeType := http.DetectContentType(buffer[:readBytes])
	if imagesOnly && !strings.HasPrefix(mimeType, "image/") {
		return nil, fmt.Errorf("unsupported image mime type: %s", mimeType)
	}

	filename := buildFilename(header.Filename)
	directory := filepath.Join(s.baseDir, category)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, err
	}

	destinationPath := filepath.Join(directory, filename)
	destination, err := os.Create(destinationPath)
	if err != nil {
		return nil, err
	}
	defer destination.Close()

	writtenBytes, err := io.Copy(destination, source)
	if err != nil {
		return nil, err
	}

	return &StoredFile{
		RelativePath: "/uploads/" + category + "/" + filename,
		Filename:     filename,
		Size:         writtenBytes,
		MIMEType:     mimeType,
	}, nil
}

func buildFilename(original string) string {
	extension := filepath.Ext(original)
	if extension == "" {
		extension = ".bin"
	}
	return uuid.NewString() + extension
}
