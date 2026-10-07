// Package storage 对象存储(MinIO/S3)的薄封装。
package storage

import (
	"context"
	"io"
	"strings"

	"demo-shop/services/bff/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectStorage struct {
	client *minio.Client
	bucket string
	// publicBase 生成给前端的 URL 用,必须是**浏览器可达**的地址。
	// 与上传用的 Endpoint 可能不同(容器内 http://minio:9000 vs
	// 宿主机 http://localhost:9000),配错会让前端图片全部裂开。
	publicBase string
}

// New 建客户端并确保 bucket 存在且**匿名可读**。
//
// 匿名可读是刻意的:前端把 file_url 直接用于 img src,浏览器不带任何凭据。
// 生产若需隔离,应改成预签名 URL 并加有效期。
func New(c config.S3Config) (*ObjectStorage, error) {
	cli, err := minio.New(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.AccessKey, c.SecretKey, ""),
		Secure: c.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	exists, err := cli.BucketExists(ctx, c.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := cli.MakeBucket(ctx, c.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}
	// 幂等地设成公开读;已设过再设一次也无害。
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow",` +
		`"Principal":{"AWS":["*"]},"Action":["s3:GetObject"],` +
		`"Resource":["arn:aws:s3:::` + c.Bucket + `/*"]}]}`
	if err := cli.SetBucketPolicy(ctx, c.Bucket, policy); err != nil {
		return nil, err
	}

	return &ObjectStorage{
		client:     cli,
		bucket:     c.Bucket,
		publicBase: strings.TrimRight(c.PublicBaseURL, "/") + "/" + c.Bucket,
	}, nil
}

// Put 上传一个对象并返回其公开 URL。
func (s *ObjectStorage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return s.publicBase + "/" + key, nil
}
