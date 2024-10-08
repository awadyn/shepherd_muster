package main

import (
	"fmt"
	"strconv"
	"encoding/csv"
	"time"
	"io"
)

/************************************/
/****** MUSTER SPECIALIZATION  ******/
/************************************/

type nop_muster struct {
	remote_muster
	logs_dir string
}

func (nop_m *nop_muster) init_remote() {
	nop_m.init_log_files(nop_m.logs_dir)
}

/*
   reads one full memory buffer of log entries
*/
func do_nop_log(shared_log *log, reader *csv.Reader) error {
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

func (nop_m *nop_muster) assign_log_files(sheep_id string) {
}


/* The following functions associate each sheep with its
   intlog data in /proc/ixgbe_stats/core/sheep.core
*/
func (nop_m *nop_muster) attach_native_logger(sheep_id string, log_id string) error {
	return nil
}


func (nop_m *nop_muster) ctrl_manage(sheep_id string) {
	fmt.Printf("\033[36m-- MUSTER %v -- SHEEP %v - STARTING CONTROL MANAGER\n\033[0m", nop_m.id, sheep_id)
	sheep := nop_m.pasture[sheep_id]
	var err error
	for {
		select {
		case new_ctrls := <- sheep.new_ctrl_chan:
			for ctrl_id, ctrl_val := range(new_ctrls) {
				err = sheep.controls[ctrl_id].setter(sheep.core, ctrl_val)
				if err != nil { panic(err) }
				sheep.controls[ctrl_id].value = ctrl_val
			}
			nop_m.pasture[sheep.id].ready_ctrl_chan <- control_reply{done: true, ctrls: new_ctrls}
		}
	}
}


func (nop_m *nop_muster) log_manage(sheep_id string) {
	fmt.Printf("\033[34m-- MUSTER %v --SHEEP %v - STARTING LOG MANAGER\n\033[0m", nop_m.id, sheep_id)
	var err error
	for {
		select {
		case req := <- nop_m.pasture[sheep_id].request_log_chan:
			log_id := req[0]
			cmd := req[1]
			switch {
			case cmd == "start":
				// start communication with native logger
				nop_m.assign_log_files(sheep_id)
				err = nop_m.attach_native_logger(sheep_id, log_id)
				if err != nil { 
					nop_m.pasture[sheep_id].logs[log_id].ready_request_chan <- false
				} else {
					nop_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
				}
			case cmd == "stop":
				// stop communication with native logger
				nop_m.pasture[sheep_id].detach_native_logger <- true
				nop_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
			case cmd == "first":
				// get first instance from native logger
				go func() {
					sheep_id := sheep_id
					log_id := log_id
					f := nop_m.pasture[sheep_id].log_f_map[log_id]
					reader := nop_m.pasture[sheep_id].log_reader_map[log_id]
					f.Seek(0, io.SeekStart)
					err := nop_m.sync_with_logger(sheep_id, log_id, reader, do_nop_log, 1)
					if err == io.EOF {
						fmt.Println("************** FILE IS EMPTY *************", log_id) 
						nop_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
						return
					}
					if err != nil { panic(err) }
					nop_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
				} ()
			case cmd == "last":
				// get last instance from native logger
				go func() {
					sheep_id := sheep_id
					log_id := log_id
					f := nop_m.pasture[sheep_id].log_f_map[log_id]
					f.Seek(0, io.SeekStart)

					// get length of log file
					reader1 := csv.NewReader(f)
					reader1.Comma = ' '
					rows, err := reader1.ReadAll() 
					if err != nil { panic(err) }
					len_rows := len(rows)
					if len_rows == 0 {
						fmt.Println("************** FILE IS EMPTY *************", log_id) 
						nop_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
						return
					}

					// read log file except last entry
					f.Seek(0, io.SeekStart)
					reader2 := csv.NewReader(f)
					reader2.Comma = ' '
					counter := 0
					for {
						if counter == len_rows - 1 { break }
						_, err := reader2.Read()
						if err == io.EOF { 
							fmt.Println("************** FILE IS EMPTY *************", log_id) 
							break
						}
						if err != nil { panic(err) }
						counter ++
					}

					err = nop_m.sync_with_logger(sheep_id, log_id, reader2, do_nop_log, 1)
					if err != nil { panic(err) }
					nop_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
				} ()
			default:
				fmt.Println("************ UNKNOWN BAYOPT_LOG COMMAND: ", cmd)
			}
		}
	}
}






