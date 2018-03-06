package maatq

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

// 健康度检查报警代码

func NewEmailAlerter(reciever string) func(*checkItem) error {
	return func(i *checkItem) error {
		m := gomail.NewMessage()
		m.SetHeader("From", "no-reply@youplus.cc")
		m.SetHeader("To", reciever)
		m.SetHeader("Subject", fmt.Sprintf("maatq: %s DOWN!", i.name))
		m.SetBody("text/html", fmt.Sprintf("<b>%s</b> DOWN!", i.name))
		d := gomail.NewDialer("smtp.exmail.qq.com", 465, "no-reply@youplus.cc", "2nV66egYMqh7gRr8")
		return d.DialAndSend(m)
	}
}
