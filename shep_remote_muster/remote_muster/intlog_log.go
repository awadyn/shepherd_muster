package main

import (
	"fmt"
	"strconv"
	"os"
	"os/exec"
	"encoding/csv"
	"io"
	"time"
)

/*********************************************************/

func do_log(shared_log *log, reader *csv.Reader) error {
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
			for i := range(len(intlog_cols)) {
				val, _ := strconv.Atoi(row[i])
				(*shared_log.mem_buff)[counter] = append((*shared_log.mem_buff)[counter], uint64(val))
			}
			counter ++
		case counter == shared_log.max_size:
			return nil
		}
	}
}

func (r_m *muster) sync_with_logger(sheep_id string, log_id string, core uint8, reader *csv.Reader, logger_func func(*log, *csv.Reader)error) error {
	shared_log := r_m.pasture[sheep_id].logs[log_id] 
	for {
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
			return io.EOF
		}
	}
}

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
		// --> do_log: fills one mem_buff at a time from log file
		select {
		case <- r_m.pasture[sheep_id].logs[log_id].kill_log_chan:
			return
		default:
			err = r_m.sync_with_logger(sheep_id, log_id, core, reader, do_log)
			if err == io.EOF { 
				//fmt.Printf("*** EOF  -  CORE :  %v  -  LOG :  %v ***\n", core, log_fname)
				time.Sleep(time.Second)
			}
		}
	}
}







