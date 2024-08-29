package main

import (
	"fmt"
	"os"
//	"os/exec"
	"strconv"
	"encoding/csv"
	"time"
	"io"
//	"strings"
//	"bytes"
)

/*********************************************/

func (source_m *flink_source_muster) init() {
	source_m.flink_metrics = []string{"i", "backpressure", "timestamp"}
	source_m.buff_max_size = 1
	var core uint8
	for core = 0; core < source_m.ncores; core ++ {
		c_str := strconv.Itoa(int(core))
		sheep_id := "sheep-" + c_str + "-" + source_m.ip
		log_id := "log-" + c_str + "-" + source_m.ip
		mem_buff := make([][]uint64, source_m.buff_max_size)
		log_c := log{id: log_id,
			     metrics: source_m.flink_metrics,
			     max_size: source_m.buff_max_size,
			     mem_buff: &mem_buff,
			     log_wait_factor: 3,
			     //kill_log_chan: make(chan bool, 1),
			     ready_request_chan: make(chan bool, 1),
			     ready_buff_chan: make(chan bool, 1)}
		source_m.pasture[sheep_id].logs[log_id] = &log_c
	}
	source_m.init_log_files(source_m.logs_dir)
}

func do_flink_log(shared_log *log, reader *csv.Reader) error {
	*shared_log.mem_buff = make([][]uint64, 0)
	var counter uint64 = 0
	//read log entries until buffer is full
	for {
		switch {
		case counter < shared_log.max_size:
			var row []string
			var err error
			for {
				//try until log file non-empty
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

func (source_m *flink_source_muster) assign_log_files(sheep_id string) {
	c_str := strconv.Itoa(int(source_m.pasture[sheep_id].core))
	log_fname := source_m.logs_dir + "/" + c_str
	if source_m.log_f_map[sheep_id] != nil {
		source_m.log_f_map[sheep_id].Close()
	}
	f, err := os.Create(log_fname)
	if err != nil { panic(err) }
	source_m.log_f_map[sheep_id] = f
}

/* 
   The following functions associate each sheep with its
   intlog data in /proc/ixgbe_stats/core/sheep.core
*/
func (source_m *flink_source_muster) attach_native_logger(sheep_id string, log_id string) {
	go func() {
		sheep_id := sheep_id
		for {
			select {
			case <- source_m.pasture[sheep_id].detach_native_logger:
				return
			default:
			}
		}
	} ()
}

func (source_m *flink_source_muster) log_manage(sheep_id string) {
	fmt.Printf("\033[34m-- MUSTER %v -- SHEEP %v - STARTING LOG MANAGER\n\033[0m", source_m.id, sheep_id)
	for {
		select {
		case req := <- source_m.pasture[sheep_id].request_log_chan:
			log_id := req[0]
			cmd := req[1]
			switch {
			case cmd == "start":
				// start communication with native logger
				source_m.assign_log_files(sheep_id)
				source_m.attach_native_logger(sheep_id, log_id)
				source_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
			case cmd == "stop":
				// stop communication with native logger
				source_m.pasture[sheep_id].detach_native_logger <- true
				source_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
			case cmd == "first":
				// get first instance from native logger
				go func() {
					sheep_id := sheep_id
					log_id := log_id
					f := source_m.log_f_map[sheep_id]
					reader := csv.NewReader(f)
					reader.Comma = ' '
					f.Seek(0, io.SeekStart)
					err := source_m.sync_with_logger(sheep_id, log_id, reader, do_flink_log, 1)
					if err == io.EOF {
						fmt.Println("************** FILE IS EMPTY *************", log_id) 
						source_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
						return
					}
					if err != nil { panic(err) }
					source_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
				} ()
			case cmd == "last":
				// get last instance from native logger
				go func() {
					sheep_id := sheep_id
					log_id := log_id
					f := source_m.log_f_map[sheep_id]
					f.Seek(0, io.SeekStart)

					// get length of log file
					reader1 := csv.NewReader(f)
					reader1.Comma = ' '
					rows, err := reader1.ReadAll() 
					if err != nil { panic(err) }
					len_rows := len(rows)
					if len_rows == 0 {
						fmt.Println("************** FILE IS EMPTY *************", log_id) 
						source_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
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

					err = source_m.sync_with_logger(sheep_id, log_id, reader2, do_flink_log, 1)
					if err != nil { panic(err) }
					source_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
				} ()
			default:
				fmt.Println("************ UNKNOWN LOG MANAGEMENT COMMAND: ", cmd)
			}
		}
	}
}


