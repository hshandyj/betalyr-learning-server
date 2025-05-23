package storage

import (
	"betalyr-learning-server/internal/config"
	"betalyr-learning-server/internal/pkg/logger"
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

// 全局R2客户端
var R2Client *s3.Client

// R2配置信息
var (
	R2Bucket    string
	R2PublicURL string
)

// InitializeR2 初始化R2连接
func InitializeR2(cfg *config.Config) error {
	// 检查必要配置
	if cfg.R2.Endpoint == "" || cfg.R2.AccessKeyID == "" || cfg.R2.SecretAccessKey == "" || cfg.R2.Bucket == "" {
		return fmt.Errorf("R2 settings are incomplete, please check R2_ENDPOINT, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_BUCKET environment variables")
	}

	// 创建凭据
	creds := credentials.NewStaticCredentialsProvider(
		cfg.R2.AccessKeyID,
		cfg.R2.SecretAccessKey,
		"")

	// 配置
	s3Config := aws.Config{
		Region:      "auto",
		Credentials: creds,
		// 设置HTTP客户端
		HTTPClient: &http.Client{
			Transport: &http.Transport{},
		},
	}

	// 创建S3客户端
	client := s3.NewFromConfig(s3Config, func(o *s3.Options) {
		o.EndpointResolver = s3.EndpointResolverFunc(
			func(region string, options s3.EndpointResolverOptions) (aws.Endpoint, error) {
				return aws.Endpoint{URL: cfg.R2.Endpoint}, nil
			})
	})

	// 设置全局变量
	R2Client = client
	R2Bucket = cfg.R2.Bucket
	R2PublicURL = cfg.R2.PublicURL

	// 测试连接
	_, err := R2Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		logger.Error("R2 connection test failed", zap.Error(err))
		return fmt.Errorf("R2 connection test failed: %w", err)
	}

	logger.Info("R2 storage connection successful",
		zap.String("endpoint", cfg.R2.Endpoint),
		zap.String("bucket", R2Bucket))

	return nil
}
