package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/zgs225/maatq"
)

var (
	try      = flag.Int("try", 0, "单个任务最大尝试次数")
	parallel = flag.Int("parallel", runtime.NumCPU(), "同时执行队列任务的并发数")
	addr     = flag.String("addr", "localhost:6379", "Redis主机和端口")
	password = flag.String("password", "", "Redis密码")
	debug    = flag.Bool("debug", true, "是否开启Debug")
	receiver = flag.String("receiver", "zhaigenshen@youplus.cc", "邮件告警人")
)

func SayHello(arg interface{}) (interface{}, error) {
	sval, err := maatq.AtoString(arg)
	if err != nil {
		return "", err
	}
	fmt.Printf("Hello %v\n", sval)
	return "ok", nil
}

func setLogger() {
	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	flag.Parse()
	setLogger()

	config := &maatq.BrokerOptions{
		Parallel:            *parallel,
		Addr:                *addr,
		Password:            *password,
		Try:                 *try,
		Queues:              []string{"sms"},
		Scheduler:           true,
		HealthCheckInterval: time.Second,
		AlertReceiver:       *receiver,
	}
	broker, err := maatq.NewBroker(config)
	if err != nil {
		log.Panic(err)
	}

	broker.AddEventHandler("hello", maatq.EventHandler(SayHello))
	broker.ServeLoop(":8181")
}
