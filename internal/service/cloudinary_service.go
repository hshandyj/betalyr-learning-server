package service

import (
	"betalyr-learning-server/internal/config"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// CloudinaryService 定义Cloudinary服务接口
type CloudinaryService interface {
	SignRequest(paramsToSign map[string]interface{}) (string, error)
}

// cloudinaryService Cloudinary服务实现
type cloudinaryService struct {
	cfg *config.Config
}

// NewCloudinaryService 创建新的Cloudinary服务实例
func NewCloudinaryService(cfg *config.Config) CloudinaryService {
	return &cloudinaryService{
		cfg: cfg,
	}
}

// SignRequest 生成Cloudinary签名，模拟cloudinary.utils.api_sign_request功能
func (s *cloudinaryService) SignRequest(paramsToSign map[string]interface{}) (string, error) {
	// 按照Cloudinary的官方文档实现签名算法
	// 1. 收集所有参数（除了file和api_key）
	// 2. 按照字典顺序排序
	// 3. 将参数拼接成key=value&key=value格式
	// 4. 添加API Secret
	// 5. 计算SHA1哈希

	// 收集所有参数
	var keys []string
	for k := range paramsToSign {
		if k != "file" && k != "api_key" {
			keys = append(keys, k)
		}
	}

	// 按字典顺序排序
	sort.Strings(keys)

	// 构建签名字符串
	var parts []string
	for _, k := range keys {
		var valueStr string

		// 根据不同类型转换为字符串
		switch v := paramsToSign[k].(type) {
		case string:
			valueStr = v
		case int:
			valueStr = fmt.Sprintf("%d", v)
		case float64:
			valueStr = fmt.Sprintf("%d", int(v))
		case bool:
			if v {
				valueStr = "true"
			} else {
				valueStr = "false"
			}
		case []interface{}:
			var items []string
			for _, item := range v {
				items = append(items, fmt.Sprintf("%v", item))
			}
			valueStr = strings.Join(items, ",")
		default:
			valueStr = fmt.Sprintf("%v", v)
		}

		parts = append(parts, fmt.Sprintf("%s=%s", k, valueStr))
	}

	// 拼接成key=value&key=value格式
	signingStr := strings.Join(parts, "&")

	// 添加API Secret
	signingStr = signingStr + s.cfg.Cloudinary.APISecret

	// 计算SHA1哈希
	h := sha1.New()
	h.Write([]byte(signingStr))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature, nil
}
