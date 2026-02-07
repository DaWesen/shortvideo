package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"shortvideo/pkg/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage interface {
	//上传文件
	Upload(ctx context.Context, bucket, object string, data io.Reader, size int64, contentType string) (string, error)

	//下载文件
	Download(ctx context.Context, bucket, object string) (io.Reader, error)

	//删除文件
	Delete(ctx context.Context, bucket, object string) error

	//获取预签名URL
	GetPresignedURL(ctx context.Context, bucket, object string, expiry time.Duration) (string, error)

	//检查文件是否存在
	Exists(ctx context.Context, bucket, object string) (bool, error)

	//创建桶
	CreateBucket(ctx context.Context, bucket string) error

	//关闭连接
	Close() error
}

type MinioStorage struct {
	client    *minio.Client
	endpoint  string
	bucket    string
	useSSL    bool
	accessKey string
	secretKey string
	baseURL   string
}

var (
	storageInstance *MinioStorage
	storageOnce     sync.Once
)

func NewMinioStorage() Storage {
	storageOnce.Do(func() {
		minioConfig := config.Get().Minio
		storageInstance = &MinioStorage{
			endpoint:  minioConfig.Endpoint,
			bucket:    minioConfig.Bucket,
			useSSL:    minioConfig.UseSSL,
			accessKey: minioConfig.AccessKey,
			secretKey: minioConfig.SecretKey,
		}

		var err error
		storageInstance.client, err = storageInstance.initClient()
		if err != nil {
			log.Printf("初始化MinIO客户端失败: %v", err)
			return
		}

		storageInstance.baseURL = storageInstance.buildBaseURL()

		if err := storageInstance.CreateBucket(context.Background(), minioConfig.Bucket); err != nil {
			log.Printf("创建MinIO桶失败: %v", err)
		}
	})
	return storageInstance
}

func (s *MinioStorage) initClient() (*minio.Client, error) {
	client, err := minio.New(s.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.accessKey, s.secretKey, ""),
		Secure: s.useSSL,
	})
	if err != nil {
		return nil, err
	}

	log.Printf("MinIO客户端初始化成功: %s", s.endpoint)
	return client, nil
}

func (s *MinioStorage) buildBaseURL() string {
	protocol := "http"
	if s.useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s", protocol, s.endpoint, s.bucket)
}

func (s *MinioStorage) Upload(ctx context.Context, bucket, object string, data io.Reader, size int64, contentType string) (string, error) {
	if bucket == "" {
		bucket = s.bucket
	}

	_, err := s.client.PutObject(ctx, bucket, object, data, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		log.Printf("上传文件失败 %s/%s: %v", bucket, object, err)
		return "", err
	}

	fileURL := fmt.Sprintf("%s/%s", s.baseURL, object)
	log.Printf("文件上传成功: %s", fileURL)
	return fileURL, nil
}

func (s *MinioStorage) Download(ctx context.Context, bucket, object string) (io.Reader, error) {
	if bucket == "" {
		bucket = s.bucket
	}

	file, err := s.client.GetObject(ctx, bucket, object, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("下载文件失败 %s/%s: %v", bucket, object, err)
		return nil, err
	}

	return file, nil
}

func (s *MinioStorage) Delete(ctx context.Context, bucket, object string) error {
	if bucket == "" {
		bucket = s.bucket
	}

	err := s.client.RemoveObject(ctx, bucket, object, minio.RemoveObjectOptions{})
	if err != nil {
		log.Printf("删除文件失败 %s/%s: %v", bucket, object, err)
		return err
	}

	log.Printf("文件删除成功: %s/%s", bucket, object)
	return nil
}

func (s *MinioStorage) GetPresignedURL(ctx context.Context, bucket, object string, expiry time.Duration) (string, error) {
	if bucket == "" {
		bucket = s.bucket
	}

	if expiry == 0 {
		expiry = 7 * 24 * time.Hour
	}

	presignedURL, err := s.client.PresignedGetObject(ctx, bucket, object, expiry, nil)
	if err != nil {
		log.Printf("获取预签名URL失败 %s/%s: %v", bucket, object, err)
		return "", err
	}

	return presignedURL.String(), nil
}

func (s *MinioStorage) Exists(ctx context.Context, bucket, object string) (bool, error) {
	if bucket == "" {
		bucket = s.bucket
	}

	_, err := s.client.StatObject(ctx, bucket, object, minio.StatObjectOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "object does not exist") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *MinioStorage) CreateBucket(ctx context.Context, bucket string) error {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}

	if !exists {
		err = s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}

		// 设置桶策略为公开读取
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Action": ["s3:GetObject"],
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Resource": ["arn:aws:s3:::%s/*"],
					"Sid": ""
				}
			]
		}`, bucket)

		err = s.client.SetBucketPolicy(ctx, bucket, policy)
		if err != nil {
			log.Printf("设置桶策略失败: %v", err)
		}

		log.Printf("MinIO桶创建成功: %s", bucket)
	} else {
		log.Printf("MinIO桶已存在: %s", bucket)
	}

	return nil
}

func (s *MinioStorage) Close() error {
	return nil
}

// 根据文件扩展名获取Content-Type
func GetContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".avi":
		return "video/x-msvideo"
	case ".mov":
		return "video/quicktime"
	case ".zip":
		return "application/zip"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}

// 生成唯一的对象名称
func GenerateObjectName(prefix, originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s/%d%s", prefix, timestamp, ext)
}
