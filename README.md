# SportsPlatform - 運動電商平台

## 技術棧
- 前台: HTML + Alpine.js + Tailwind CSS
- 後台: HTML + AdminLTE 3
- 後端: Golang + Gin
- 資料庫: MySQL 8.0
- 快取: Redis
- 檔案儲存: MinIO
- 容器: Docker Compose

## 快速啟動

```bash
# 1. 複製環境設定
cp .env.example .env
# 編輯 .env 填入必要的 key

# 2. 啟動所有服務
docker compose up -d

# 3. 確認服務狀態
docker compose ps

# 4. 查看後端 log
docker compose logs -f backend
```

## 服務端口
| 服務 | 端口 |
|------|------|
| 前台 | http://localhost |
| 後台 | http://localhost/admin |
| API  | http://localhost/api/v1 |
| MinIO Console | http://localhost:9001 |
| MySQL | localhost:3306 |

## 目錄結構
```
sports-platform/
├── backend/          # Golang API Server
│   ├── cmd/server/   # 進入點
│   ├── internal/     # 核心邏輯
│   │   ├── config/
│   │   ├── middleware/  # JWT, CORS
│   │   ├── router/      # 路由
│   │   ├── handler/     # HTTP Handler
│   │   ├── service/     # 業務邏輯
│   │   ├── repository/  # DB 操作
│   │   └── model/       # 資料結構
│   └── pkg/          # 工具套件
│       ├── jwt/
│       ├── ecpay/    # 綠界金流
│       ├── garmin/   # Garmin OAuth
│       └── upload/   # MinIO 上傳
├── frontend/         # 前台
├── admin/            # 後台
├── db/
│   └── init.sql      # MySQL Schema
├── nginx/conf.d/     # Nginx 設定
├── docker-compose.yml
└── .env.example
```

## 開發注意事項

### Garmin OAuth 1.0a
需安裝: `go get github.com/dghubble/oauth1`
Garmin 使用 OAuth 1.0a，**不是** OAuth 2.0

### 金流 (ECPay 綠界)
- 先用沙箱測試: ECPAY_IS_SANDBOX=true
- 測試用 MerchantID: 2000132
- callback URL 需為公開可訪問的 IP/domain

### 身份證加密
event_registrations.id_number 欄位需 AES-256-GCM 加密後存入

### 社群分享
- FB: 使用 og:meta tag + sharer.php (無需 API key)
- IG: 不支援直接分享 URL，引導用戶複製連結

## API 認證
所有需要登入的 API 在 Header 帶入:
```
Authorization: Bearer <JWT_TOKEN>
```
