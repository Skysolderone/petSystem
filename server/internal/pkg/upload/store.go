package upload

import "mime/multipart"

type Store interface {
	SaveImage(category string, header *multipart.FileHeader) (*StoredFile, error)
	SaveFile(category string, header *multipart.FileHeader) (*StoredFile, error)
}

type StoredFile struct {
	RelativePath string
	PublicURL    string
	Filename     string
	Size         int64
	MIMEType     string
}
