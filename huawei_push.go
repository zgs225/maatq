package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
)

// 封装华为推送

var (
	HW_LOGIN_URL      = "https://login.vmall.com/oauth2/token"
	HW_PUSH_URL       = "https://api.vmall.com/rest.php"
	HW_MAMC_APPID     = "10721396"
	HW_MAMC_APPSECRET = "94c5bec53a3099b503ae8417aa97b47e"
	HWMamcPusherPtr   = &HuaweiPusher{HW_MAMC_APPID, HW_MAMC_APPSECRET, &AuthToken{}}
)

// Type AuthToken
type AuthToken struct {
	Token     string
	ExpiresAt time.Time
}

func (t *AuthToken) Valid() bool {
	if len(t.Token) == 0 {
		return false
	}
	return t.ExpiresAt.After(time.Now())
}

// End type

// Type HuaweiMessage
type HuaweiMessage map[string]interface{}

func (hwm HuaweiMessage) has(key string) bool {
	_, ok := hwm[key]
	return ok
}

func (hwm HuaweiMessage) LoadPostData(data url.Values) {
	for key, val := range hwm {
		sval, err := AtoString(val)
		if err == nil {
			data.Set(key, sval)
		}
	}

	if !hwm.has("nsp_fmt") {
		data.Set("nsp_fmt", "JSON")
	}

	if !hwm.has("nsp_ts") {
		data.Set("nsp_ts", strconv.FormatInt(time.Now().Unix()*1000, 10))
	}

	if !hwm.has("nsp_svc") {
		data.Set("nsp_svc", "openpush.openapi.notification_send")
	}

	if !hwm.has("push_type") {
		data.Set("push_type", "3")
	}
}

// End type

// Type HuaweiPusher
type HuaweiPusher struct {
	AppId     string
	AppSecret string
	TokenPtr  *AuthToken
}

func (p *HuaweiPusher) Push(message HuaweiMessage) (interface{}, error) {
	if !p.TokenPtr.Valid() {
		p.Login()
	}

	formData := url.Values{}
	formData.Set("access_token", p.TokenPtr.Token)
	formData.Set("dev_app_id", p.AppId)
	message.LoadPostData(formData)
	log.Debug(formData)

	var body interface{}
	resp, err := http.PostForm(HW_PUSH_URL, formData)
	if err != nil {
		return body, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}
	err = json.Unmarshal(bytes, &body)
	if err != nil {
		return body, err
	}

	if resp.StatusCode != 200 {
		errMsg := (body.(map[string]interface{})["error_description"]).(string)
		return body, errors.New(errMsg)
	}

	return body, nil
}

// 通过华为接口登录
func (p *HuaweiPusher) Login() error {
	if err := p.check(); err != nil {
		return err
	}

	var formData = url.Values{}
	formData.Set("grant_type", "client_credentials")
	formData.Set("client_id", p.AppId)
	formData.Set("client_secret", p.AppSecret)

	resp, err := http.PostForm(HW_LOGIN_URL, formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var body map[string]interface{}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		errMsg := (body["error_description"]).(string)
		return errors.New(errMsg)
	}
	log.Debug(body)

	expires := time.Duration(int64((body["expires_in"]).(float64)))
	now := time.Now()
	p.TokenPtr.Token = (body["access_token"]).(string)
	p.TokenPtr.ExpiresAt = now.Add(expires * time.Second)

	return nil
}

func (p *HuaweiPusher) check() error {
	if len(p.AppId) == 0 || len(p.AppSecret) == 0 {
		return errors.New("AppId and AppSecret required")
	}
	return nil
}

// End type

func MamcHuaweiPushTask(arg interface{}) (interface{}, error) {
	msg, ok := arg.(map[string]interface{})
	if !ok {
		return nil, errors.New("参数错误")
	}
	message := HuaweiMessage(msg)
	log.Debug(message)
	return HWMamcPusherPtr.Push(message)
}
