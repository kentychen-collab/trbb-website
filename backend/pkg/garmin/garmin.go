package garmin

// Garmin Connect API 使用 OAuth 1.0a
// 文件: https://developer.garmin.com/connect-iq/

const (
	BaseURL         = "https://connectapi.garmin.com"
	RequestTokenURL = "https://connectapi.garmin.com/oauth-service/oauth/request_token"
	AuthURL         = "https://connect.garmin.com/oauthConfirm"
	AccessTokenURL  = "https://connectapi.garmin.com/oauth-service/oauth/access_token"
	ActivitiesURL   = BaseURL + "/activitylist-service/activities/search/activities"
)

type Client struct {
	ConsumerKey    string
	ConsumerSecret string
}

func New(key, secret string) *Client {
	return &Client{ConsumerKey: key, ConsumerSecret: secret}
}

// NOTE: Garmin 使用 OAuth 1.0a，需要以下 Go 套件:
//   go get github.com/dghubble/oauth1
//
// 實作步驟:
// 1. GetRequestToken() -> 取得 oauth_token, oauth_token_secret
// 2. 導向 AuthURL?oauth_token=xxx (用戶授權)
// 3. Garmin 回呼帶 oauth_verifier
// 4. GetAccessToken(token, secret, verifier) -> 取得最終 access_token
// 5. 用 access_token 呼叫 Activities API

type Activity struct {
	ActivityID   int64   `json:"activityId"`
	ActivityName string  `json:"activityName"`
	ActivityType struct {
		TypeKey string `json:"typeKey"`
	} `json:"activityType"`
	StartTimeLocal string  `json:"startTimeLocal"`
	Duration       float64 `json:"duration"`
	Distance       float64 `json:"distance"`
	AverageHR      float64 `json:"averageHR"`
	Calories       float64 `json:"calories"`
}
