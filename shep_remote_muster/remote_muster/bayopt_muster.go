package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"encoding/csv"
	"time"
	"io"
//	"slices"
//	"bufio"
)

/*********************************************/

var intlog_cols []string = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
				    "instructions", "cycles", "ref_cycles", "llc_miss", 
				    "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
//var joules_idx int = slices.Index(intlog_cols, "joules")
//var timestamp_idx int = slices.Index(intlog_cols, "timestamp")
//var bayopt_cols []string = []string{"joules","latency"}
var max_rows uint64 = 1

func (bayopt_m *bayopt_muster) init() {
	var core uint8
	for core = 0; core < bayopt_m.ncores; core ++ {
		var max_size uint64 = max_rows
		mem_buff := make([][]uint64, max_size)
		c_str := strconv.Itoa(int(core))
		sheep_id := c_str + "-" + bayopt_m.ip
		log_id := "log-" + c_str + "-" + bayopt_m.ip
		log_c := log{id: log_id,
			     metrics: intlog_cols,
			     //metrics: bayopt_cols,
			     max_size: max_size,
			     mem_buff: &mem_buff,
			     kill_log_chan: make(chan bool, 1),
			     request_log_chan: make(chan string),
			     done_log_chan: make(chan bool, 1),
			     ready_buff_chan: make(chan bool, 1)}
		ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + bayopt_m.ip
		ctrl_dvfs := control{id: ctrl_dvfs_id, n_ip: bayopt_m.ip, 
				     knob: "dvfs", value: 0xc00, dirty: false} 
		ctrl_itr_id := "ctrl-itr-" + c_str + "-" + bayopt_m.ip
		ctrl_itr := control{id: ctrl_itr_id, n_ip: bayopt_m.ip,
				    knob: "itr-delay", value: 1, dirty: false}  
		bayopt_m.pasture[sheep_id].logs[log_id] = &log_c
		bayopt_m.pasture[sheep_id].controls[ctrl_dvfs_id] = &ctrl_dvfs
		bayopt_m.pasture[sheep_id].controls[ctrl_itr_id] = &ctrl_itr
	}

	bayopt_m.log_f_map = make(map[string]*os.File)
	bayopt_m.log_reader_map = make(map[string]*csv.Reader)
	for sheep_id, _ := range(bayopt_m.pasture) {
		core := bayopt_m.pasture[sheep_id].core
		c_str := strconv.Itoa(int(core))
		log_fname := "/users/awadyn/shepherd_muster/shep_remote_muster/intlog_logs/" + c_str
		f, err := os.Create(log_fname)
		if err != nil { panic(err) }
		bayopt_m.log_f_map[sheep_id] = f
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
			for i := range(len(intlog_cols)) {
				val, _ := strconv.Atoi(row[i])
				(*shared_log.mem_buff)[counter] = append((*shared_log.mem_buff)[counter], uint64(val))
			}
			counter ++
//			*shared_log.mem_buff = append(*shared_log.mem_buff, []uint64{})
//			joules_val, _ := strconv.Atoi(row[joules_idx])
//			(*shared_log.mem_buff)[counter] = append((*shared_log.mem_buff)[counter], uint64(joules_val))
//			timestamp_val, _ := strconv.Atoi(row[timestamp_idx])
//			(*shared_log.mem_buff)[counter] = append((*shared_log.mem_buff)[counter], uint64(timestamp_val))	
//			counter ++
		case counter == shared_log.max_size:
			return nil
		}
	}
}

func (r_m *muster) sync_with_logger(sheep_id string, log_id string, reader *csv.Reader, logger_func func(*log, *csv.Reader)error) error {
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

/********************************************/

func (bayopt_m *bayopt_muster) assign_log_files(sheep_id string) {
	c_str := strconv.Itoa(int(bayopt_m.pasture[sheep_id].core))
	log_fname := "/users/awadyn/shepherd_muster/shep_remote_muster/intlog_logs/" + c_str
	var f *os.File
	var err error
	if bayopt_m.log_f_map[sheep_id] != nil { // log file has been previously open
		return				 // do nothing..
	} else {
		f, err = os.Open(log_fname)
		if err != nil { panic(err) }
		bayopt_m.log_f_map[sheep_id] = f
	}
}


/* The following functions associate each sheep with its
   intlog data in /proc/ixgbe_stats/core/sheep.core
*/
func (bayopt_m *bayopt_muster) attach_native_logger(sheep_id string) {
	c_str := strconv.Itoa(int(bayopt_m.pasture[sheep_id].core))
	src_fname := "/proc/ixgbe_stats/core/" + c_str
	log_fname := "/users/awadyn/shepherd_muster/shep_remote_muster/intlog_logs/" + c_str

	cmd := exec.Command("bash", "-c", "cat " + src_fname)
	if err := cmd.Run(); err != nil { panic(err) }
	for {
		select {
		case <- bayopt_m.pasture[sheep_id].detach_native_logger:
			return
		default:
			cmd = exec.Command("bash", "-c", "cat " + src_fname + " >> " + log_fname)
			if err := cmd.Run(); err != nil { panic(err) }
			time.Sleep(time.Second * 2)
		}
	}
}

func (bayopt_m *bayopt_muster) log_manage(sheep_id string, log_id string) {
	fmt.Printf("-- -- STARTING LOG MANAGER FOR SHEEP %v - LOG %v \n", sheep_id, log_id)
	for {
		select {
		case cmd := <- bayopt_m.pasture[sheep_id].logs[log_id].request_log_chan:
			switch {
			case cmd == "start":
				// start communication with native logger
				bayopt_m.assign_log_files(sheep_id)
				go bayopt_m.attach_native_logger(sheep_id)
				bayopt_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
			case cmd == "stop":
				// stop communication with native logger
				bayopt_m.pasture[sheep_id].detach_native_logger <- true
				bayopt_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
			case cmd == "first":
				// get first instance from native logger
				go func() {
					sheep_id := sheep_id
					log_id := log_id
					f := bayopt_m.log_f_map[sheep_id]
					reader := csv.NewReader(f)
					reader.Comma = ' '
					f.Seek(0, io.SeekStart)
					err := bayopt_m.sync_with_logger(sheep_id, log_id, reader, do_log)
					if err == io.EOF {
						fmt.Println("************** FILE IS EMPTY *************", log_id) 
						bayopt_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
						return
					}
					if err != nil { panic(err) }
					bayopt_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
				} ()
			case cmd == "last":
				// get last instance from native logger
				go func() {
					sheep_id := sheep_id
					log_id := log_id
					f := bayopt_m.log_f_map[sheep_id]
					f.Seek(0, io.SeekStart)
					reader1 := csv.NewReader(f)
					reader2 := csv.NewReader(f)
					reader1.Comma = ' '
					reader2.Comma = ' '
					var err error
					_, err = reader1.Read()
					if err == io.EOF { 
						fmt.Println("************** FILE IS EMPTY *************", log_id) 
						bayopt_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
						return
					}
					if err != nil { panic(err) }
					// read until end of file to get last log entry
					for {
						_, err = reader1.Read()
						if err != nil { panic(err) }
						if err == io.EOF { break }
						reader2.Read()
					}
					err = bayopt_m.sync_with_logger(sheep_id, log_id, reader2, do_log)
					if err != nil { panic(err) }
					bayopt_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
				} ()
			default:
				fmt.Println("************ UNKNOWN BAYOPT_LOG COMMAND: ", cmd)
			}
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





