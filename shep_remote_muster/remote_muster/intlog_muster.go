package main

import (
	"fmt"
	"os"
	"strconv"
	"encoding/csv"
	"time"
)

/*********************************************/

var intlog_cols []string = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
				    "instructions", "cycles", "ref_cycles", "llc_miss", 
				    "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
var max_rows uint64 = 4096
var max_bytes uint64 = uint64(len(intlog_cols) * 64) * max_rows
var exp_timeout time.Duration = time.Second * 75
var mirror_ip string;

func (intlog_m *intlog_muster) init() {
	var core uint8
	for core = 0; core < intlog_m.ncores; core ++ {
		var max_size uint64 = max_rows
		mem_buff := make([][]uint64, max_size)
		c_str := strconv.Itoa(int(core))
		sheep_id := c_str + "-" + intlog_m.ip
		log_id := "log-" + c_str + "-" + intlog_m.ip
		log_c := log{id: log_id,
			     metrics: intlog_cols,
			     max_size: max_size,
			     mem_buff: &mem_buff,
			     kill_log_chan: make(chan bool, 1),
			     ready_buff_chan: make(chan bool, 1)}
		ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + intlog_m.ip
		ctrl_itr_id := "ctrl-itr-" + c_str + "-" + intlog_m.ip
		ctrl_dvfs := control{id: ctrl_dvfs_id, n_ip: intlog_m.ip, 
				     knob: "dvfs", value: 0xc00, dirty: false} 
		ctrl_itr := control{id: ctrl_itr_id, n_ip: intlog_m.ip,
				    knob: "itr-delay", value: 1, dirty: false}  
		intlog_m.pasture[sheep_id].logs[log_id] = &log_c
		intlog_m.pasture[sheep_id].controls[ctrl_dvfs_id] = &ctrl_dvfs
		intlog_m.pasture[sheep_id].controls[ctrl_itr_id] = &ctrl_itr
	}

	intlog_m.log_f_map = make(map[string]*os.File)
	intlog_m.log_reader_map = make(map[string]*csv.Reader)
	for sheep_id, _ := range(intlog_m.pasture) {
		core := intlog_m.pasture[sheep_id].core
		c_str := strconv.Itoa(int(core))
		log_fname := "/users/awadyn/shepherd_muster/shep_remote_muster/intlog_logs/" + c_str
		f, err := os.Create(log_fname)
		if err != nil { panic(err) }
		reader := csv.NewReader(f)
		reader.Comma = ' '
		intlog_m.log_f_map[sheep_id] = f
		intlog_m.log_reader_map[sheep_id] = reader
	}
}

/* This function associates each muster's sheep with its
   intlog data in /proc/ixgbe_stats/core/sheep.core
*/
func (intlog_m *intlog_muster) start_native_logger() {
	for sheep_id, _ := range(intlog_m.pasture) {
		core := intlog_m.pasture[sheep_id].core
		for log_id, _ := range(intlog_m.pasture[sheep_id].logs) { 
			go intlog_m.intlog_log(sheep_id, log_id, core) 
		}
	}
}

func (r_m *intlog_muster) cleanup() {
	for sheep_id, _ := range(r_m.pasture) {
		for log_id, _ := range(r_m.pasture[sheep_id].logs) {
			r_m.pasture[sheep_id].logs[log_id].kill_log_chan <- true
		}
		r_m.log_f_map[sheep_id].Close()
	}
}

/*****************/

func main() {
	n_ip := os.Args[1]
	mirror_ip = os.Args[2]
	n_cores, err := strconv.Atoi(os.Args[3])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_cores argument: %v\n", err)}
	pulse_server_port, err := strconv.Atoi(os.Args[4])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}
	log_server_port := os.Args[5]
	ctrl_server_port, err := strconv.Atoi(os.Args[6])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}
	coordinate_server_port := os.Args[7]

	m := muster{}
	m.init(n_ip, n_cores)

	r_m := remote_muster{muster: m}
	r_m.init(n_ip, n_cores, pulse_server_port, ctrl_server_port, log_server_port, coordinate_server_port)
	r_m.show()

	intlog_m := intlog_muster{remote_muster: r_m}
	intlog_m.init()

	intlog_m.start_native_logger()

	go intlog_m.start_pulser()
	go intlog_m.start_logger()
	go intlog_m.start_controller()
	go intlog_m.handle_new_ctrl()

	// cleanup
//	go test_m.wait_done()
//	<- test_m.exit_chan
	time.Sleep(exp_timeout)
	intlog_m.cleanup()
}


