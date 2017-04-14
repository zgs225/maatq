package main

import (
	"testing"
)

var pusherp = &HuaweiPusher{HW_MAMC_APPID, HW_MAMC_APPSECRET, &AuthToken{}}

func TestLogin(t *testing.T) {
	err := pusherp.Login()

	if err != nil {
		t.Error(err)
	} else {
		t.Log(pusherp.TokenPtr.Token)
		t.Log(pusherp.TokenPtr.ExpiresAt)
		t.Log("Login success")
	}
}
