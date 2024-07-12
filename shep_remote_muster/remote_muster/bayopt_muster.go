package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"encoding/csv"
	"time"
	"io"
	"slices"
//	"bufio"
)

/*********************************************/

var intlog_cols []string = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
				    "instructions", "cycles", "ref_cycles", "llc_miss", 
				    "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
var joules_idx int = slices.Index(intlog_cols, "joules")
var timestamp_idx int = slices.Index(intlog_cols, "timestamp")

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
			     //do_log_chan: make(chan bool, 1),
			     do_log_chan: make(chan string, 1),
			     done_log_chan: make(chan bool, 1),
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

/* The following functions associate each sheep with its
   intlog data in /proc/ixgbe_stats/core/sheep.core
*/
func (bayopt_m *bayopt_muster) attach_native_logger(sheep_id string, core uint8) {
	c_str := strconv.Itoa(int(core))
	src_fname := "/proc/ixgbe_stats/core/" + c_str
	log_fname := "/users/awadyn/shepherd_muster/shep_remote_muster/intlog_logs/" + c_str

	var f *os.File
	var reader *csv.Reader
	var err error
	if bayopt_m.log_f_map[sheep_id] != nil { // log file has been previously open
		f = bayopt_m.log_f_map[sheep_id]
		reader = bayopt_m.log_reader_map[sheep_id]
	} else {
		f, err = os.Open(log_fname)
		if err != nil { panic(err) }
		reader = csv.NewReader(f)
		reader.Comma = ' '
		bayopt_m.log_f_map[sheep_id] = f
		bayopt_m.log_reader_map[sheep_id] = reader
	}

	// clear intlogger logs
	cmd := exec.Command("bash", "-c", "cat " + src_fname)
	if err := cmd.Run(); err != nil { panic(err) }
	// start appending intlogger logs to filesystem logs
	for {
		select {
		case <- bayopt_m.pasture[sheep_id].detach_native_logger:
			return
		default:
			cmd = exec.Command("bash", "-c", "cat " + src_fname + " >> " + log_fname)
			if err = cmd.Run(); err != nil { panic(err) }
			time.Sleep(time.Second)
		}
	}
}

//func (bayopt_m *bayopt_muster) start_native_logger() {
//	<- bayopt_m.hb_chan
//
//	for sheep_id, _ := range(bayopt_m.pasture) {
//		core := bayopt_m.pasture[sheep_id].core
//		go bayopt_m.access_native_logger(sheep_id, core)
//		for log_id, _ := range(bayopt_m.pasture[sheep_id].logs) { 
//			go bayopt_m.bayopt_log(sheep_id, log_id, core) 
//		}
//	}
//}

func (r_m *bayopt_muster) cleanup() {
	for sheep_id, _ := range(r_m.pasture) {
		for log_id, _ := range(r_m.pasture[sheep_id].logs) {
			r_m.pasture[sheep_id].logs[log_id].kill_log_chan <- true
		}
		r_m.log_f_map[sheep_id].Close()
	}
}

/*
   These functions implement the logging functionality of a bayopt_muster.
   Bayopt_muster is concerned with the total joules consumed and the overall
   99th latency of an execution. 
   These 2 quantities constitute a bayopt_muster log that must be synced with 
   the bayopt_muster local mirror.
   --> at which point the shepherd can run bayopt to choose different
       control settings.
   ** No log syncing will be done during execution.
   ** Upon completing an execution, bayopt_muster is signalled to sync logs.
*/

func do_log(shared_log *log, reader *csv.Reader) error {
	*shared_log.mem_buff = make([][]uint64, 0)
	var counter uint64 = 0
	for {
		switch {
		case counter < shared_log.max_size:
			var row []string
			var err error
			for {
				row, err = reader.Read()
				if err == io.EOF { 
					time.Sleep(time.Second * 2)
					continue
				}
				if err != nil { panic(err) }
				break
			}
			*shared_log.mem_buff = append(*shared_log.mem_buff, []uint64{})
			joules_val, _ := strconv.Atoi(row[joules_idx])
			(*shared_log.mem_buff)[counter] = append((*shared_log.mem_buff)[counter], uint64(joules_val))
			timestamp_val, _ := strconv.Atoi(row[timestamp_idx])
			(*shared_log.mem_buff)[counter] = append((*shared_log.mem_buff)[counter], uint64(timestamp_val))	
			counter ++
		case counter == shared_log.max_size:
			return nil
		}
	}
}

func (r_m *muster) sync_with_logger(sheep_id string, log_id string, core uint8, reader *csv.Reader, logger_func func(*log, *csv.Reader)error) error {
	shared_log := r_m.pasture[sheep_id].logs[log_id] 
	err := logger_func(shared_log, reader)
	switch {
	case err == nil:	// => logged one buff
		r_m.full_buff_chan <- []string{sheep_id, log_id}
		<- shared_log.ready_buff_chan
		*shared_log.mem_buff = make([][]uint64, shared_log.max_size)
	case err == io.EOF:	// => log reader at EOF
		// do nothing if nothing has been logged yet
		if len(*(shared_log.mem_buff)) == 0 { return io.EOF }
		// otherwise sync whatever has been logged with mirror
		r_m.full_buff_chan <- []string{sheep_id, log_id}
		<- shared_log.ready_buff_chan
	}
	return err
}

func (bayopt_m *bayopt_muster) bayopt_log(sheep_id string, log_id string, core uint8) {
	<- bayopt_m.hb_chan

	fmt.Printf("-- -- STARTING BAYOPT LOG FOR SHEEP %v \n", core)

	for {
		// sync all new log entries with local muster
		// --> sync_with_logger: syncs one mem_buff at a time with local muster
		// --> do_log: fills one mem_buff at a time from log file
		select {
		case cmd := <- bayopt_m.pasture[sheep_id].logs[log_id].do_log_chan:
			switch {
			case cmd == "start":
				// start communication with native logger
				go bayopt_m.attach_native_logger(sheep_id, core)
			case cmd == "stop":
				// stop communication with native logger
				bayopt_m.pasture[sheep_id].detach_native_logger <- true
			case cmd == "first":
				// get first instance from native logger
				f := bayopt_m.log_f_map[sheep_id]
				f.Seek(0, io.SeekStart)
				reader := bayopt_m.log_reader_map[sheep_id]
				var log_entry []string
				var err error
				for {
					log_entry, err = reader.Read()
					if err == io.EOF { 
						time.Sleep(time.Second)
						continue
					}
					if err != nil { panic(err) }
					break
				}
				fmt.Println(log_entry)
			case cmd == "last":
				// get last instance from native logger
				f := bayopt_m.log_f_map[sheep_id]
				f.Seek(0, io.SeekStart)
				reader := bayopt_m.log_reader_map[sheep_id]
				var tmp []string
				var log_entry []string
				var err error
				for {
					tmp, err = reader.Read()
					if err == io.EOF { 
						break
					}
					log_entry = tmp
				}
				fmt.Println(log_entry)
			default:
				fmt.Println("************ UNKNOWN BAYOPT_LOG COMMAND: ", cmd)
			}
//			err := r_m.sync_with_logger(sheep_id, log_id, core, reader, do_log)
//			if err == nil { 
//				r_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
//			}
		}
	}
}


