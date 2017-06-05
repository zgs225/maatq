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

	group, err := maatq.NewWorkerGroup(&maatq.GroupOptions{*parallel, *addr, *password, *try, []string{"default", "sms"}})
	if err != nil {
		log.Panic(err)
	}

	c := make(chan int)

	group.AddEventHandler("hello", maatq.EventHandler(SayHello))
	go group.ServeLoop()

	s := maatq.NewDefaultScheduler(*addr, *password)

	for i := 0; i < 10; i++ {
		u := uuid.New()
		m := &maatq.Message{
			Id:        u.String(),
			Event:     "hello",
			Timestamp: time.Now().Unix(),
			Try:       0,
			Data:      fmt.Sprintf("yuez %d", i),
		}

		var t time.Duration
		if i%3 == 0 {
			t = time.Second
		} else {
			t = time.Minute
		}

		d := time.Duration(i+1) * t
		s.Delay(m, d)
	}

	go s.ServeLoop()

	<-c
}
