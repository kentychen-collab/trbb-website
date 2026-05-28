package config

import (
	"os"
	"strconv"
)

type Config struct {
	App   AppConfig
	DB    DBConfig
	Redis RedisConfig
	Minio MinioConfig
	Log   LogConfig
	Third ThirdPartyConfig
}

type AppConfig struct {
	Env       string
	Port      string
	SecretKey string
}

type DBConfig struct {
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}

type MinioConfig struct {
	Endpoint    string
	AccessKey   string
	SecretKey   string
	UseSSL      bool
	BucketPublic  string
	BucketPrivate string
	BucketImages  string
	ExternalURL string
}

type LogConfig struct {
	Level string
	Dir   string
}

// ThirdPartyConfig holds all third-party integration configs.
// Add/extend as new providers are integrated.
type ThirdPartyConfig struct {
	Payment  PaymentConfig
	Garmin   GarminConfig
	Line     LineConfig
	Google   GoogleConfig
	Facebook FacebookConfig
	Email    EmailConfig
	SMS      SMSConfig
}

type PaymentConfig struct {
	Provider    string // ecpay | stripe | linepay
	MerchantID  string
	HashKey     string
	HashIV      string
	ReturnURL   string
	NotifyURL   string
}

type GarminConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	AuthURL      string
	TokenURL     string
	APIBase      string
}

type LineConfig struct {
	NotifyToken  string
	ChannelID    string
	ChannelSecret string
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type FacebookConfig struct {
	AppID       string
	AppSecret   string
	RedirectURI string
}

type EmailConfig struct {
	SendGridAPIKey string
	From           string
	FromName       string
}

type SMSConfig struct {
	Provider string // every8d | mitake
	UID      string
	PWD      string
	From     string
}

func Load() *Config {
	return &Config{
		App: AppConfig{
			Env:       getEnv("APP_ENV", "development"),
			Port:      getEnv("APP_PORT", "8080"),
			SecretKey: getEnv("APP_SECRET_KEY", "change-me"),
		},
		DB: DBConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "3306"),
			Name:            getEnv("DB_NAME", "trbb_pro"),
			User:            getEnv("DB_USER", "trbb_pro"),
			Password:        getEnv("DB_PASSWORD", ""),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 300),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
			PoolSize: getEnvInt("REDIS_POOL_SIZE", 10),
		},
		Minio: MinioConfig{
			Endpoint:      getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:     getEnv("MINIO_ACCESS_KEY", ""),
			SecretKey:     getEnv("MINIO_SECRET_KEY", ""),
			UseSSL:        getEnv("MINIO_USE_SSL", "false") == "true",
			BucketPublic:  getEnv("MINIO_BUCKET_PUBLIC", "trbb-public"),
			BucketPrivate: getEnv("MINIO_BUCKET_PRIVATE", "trbb-private"),
			BucketImages:  getEnv("MINIO_BUCKET_IMAGES", "images"),
			ExternalURL:   getEnv("MINIO_EXTERNAL_URL", ""),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
			Dir:   getEnv("LOG_DIR", "./logs"),
		},
		Third: ThirdPartyConfig{
			Payment: PaymentConfig{
				Provider:   getEnv("PAYMENT_PROVIDER", "ecpay"),
				MerchantID: getEnv("PAYMENT_MERCHANT_ID", ""),
				HashKey:    getEnv("PAYMENT_HASH_KEY", ""),
				HashIV:     getEnv("PAYMENT_HASH_IV", ""),
				ReturnURL:  getEnv("PAYMENT_RETURN_URL", ""),
				NotifyURL:  getEnv("PAYMENT_NOTIFY_URL", ""),
			},
			Garmin: GarminConfig{
				ClientID:     getEnv("GARMIN_CLIENT_ID", ""),
				ClientSecret: getEnv("GARMIN_CLIENT_SECRET", ""),
				RedirectURI:  getEnv("GARMIN_REDIRECT_URI", ""),
				AuthURL:      getEnv("GARMIN_AUTH_URL", ""),
				TokenURL:     getEnv("GARMIN_TOKEN_URL", ""),
				APIBase:      getEnv("GARMIN_API_BASE", ""),
			},
			Line: LineConfig{
				NotifyToken:   getEnv("LINE_NOTIFY_TOKEN", ""),
				ChannelID:     getEnv("LINE_CHANNEL_ID", ""),
				ChannelSecret: getEnv("LINE_CHANNEL_SECRET", ""),
			},
			Google: GoogleConfig{
				ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
				RedirectURI:  getEnv("GOOGLE_REDIRECT_URI", ""),
			},
			Facebook: FacebookConfig{
				AppID:       getEnv("FB_APP_ID", ""),
				AppSecret:   getEnv("FB_APP_SECRET", ""),
				RedirectURI: getEnv("FB_REDIRECT_URI", ""),
			},
			Email: EmailConfig{
				SendGridAPIKey: getEnv("SENDGRID_API_KEY", ""),
				From:           getEnv("EMAIL_FROM", ""),
				FromName:       getEnv("EMAIL_FROM_NAME", "TRBB"),
			},
			SMS: SMSConfig{
				Provider: getEnv("SMS_PROVIDER", "every8d"),
				UID:      getEnv("SMS_UID", ""),
				PWD:      getEnv("SMS_PWD", ""),
				From:     getEnv("SMS_FROM", "TRBB"),
			},
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
