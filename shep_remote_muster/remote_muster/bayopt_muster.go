package main

import (
//	"fmt"
	"os"
	"strconv"
	"encoding/csv"
	"time"
)

/*********************************************/

var bayopt_cols []string = []string{"joules","latency"}
var max_rows uint64 = 1
var max_bytes uint64 = uint64(len(bayopt_cols) * 64) * max_rows
var exp_timeout time.Duration = time.Second * 75
var mirror_ip string;

func (bayopt_m *bayopt_muster) init() {
	var core uint8
	for core = 0; core < bayopt_m.ncores; core ++ {
		var max_size uint64 = max_rows
		mem_buff := make([][]uint64, max_size)
		c_str := strconv.Itoa(int(core))
		sheep_id := c_str + "-" + bayopt_m.ip
		log_id := "log-" + c_str + "-" + bayopt_m.ip
		log_c := log{id: log_id,
			     metrics: bayopt_cols,
			     max_size: max_size,
			     mem_buff: &mem_buff,
			     kill_log_chan: make(chan bool, 1),
			     do_log_chan: make(chan bool, 1),
			     ready_buff_chan: make(chan bool, 1)}
		ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + bayopt_m.ip
		ctrl_dvfs := control{id: ctrl_dvfs_id, n_ip: bayopt_m.ip, 
				     knob: "dvfs", value: 0xc00, dirty: false} 
		//ctrl_itr_id := "ctrl-itr-" + c_str + "-" + bayopt_m.ip
		//ctrl_itr := control{id: ctrl_itr_id, n_ip: bayopt_m.ip,
		//		    knob: "itr-delay", value: 1, dirty: false}  
		bayopt_m.pasture[sheep_id].logs[log_id] = &log_c
		bayopt_m.pasture[sheep_id].controls[ctrl_dvfs_id] = &ctrl_dvfs
		//bayopt_m.pasture[sheep_id].controls[ctrl_itr_id] = &ctrl_itr
	}

	bayopt_m.log_f_map = make(map[string]*os.File)
	bayopt_m.log_reader_map = make(map[string]*csv.Reader)
	for sheep_id, _ := range(bayopt_m.pasture) {
		core := bayopt_m.pasture[sheep_id].core
		c_str := strconv.Itoa(int(core))
		log_fname := "/users/awadyn/shepherd_muster/shep_remote_muster/intlog_logs/" + c_str
		f, err := os.Create(log_fname)
		if err != nil { panic(err) }
		reader := csv.NewReader(f)
		reader.Comma = ' '
		bayopt_m.log_f_map[sheep_id] = f
		bayopt_m.log_reader_map[sheep_id] = reader
	}
}

/* This function associates each muster's sheep with its
   intlog data in /proc/ixgbe_stats/core/sheep.core
*/
func (bayopt_m *bayopt_muster) start_native_logger() {
	for sheep_id, _ := range(bayopt_m.pasture) {
		core := bayopt_m.pasture[sheep_id].core
		for log_id, _ := range(bayopt_m.pasture[sheep_id].logs) { 
			go bayopt_m.bayopt_log(sheep_id, log_id, core) 
		}
	}
}

func (r_m *bayopt_muster) cleanup() {
	for sheep_id, _ := range(r_m.pasture) {
		for log_id, _ := range(r_m.pasture[sheep_id].logs) {
			r_m.pasture[sheep_id].logs[log_id].kill_log_chan <- true
		}
		r_m.log_f_map[sheep_id].Close()
	}
}



