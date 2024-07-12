package main

//import (
//	"fmt"
//	"slices"
//	"strconv"
////	"os"
////	"os/exec"
//	"encoding/csv"
//	"io"
//	"time"
//)

/*********************************************************/

///*
//   These functions implement the logging functionality of a bayopt_muster.
//   Bayopt_muster is concerned with the total joules consumed and the overall
//   99th latency of an execution. 
//   These 2 quantities constitute a bayopt_muster log that must be synced with 
//   the bayopt_muster local mirror.
//   --> at which point the shepherd can run bayopt to choose different
//       control settings.
//   ** No log syncing will be done during execution.
//   ** Upon completing an execution, bayopt_muster is signalled to sync logs.
//*/
//
//var intlog_cols []string = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
//				    "instructions", "cycles", "ref_cycles", "llc_miss", 
//				    "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
//var joules_idx int = slices.Index(intlog_cols, "joules")
//var timestamp_idx int = slices.Index(intlog_cols, "timestamp")
//
//func do_log(shared_log *log, reader *csv.Reader) error {
//	*shared_log.mem_buff = make([][]uint64, 0)
//	var counter uint64 = 0
//	for {
//		switch {
//		case counter < shared_log.max_size:
//			var row []string
//			var err error
//			for {
//				row, err = reader.Read()
//				if err == io.EOF { 
//					time.Sleep(time.Second * 2)
//					continue
//				}
//				if err != nil { panic(err) }
//				break
//			}
//			*shared_log.mem_buff = append(*shared_log.mem_buff, []uint64{})
//			joules_val, _ := strconv.Atoi(row[joules_idx])
//			(*shared_log.mem_buff)[counter] = append((*shared_log.mem_buff)[counter], uint64(joules_val))
//			timestamp_val, _ := strconv.Atoi(row[timestamp_idx])
//			(*shared_log.mem_buff)[counter] = append((*shared_log.mem_buff)[counter], uint64(timestamp_val))	
//			counter ++
//		case counter == shared_log.max_size:
//			return nil
//		}
//	}
//}
//
//func (r_m *muster) sync_with_logger(sheep_id string, log_id string, core uint8, reader *csv.Reader, logger_func func(*log, *csv.Reader)error) error {
//	shared_log := r_m.pasture[sheep_id].logs[log_id] 
//	err := logger_func(shared_log, reader)
//	switch {
//	case err == nil:	// => logged one buff
//		r_m.full_buff_chan <- []string{sheep_id, log_id}
//		<- shared_log.ready_buff_chan
//		*shared_log.mem_buff = make([][]uint64, shared_log.max_size)
//	case err == io.EOF:	// => log reader at EOF
//		// do nothing if nothing has been logged yet
//		if len(*(shared_log.mem_buff)) == 0 { return io.EOF }
//		// otherwise sync whatever has been logged with mirror
//		r_m.full_buff_chan <- []string{sheep_id, log_id}
//		<- shared_log.ready_buff_chan
//	}
//	return err
//}
//
//func (r_m *bayopt_muster) bayopt_log(sheep_id string, log_id string, core uint8) {
////	<- r_m.hb_chan
//
////	c_str := strconv.Itoa(int(core))
////	src_fname := "/proc/ixgbe_stats/core/" + c_str
////	log_fname := "/users/awadyn/shepherd_muster/shep_remote_muster/intlog_logs/" + c_str
//	fmt.Printf("-- -- STARTING BAYOPT LOG FOR SHEEP %v \n", core)
//	
////	var f *os.File
////	var reader *csv.Reader
////	var err error
////	if r_m.log_f_map[sheep_id] != nil { // log file has been previously open
////		f = r_m.log_f_map[sheep_id]
////		reader = r_m.log_reader_map[sheep_id]
////	} else {
////		f, err = os.Open(log_fname)
////		if err != nil { panic(err) }
////		reader = csv.NewReader(f)
////		reader.Comma = ' '
////		r_m.log_f_map[sheep_id] = f
////		r_m.log_reader_map[sheep_id] = reader
////	}
//
////	go func() {
////		cmd := exec.Command("bash", "-c", "cat " + src_fname)
////		if err := cmd.Run(); err != nil { panic(err) }
////		for {
////			cmd = exec.Command("bash", "-c", "cat " + src_fname + " >> " + log_fname)
////			if err = cmd.Run(); err != nil { panic(err) }
////			time.Sleep(time.Second)
////		}
////	} ()
//
//
////	reader := r_m.log_reader_map[sheep_id]
//	for {
//		// sync all new log entries with local muster
//		// --> sync_with_logger: syncs one mem_buff at a time with local muster
//		// --> do_log: fills one mem_buff at a time from log file
//		select {
//		case cmd := <- r_m.pasture[sheep_id].logs[log_id].do_log_chan:
//			fmt.Println("********* BAYOPT_LOG SIGNAL ********* 	", sheep_id, log_id, cmd)
//			switch {
//			case cmd == "start":
//			case cmd == "stop":
//			case cmd == "first":
//			case cmd == "last":
//			default:
//				fmt.Println("************ UNKNOWN BAYOPT_LOG COMMAND: ", cmd)
//			}
////			err := r_m.sync_with_logger(sheep_id, log_id, core, reader, do_log)
////			if err == nil { 
////				r_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
////			}
//		}
//	}
//}
//






