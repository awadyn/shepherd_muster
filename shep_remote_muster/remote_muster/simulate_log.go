package main

//import (
//	"fmt"
//	"strconv"
//	"os"
//	"encoding/csv"
//	"io"
////	"time"
//)
//
///*********************************************************/
///* These functions are written to simulate a logging scenario. Here, logging is equivalent
//   to reading log entries from an existing log file.
//   This code is meant to be re-written to enable actual logging and expose that 
//   logging to the shepherd-muster system.
//   An implementation should:
//   1. setup memory buffers for the shepherd-muster system
//   2. expose memory buffers to external logging process
//   3. execute logging binary, whereby memory buffers are shared such that logging
//      must wait for a continue signal from the shepherd-muster system when
//      a buffer is full
//   The core functionality which enables communication between a logger and the
//   shepherd-muster system would remain unchanged.
//*/
//
////mcd_cols := []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
////		     "instructions", "cycles", "ref_cycles", "llc_miss", 
////		     "c1", "c1e", "c3", "c6", "c7", "joules","timestamp"}
//
//func do_log(shared_log *log, reader *csv.Reader) error {
//	var counter uint64 = 0
//	for {
//		switch {
//		case counter < shared_log.max_size:
//			row, err := reader.Read()
//			if err == io.EOF { 
//				return err
//			}
//			if err != nil { panic(err) }
//			joules_val, _ := strconv.Atoi(row[0])
//			timestamp_val, _ := strconv.Atoi(row[1])
//			(*shared_log.mem_buff)[counter] = []uint64{uint64(joules_val), uint64(timestamp_val)}
//			counter ++
//		case counter == shared_log.max_size:
//			return nil
//		}
//	}
//}
//
////func (r_m *muster) sync_with_logger(sheep_id string, log_id string, core uint8, reader *csv.Reader, logger_func func(*log, T)error) error {
//func (r_m *muster) sync_with_logger(sheep_id string, log_id string, core uint8, reader *csv.Reader, logger_func func(*log, *csv.Reader)error) error {
//	shared_log := r_m.pasture[sheep_id].logs[log_id] 
//	for {
//		select {
//		case <- r_m.pasture[sheep_id].kill_log_chan:
//			fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! KILLING LOGGER :  ", sheep_id)
//			return nil
//		default:
//		}
//		err := logger_func(shared_log, reader)
//		switch {
//		case err == nil:	// => logged one buff
//			r_m.full_buff_chan <- []string{sheep_id, log_id}
//			<- shared_log.ready_buff_chan
//			*shared_log.mem_buff = make([][]uint64, shared_log.max_size)
//		case err == io.EOF:	// => logging done
//			r_m.full_buff_chan <- []string{sheep_id, log_id}
//			<- shared_log.ready_buff_chan
//			return io.EOF
//		}
//	}
//}
//
//func (r_m *test_muster) simulate_remote_log(sheep_id string, log_id string, core uint8) {
//	<- r_m.hb_chan
//
//	c_str := strconv.Itoa(int(core))
//	ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + r_m.ip
//	ctrl_itr_id := "ctrl-itr-" + c_str + "-" + r_m.ip
//	dvfs_val := r_m.pasture[sheep_id].controls[ctrl_dvfs_id].value
//	itr_val := r_m.pasture[sheep_id].controls[ctrl_itr_id].value
//	dvfs_str := fmt.Sprintf("0x%x", dvfs_val)
//	itr_str := strconv.Itoa(int(itr_val))
//	log_fname := "/home/tanneen/shepherd_muster/mcd_logs/linux.mcd.dmesg.0_" + c_str + "_" + itr_str + "_" + dvfs_str + "_135_200000.ep.csv.ep"
//
//	fmt.Printf("-- -- CORE %v -- -- SIMULATING REMOTE LOG :  %v \n", c_str, log_fname)
//	log_f_map := r_m.log_f_map[sheep_id]
//	log_reader_map := r_m.log_reader_map[sheep_id]
//	var f *os.File
//	var reader *csv.Reader
//	var err error
//	if log_f_map[log_fname] != nil { // log file has been previously open
//		f = log_f_map[log_fname]
//		reader = log_reader_map[log_fname]
//	} else {
//		f, err = os.Open(log_fname)
//		if err != nil { panic(err) }
//		reader = csv.NewReader(f)
//		reader.Comma = ' '
//		log_f_map[log_fname] = f
//		log_reader_map[log_fname] = reader
//		r_m.done_log_map[sheep_id][log_fname] = make(chan bool, 1)
//	}
//
//	for {
//		err := r_m.sync_with_logger(sheep_id, log_id, core, reader, do_log)
////		err := r_m.log_sync(sheep_id, log_id, core, reader)
//		if err == io.EOF { 
//			select {
//			case r_m.done_log_map[sheep_id][log_fname] <- true:
//			default:
//			}
//			fmt.Printf("***** DONE SIMULATING LOGGER  -  CORE :  %v  -  LOG :  %v *****\n", core, log_fname)
//			return
//		}
//		// otherwise, log was killed after new ctrl
//		r_m.pasture[sheep_id].done_kill_chan <- true
//		return
//	}
//}







