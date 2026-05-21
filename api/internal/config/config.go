package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port   string
	AppEnv string

	DBHost string
	DBPort string
	DBName string
	DBUser string
	DBPass string

	RedisHost string
	RedisPort string

	JWTSecret      string
	JWTExpireHours int

	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string

	// 圖片上傳（本地目錄）
	UploadDir       string // 容器內路徑，如 /var/www/uploads
	UploadURLPrefix string // 對外 URL 前綴，如 /uploads

	FrontendURL string
	AdminURL    string

	AdminInitUsername string
	AdminInitEmail    string
	AdminInitPassword string
	AdminInitName     string

	ECPayMerchantID string
	ECPayHashKey    string
	ECPayHashIV     string
	ECPaySandbox    bool
	ECPayReturnURL  string

	GarminConsumerKey    string
	GarminConsumerSecret string
}

func Load() *Config {
	return &Config{
		Port:   getEnv("PORT", "8080"),
		AppEnv: getEnv("APP_ENV", "development"),

		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnv("DB_PORT", "3306"),
		DBName: getEnv("DB_NAME", "sports_platform"),
		DBUser: getEnv("DB_USER", "root"),
		DBPass: getEnv("DB_PASS", ""),

		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),

		JWTSecret:      getEnv("JWT_SECRET", "change-this-secret"),
		JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 168),

		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:    getEnv("MINIO_BUCKET", "sports-platform"),

		UploadDir:       getEnv("UPLOAD_DIR", "/var/www/uploads"),
		UploadURLPrefix: getEnv("UPLOAD_URL_PREFIX", "/uploads"),

		FrontendURL: getEnv("APP_FRONTEND_URL", "http://localhost"),
		AdminURL:    getEnv("APP_ADMIN_URL", "http://localhost/admin"),

		AdminInitUsername: getEnv("ADMIN_INIT_USERNAME", "superadmin"),
		AdminInitEmail:    getEnv("ADMIN_INIT_EMAIL", "admin@example.com"),
		AdminInitPassword: getEnv("ADMIN_INIT_PASSWORD", "ChangeMe@2025"),
		AdminInitName:     getEnv("ADMIN_INIT_NAME", "超級管理員"),

		ECPayMerchantID: getEnv("ECPAY_MERCHANT_ID", ""),
		ECPayHashKey:    getEnv("ECPAY_HASH_KEY", ""),
		ECPayHashIV:     getEnv("ECPAY_HASH_IV", ""),
		ECPaySandbox:    getEnv("ECPAY_IS_SANDBOX", "true") == "true",
		ECPayReturnURL:  getEnv("ECPAY_RETURN_URL", ""),

		GarminConsumerKey:    getEnv("GARMIN_CONSUMER_KEY", ""),
		GarminConsumerSecret: getEnv("GARMIN_CONSUMER_SECRET", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	var n int
	fmt.Sscanf(v, "%d", &n)
	if n == 0 {
		return defaultVal
	}
	return n
}
