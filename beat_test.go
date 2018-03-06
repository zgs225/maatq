package maatq

// This example demostrate how to use a scheduler
// func Example_ServeLoop() {
// 	s := maatq.NewDefaultScheduler("127.0.0.1:6379", "")
// 	s.ServeLoop()
// 	fmt.Println("hello")
// 	// Output: hello
// }
//
// func TestDumpsAndLoads(t *testing.T) {
// 	s := NewDefaultScheduler("127.0.0.1:6379", "")
// 	for i := 0; i < 3; i++ {
// 		s.Delay(&Message{
// 			Id:    fmt.Sprintf("ID(%d)", i+1),
// 			Event: "hello",
// 			Try:   3,
// 			Data:  "yuez",
// 		}, time.Duration(i+1)*time.Second)
//
// 		p, _ := NewPeriod(10)
// 		s.Period(&Message{
// 			Id:    fmt.Sprintf("ID(%d)", i+2),
// 			Event: "hello",
// 			Try:   3,
// 			Data:  "yuez",
// 		}, p)
//
// 		cron, _ := NewCrontab("* * * * *")
// 		s.Crontab(&Message{
// 			Id:    fmt.Sprintf("ID(%d)", i+3),
// 			Event: "hello",
// 			Try:   3,
// 			Data:  "yuez",
// 		}, cron)
// 	}
// 	if err := s.dumps(); err != nil {
// 		t.Error("Dumps error: ", err)
// 	}
// 	h, err := s.loads()
// 	if err != nil {
// 		t.Error("Loads error: ", err)
// 	}
// 	t.Log(h.Items)
// 	for j := 0; j < s.heap.Len(); j++ {
// 		if !reflect.DeepEqual((*h.Items)[j], (*s.heap.Items)[j]) {
// 			t.Error("Loads heap does not equals original heap: ", (*h.Items)[j], (*s.heap.Items)[j])
// 		}
// 	}
// }
