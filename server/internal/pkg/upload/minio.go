package upload

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOConfig struct {
	Endpoint      string
	AccessKey     string
	SecretKey     string
	UseSSL        bool
	Bucket        string
	PublicBaseURL string
}

type MinIOStore struct {
	client        *minio.Client
	bucket        string
	endpoint      string
	useSSL        bool
	publicBaseURL string
	ensureBucket  sync.Once
	ensureErr     error
}

func NewMinIOStore(cfg MinIOConfig) (*MinIOStore, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	bucket := strings.TrimSpace(cfg.Bucket)
	if endpoint == "" {
		return nil, fmt.Errorf("minio endpoint is required")
	}
	if bucket == "" {
		return nil, fmt.Errorf("minio bucket is required")
	}

	client, err := minio.New(strings.TrimSpace(cfg.Endpoint), &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinIOStore{
		client:        client,
		bucket:        bucket,
		endpoint:      endpoint,
		useSSL:        cfg.UseSSL,
		publicBaseURL: strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/"),
	}, nil
}

func (s *MinIOStore) SaveImage(category string, header *multipart.FileHeader) (*StoredFile, error) {
	return s.save(category, header, true)
}

func (s *MinIOStore) SaveFile(category string, header *multipart.FileHeader) (*StoredFile, error) {
	return s.save(category, header, false)
}

func (s *MinIOStore) save(category string, header *multipart.FileHeader, imagesOnly bool) (*StoredFile, error) {
	if err := s.ensureBucketExists(context.Background()); err != nil {
		return nil, err
	}

	file, err := header.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	readBytes, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
	}

	mimeType := http.DetectContentType(buffer[:readBytes])
	if imagesOnly && !strings.HasPrefix(mimeType, "image/") {
		return nil, fmt.Errorf("unsupported image mime type: %s", mimeType)
	}

	filename := buildFilename(header.Filename)
	objectName := path.Join(category, filename)
	size := header.Size
	if _, err := s.client.PutObject(context.Background(), s.bucket, objectName, file, size, minio.PutObjectOptions{
		ContentType: mimeType,
	}); err != nil {
		return nil, err
	}

	return &StoredFile{
		RelativePath: path.Join(s.bucket, objectName),
		PublicURL:    s.resolvePublicURL(objectName),
		Filename:     filename,
		Size:         size,
		MIMEType:     mimeType,
	}, nil
}

func (s *MinIOStore) ensureBucketExists(ctx context.Context) error {
	s.ensureBucket.Do(func() {
		exists, err := s.client.BucketExists(ctx, s.bucket)
		if err != nil {
			s.ensureErr = err
			return
		}
		if !exists {
			s.ensureErr = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
			if s.ensureErr != nil {
				return
			}
		}
		s.ensureErr = s.client.SetBucketPolicy(ctx, s.bucket, fmt.Sprintf(`{
  "Version":"2012-10-17",
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":{"AWS":["*"]},
      "Action":["s3:GetBucketLocation","s3:ListBucket"],
      "Resource":["arn:aws:s3:::%s"]
    },
    {
      "Effect":"Allow",
      "Principal":{"AWS":["*"]},
      "Action":["s3:GetObject"],
      "Resource":["arn:aws:s3:::%s/*"]
    }
  ]
}`, s.bucket, s.bucket))
	})
	return s.ensureErr
}

func (s *MinIOStore) resolvePublicURL(objectName string) string {
	if s.publicBaseURL != "" {
		return s.publicBaseURL + "/" + s.bucket + "/" + objectName
	}

	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	return scheme + "://" + s.endpoint + "/" + s.bucket + "/" + path.Clean(objectName)
}
