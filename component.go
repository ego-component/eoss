package awos

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const PackageName = "component.awos"

// Component interface
type Component interface {
	WithContext(ctx context.Context) Component
	Get(key string, options ...GetOptions) (string, error)
	GetBytes(key string, options ...GetOptions) ([]byte, error)
	GetAsReader(key string, options ...GetOptions) (io.ReadCloser, error)
	GetWithMeta(key string, attributes []string, options ...GetOptions) (io.ReadCloser, map[string]string, error)
	Put(key string, reader io.ReadSeeker, meta map[string]string, options ...PutOptions) error
	Del(key string) error
	DelMulti(keys []string) error
	Head(key string, meta []string) (map[string]string, error)
	ListObject(key string, prefix string, marker string, maxKeys int, delimiter string) ([]string, error)
	SignURL(key string, expired int64) (string, error)
	GetAndDecompress(key string) (string, error)
	GetAndDecompressAsReader(key string) (io.ReadCloser, error)
	CompressAndPut(key string, reader io.ReadSeeker, meta map[string]string, options ...PutOptions) error
	Range(key string, offset int64, length int64) (io.ReadCloser, error)
	Exists(key string) (bool, error)
}

func newComponent(cfg *config) (Component, error) {
	storageType := strings.ToLower(cfg.StorageType)

	if storageType == StorageTypeOSS {
		client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
		if err != nil {
			return nil, err
		}

		var ossClient *OSS
		if cfg.Shards != nil && len(cfg.Shards) > 0 {
			buckets := make(map[string]*oss.Bucket)
			for _, v := range cfg.Shards {
				bucket, err := client.Bucket(cfg.Bucket + "-" + v)
				if err != nil {
					return nil, err
				}
				for i := 0; i < len(v); i++ {
					buckets[strings.ToLower(v[i:i+1])] = bucket
				}
			}

			ossClient = &OSS{
				Shards: buckets,
			}
		} else {
			bucket, err := client.Bucket(cfg.Bucket)
			if err != nil {
				return nil, err
			}

			ossClient = &OSS{
				Bucket: bucket,
			}
		}

		return ossClient, nil
	} else if storageType == StorageTypeS3 {
		var config *aws.Config

		// use minio
		if cfg.S3ForcePathStyle {
			config = &aws.Config{
				Region:           aws.String(cfg.Region),
				DisableSSL:       aws.Bool(!cfg.SSL),
				Credentials:      credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.AccessKeySecret, ""),
				Endpoint:         aws.String(cfg.Endpoint),
				S3ForcePathStyle: aws.Bool(true),
			}
		} else {
			config = &aws.Config{
				Region:      aws.String(cfg.Region),
				DisableSSL:  aws.Bool(!cfg.SSL),
				Credentials: credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.AccessKeySecret, ""),
			}
			if cfg.Endpoint != "" {
				config.Endpoint = aws.String(cfg.Endpoint)
			}
		}

		config.HTTPClient = &http.Client{
			Timeout: time.Second * time.Duration(cfg.S3HttpTimeoutSecs),
		}
		if cfg.EnableTraceInterceptor {
			if cfg.EnableClientTrace {
				config.HTTPClient.Transport = otelhttp.NewTransport(http.DefaultTransport,
					otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
						return otelhttptrace.NewClientTrace(ctx)
					}))
			} else {
				config.HTTPClient.Transport = otelhttp.NewTransport(http.DefaultTransport)
			}
		}
		service := s3.New(session.Must(session.NewSession(config)))

		var s3Client *S3
		if cfg.Shards != nil && len(cfg.Shards) > 0 {
			buckets := make(map[string]string)
			for _, v := range cfg.Shards {
				for i := 0; i < len(v); i++ {
					buckets[strings.ToLower(v[i:i+1])] = cfg.Bucket + "-" + v
				}
			}
			s3Client = &S3{
				ShardsBucket: buckets,
				Client:       service,
			}
		} else {
			s3Client = &S3{
				BucketName: cfg.Bucket,
				Client:     service,
			}
		}

		return s3Client, nil
	} else {
		return nil, fmt.Errorf("unknown StorageType:\"%s\", only supports oss,s3", cfg.StorageType)
	}
}
