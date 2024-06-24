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
	var counter uint64 = 0
	for {
		switch {
		case counter < shared_log.max_size:
			row, err := reader.Read()
			if err == io.EOF { 
				return err
			}
			if err != nil { panic(err) }
			(*shared_log.mem_buff)[counter] = []uint64{}
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
		case err == io.EOF:	// => logging done
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

	delta_log_chan := make(chan bool, 1)

	go func() {
		fi, err := f.Stat()
		if err != nil { panic(err) }
		log_size := fi.Size()
		var new_log_size int64
		var delta_size int64 = log_size	
		for {
			cmd := exec.Command("bash", "-c", "cat " + src_fname + " >> " + log_fname)
			if err := cmd.Run(); err != nil { panic(err) }
			fi, err := f.Stat()
			if err != nil { panic(err) }
			new_log_size = fi.Size()
			delta_size = new_log_size - log_size
			if delta_size > 0 {
				fmt.Println("old size: ", log_size, " new size: ", new_log_size, " delta: ", delta_size)
				select {
				case delta_log_chan <- true:
				default:
				}
				log_size = new_log_size
			}

			time.Sleep(time.Second * 3)

		}
	} ()

	for {
		// sync all new log entries with local muster
		// --> sync_with_logger: syncs one mem_buff at a time with local muster
		// --> do_log: fills one mem_buff at a time from log file
		select {
		case <- delta_log_chan:
			err = r_m.sync_with_logger(sheep_id, log_id, core, reader, do_log)
			if err == io.EOF { 
				//fmt.Printf("***** READER AT EOF  -  CORE :  %v  -  LOG :  %v *****\n", core, log_fname)
				time.Sleep(time.Second)
			}
		}
	}
}







