package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
)

// 封装小米推送V3

var (
	MI_REGID_URL        = "https://api.xmpush.xiaomi.com/v3/message/regid"
	MI_ALIAS_URL        = "https://api.xmpush.xiaomi.com/v3/message/alias"
	MI_TOPIC_URL        = "https://api.xmpush.xiaomi.com/v3/message/topic"
	MI_MULTI_TOPICS_URL = "https://api.xmpush.xiaomi.com/v3/message/multi_topic"
	MI_ALL_URL          = "https://api.xmpush.xiaomi.com/v3/message/all"
	MI_MAMC_APP_KEY     = "oBdZK1ex4eEq6wa1bpMGdQ=="
	RequiredErr         = errors.New("参数不足")
)

// Type MiMessage
type MiMessage map[string]interface{}

func (mc MiMessage) has(key string) bool {
	_, ok := mc[key]
	return ok
}

func (mc MiMessage) GetRestrictedPackageName() string {
	v, ok := mc["restricted_package_name"]
	if !ok {
		return ""
	}
	return v.(string)
}

func (mc MiMessage) checkRequired() error {
	var fields = []string{
		"title",
		"description",
		"restricted_package_name",
		"payload",
	}

	for _, f := range fields {
		if !mc.has(f) {
			return RequiredErr
		}
	}

	return nil
}

func (mc MiMessage) Check() error {
	if err := mc.checkRequired(); err != nil {
		return err
	}

	return nil
}

func (mc MiMessage) GetPushUrl() string {
	if mc.has("registration_id") {
		return MI_REGID_URL
	}

	if mc.has("alias") {
		return MI_ALIAS_URL
	}

	if mc.has("topic") {
		return MI_TOPIC_URL
	}

	if mc.has("topics") {
		return MI_MULTI_TOPICS_URL
	}

	return MI_ALL_URL
}

func (mc MiMessage) LoadPostData(data url.Values) {
	for key, val := range mc {
		sval, err := AtoString(val)
		if err == nil {
			data.Set(key, sval)
		}
	}

	if !mc.has("pass_through") {
		data.Set("pass_through", "0")
	}

	if !mc.has("notify_type") {
		data.Set("notify_type", "-1")
	}
}

// End Type

// Type MiPusher
type MiPusher struct {
	Key string
}

func (p *MiPusher) Push(miMsg MiMessage) (string, error) {
	formData := url.Values{}
	miMsg.LoadPostData(formData)
	log.Debug(formData)
	req, err := http.NewRequest("POST", miMsg.GetPushUrl(), bytes.NewBufferString(formData.Encode()))

	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("key=%s", p.Key))

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(bytes[:]), nil
}

// End Type

func MiPushTask(arg interface{}) (interface{}, error) {
	miMsg := MiMessage(arg.(map[string]interface{}))
	log.Debug(miMsg)

	if err := miMsg.Check(); err != nil {
		return nil, err
	}

	var pusher = &MiPusher{MI_MAMC_APP_KEY}
	return pusher.Push(miMsg)
}
