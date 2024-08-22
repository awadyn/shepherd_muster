package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"encoding/csv"
	"io"
	"time"
)

/*********************************************/

func (intlog_m *intlog_muster) init() {
	intlog_m.intlog_metrics = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
				    "instructions", "cycles", "ref_cycles", "llc_miss", 
				    "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
	intlog_m.buff_max_size = 4096
	var core uint8
	for core = 0; core < intlog_m.ncores; core ++ {
		mem_buff := make([][]uint64, intlog_m.buff_max_size)
		c_str := strconv.Itoa(int(core))
		sheep_id := c_str + "-" + intlog_m.ip
		log_id := "log-" + c_str + "-" + intlog_m.ip
		log_c := log{id: log_id,
			     metrics: intlog_m.intlog_metrics,
			     max_size: intlog_m.buff_max_size,
			     mem_buff: &mem_buff,
			     //kill_log_chan: make(chan bool, 1),
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

//func (r_m *intlog_muster) cleanup() {
//	for sheep_id, _ := range(r_m.pasture) {
//		for log_id, _ := range(r_m.pasture[sheep_id].logs) {
//			r_m.pasture[sheep_id].logs[log_id].kill_log_chan <- true
//		}
//		r_m.log_f_map[sheep_id].Close()
//	}
//}

/*****************/

func do_intlog_log(shared_log *log, reader *csv.Reader) error {
	*shared_log.mem_buff = make([][]uint64, 0)
	var counter uint64 = 0
	for {
		switch {
		case counter < shared_log.max_size:
			row, err := reader.Read()
			if err == io.EOF { 
				return err
			}
			if err != nil { panic(err) }
			*shared_log.mem_buff = append(*shared_log.mem_buff, []uint64{})
			for i := range(len(shared_log.metrics)) {
				val, _ := strconv.Atoi(row[i])
				(*shared_log.mem_buff)[counter] = append((*shared_log.mem_buff)[counter], uint64(val))
			}
			counter ++
		case counter == shared_log.max_size:
			return nil
		}
	}
}

//func (r_m *muster) sync_with_logger(sheep_id string, log_id string, core uint8, reader *csv.Reader, logger_func func(*log, *csv.Reader)error) error {
//	shared_log := r_m.pasture[sheep_id].logs[log_id] 
//	for {
//		err := logger_func(shared_log, reader)
//		switch {
//		case err == nil:	// => logged one buff
//			r_m.full_buff_chan <- []string{sheep_id, log_id}
//			<- shared_log.ready_buff_chan
//			*shared_log.mem_buff = make([][]uint64, shared_log.max_size)
//		case err == io.EOF:	// => log reader at EOF
//			// do nothing if nothing has been logged yet
//			if len(*(shared_log.mem_buff)) == 0 { return io.EOF }
//			// otherwise sync whatever has been logged with mirror
//			r_m.full_buff_chan <- []string{sheep_id, log_id}
//			<- shared_log.ready_buff_chan
//			return io.EOF
//		}
//	}
//}

func (r_m *intlog_muster) intlog_log(sheep_id string, log_id string, core uint8) {
	<- r_m.hb_chan

	c_str := strconv.Itoa(int(core))
	src_fname := "/proc/ixgbe_stats/core/" + c_str
	log_fname := "/users/awadyn/shepherd_muster/shep_remote_muster/intlog_logs/" + c_str
	fmt.Printf("-- -- CORE %v -- -- INTLOG :  %v \n", c_str, log_fname)
	
	var f *os.File
	var reader *csv.Reader
	var err error
	if r_m.log_f_map[sheep_id] != nil { // log file has been previously open
		f = r_m.log_f_map[sheep_id]
		reader = r_m.log_reader_map[sheep_id]
	} else {
		f, err = os.Open(log_fname)
		if err != nil { panic(err) }
		reader = csv.NewReader(f)
		reader.Comma = ' '
		r_m.log_f_map[sheep_id] = f
		r_m.log_reader_map[sheep_id] = reader
	}

	go func() {
		cmd := exec.Command("bash", "-c", "cat " + src_fname)
		if err := cmd.Run(); err != nil { panic(err) }
		for {
			cmd = exec.Command("bash", "-c", "cat " + src_fname + " >> " + log_fname)
			if err = cmd.Run(); err != nil { panic(err) }
			time.Sleep(time.Second)
		}
	} ()

	for {
		// sync all new log entries with local muster
		// --> sync_with_logger: syncs one mem_buff at a time with local muster
		// --> do_intlog_log: fills one mem_buff at a time from log file
		select {
//		case <- r_m.pasture[sheep_id].logs[log_id].kill_log_chan:
//			return
		default:
			err = r_m.sync_with_logger(sheep_id, log_id, reader, do_intlog_log, -1)
			if err == io.EOF { 
				//fmt.Printf("*** EOF  -  CORE :  %v  -  LOG :  %v ***\n", core, log_fname)
				time.Sleep(time.Second)
			}
		}
	}
}



