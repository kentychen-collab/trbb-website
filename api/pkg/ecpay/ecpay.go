package ecpay

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// ECPay 綠界金流
type ECPay struct {
	MerchantID string
	HashKey    string
	HashIV     string
	IsSandbox  bool
}

func New(merchantID, hashKey, hashIV string, sandbox bool) *ECPay {
	return &ECPay{
		MerchantID: merchantID,
		HashKey:    hashKey,
		HashIV:     hashIV,
		IsSandbox:  sandbox,
	}
}

func (e *ECPay) BaseURL() string {
	if e.IsSandbox {
		return "https://payment-stage.ecpay.com.tw/Cashier/AioCheckOut/V5"
	}
	return "https://payment.ecpay.com.tw/Cashier/AioCheckOut/V5"
}

// GenerateCheckMacValue 產生檢查碼
func (e *ECPay) GenerateCheckMacValue(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}

	raw := fmt.Sprintf("HashKey=%s&%s&HashIV=%s",
		e.HashKey,
		strings.Join(parts, "&"),
		e.HashIV,
	)
	raw = url.QueryEscape(raw)
	raw = strings.ToLower(raw)

	h := sha256.New()
	h.Write([]byte(raw))
	return strings.ToUpper(fmt.Sprintf("%x", h.Sum(nil)))
}

// BuildOrderForm 建立訂單表單參數
func (e *ECPay) BuildOrderForm(orderNo string, amount int, desc, returnURL, notifyURL string) map[string]string {
	params := map[string]string{
		"MerchantID":        e.MerchantID,
		"MerchantTradeNo":   orderNo,
		"MerchantTradeDate": time.Now().Format("2006/01/02 15:04:05"),
		"PaymentType":       "aio",
		"TotalAmount":       fmt.Sprintf("%d", amount),
		"TradeDesc":         url.QueryEscape(desc),
		"ItemName":          desc,
		"ReturnURL":         notifyURL,
		"OrderResultURL":    returnURL,
		"ChoosePayment":     "ALL",
		"EncryptType":       "1",
	}
	params["CheckMacValue"] = e.GenerateCheckMacValue(params)
	return params
}
