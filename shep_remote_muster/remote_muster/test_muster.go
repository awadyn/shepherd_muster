package main

//import (
//	"fmt"
//	"os"
//	"strconv"
//	"encoding/csv"
//	"time"
//)
//
///*********************************************/
//
//func (test_m *test_muster) init() {
//	test_m.log_f_map = make(map[string](map[string]*os.File))
//	test_m.log_reader_map = make(map[string](map[string]*csv.Reader))
//	test_m.done_log_map = make(map[string](map[string]chan bool))
//	for sheep_id, _ := range(test_m.pasture) {
//		test_m.log_f_map[sheep_id] = make(map[string]*os.File)
//		test_m.log_reader_map[sheep_id] = make(map[string]*csv.Reader)
//		test_m.done_log_map[sheep_id] = make(map[string]chan bool)
//	}
//}
//
//func (test_m *test_muster) start_native_logger() {
//	for sheep_id, _ := range(test_m.pasture) {
//		core := test_m.pasture[sheep_id].core
//		for log_id, _ := range(test_m.pasture[sheep_id].logs) { 
//			go test_m.simulate_remote_log(sheep_id, log_id, core) 
//		}
//	}
//}
//
//func (r_m *test_muster) cleanup() {
//	for sheep_id, _ := range(r_m.pasture) {
//		for _, f := range(r_m.log_f_map[sheep_id]) {
//			f.Close()
//		}
//	}
//}
//
///*****************/
//
//func main() {
//	n_ip := os.Args[1]
//	n_cores, err := strconv.Atoi(os.Args[2])
//	if err != nil {fmt.Printf("** ** ** ERROR: bad n_cores argument: %v\n", err)}
//	pulse_server_port, err := strconv.Atoi(os.Args[3])
//	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}
//	log_server_port := os.Args[4]
//	ctrl_server_port, err := strconv.Atoi(os.Args[5])
//	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}
//	coordinate_server_port := os.Args[6]
//
//	m := muster{}
//	m.init(n_ip, n_cores)
//
//	r_m := remote_muster{muster: m}
//	r_m.init(n_ip, n_cores, pulse_server_port, ctrl_server_port, log_server_port, coordinate_server_port)
//	r_m.show()
//
//	test_m := test_muster{remote_muster: r_m}
//	test_m.init()
//
////	test_m.start_native_logger()
//
//	go test_m.start_pulser()
////	go test_m.start_logger()
////	go test_m.start_controller()
////	go test_m.handle_new_ctrl()
//
//	// cleanup
////	go test_m.wait_done()
////	<- test_m.exit_chan
//	time.Sleep(time.Second * 45)
//	test_m.cleanup()
//}


