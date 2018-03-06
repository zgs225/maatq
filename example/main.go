package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"github.com/zgs225/maatq"
)

var (
	try      = flag.Int("try", 0, "单个任务最大尝试次数")
	parallel = flag.Int("parallel", runtime.NumCPU(), "同时执行队列任务的并发数")
	addr     = flag.String("addr", "localhost:6379", "Redis主机和端口")
	password = flag.String("password", "", "Redis密码")
	debug    = flag.Bool("debug", true, "是否开启Debug")
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

	config := &maatq.BrokerOptions{*parallel, *addr, *password, *try, []string{"default", "sms"}, true}
	broker, err := maatq.NewBroker(config)
	if err != nil {
		log.Panic(err)
	}

	c := make(chan int)

	broker.AddEventHandler("hello", maatq.EventHandler(SayHello))
	go broker.ServeLoop(":8181")

	u := uuid.New()
	m := &maatq.Message{
		Id:        u.String(),
		Event:     "hello",
		Timestamp: time.Now().Unix(),
		Try:       0,
		Data:      "Hello period",
	}
	// p, _ := maatq.NewPeriod(300)
	//
	// broker.Period(m, p)
	// time.Sleep(time.Second)
	// m.Data = "Hello cron....."
	// cron, _ := maatq.NewCrontab("*/2 * * * *")
	// broker.Crontab(m, cron)
	//
	// time.Sleep(time.Second)
	m.Data = "Hi cron....."
	cron, _ := maatq.NewCrontab("* * * * *")
	broker.Crontab(m, cron)

	if err := broker.Dumps(); err != nil {
		log.Error(err)
	}

	<-c
}
