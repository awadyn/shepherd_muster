package main

import (
	"fmt"
//	"os"
	"os/exec"
	"strconv"
	"encoding/csv"
	"time"
	"io"
//	"strings"
//	"bytes"
//	"slices"
//	"bufio"
)

/************************************/
/****** MUSTER SPECIALIZATION  ******/
/************************************/

type bayopt_muster struct {
	remote_muster
	logs_dir string
	ixgbe_metrics []string
	// bayopt_metrics []string
	buff_max_size uint64
}

/* 
   read values of bayopt specialization control settings
   and initialize log files accordingly
*/
func (bayopt_m *bayopt_muster) init_remote() {
	iface := bayopt_m.get_internal_iface()
	if iface == "" {
		fmt.Printf("**** PROBLEM: %v cannot get internal network interface name.. aborting\n", bayopt_m.id)
		return
	}
	for _, sheep := range(bayopt_m.pasture) {
		for _, ctrl := range(sheep.controls) {
			switch {
			case ctrl.knob == "itr-delay":
				rx_usecs_reading := ctrl.getter(0, iface)	
				if rx_usecs_reading == 0 {
					fmt.Printf("**** PROBLEM: %v cannot read ethtool rx_usecs value.. assuming default value 1\n", bayopt_m.id)
					ctrl.value = 1
				} else {
					ctrl.value = rx_usecs_reading
				}
			case ctrl.knob == "dvfs":
				ctrl.value = ctrl.getter(sheep.core)
				// TODO error handle
			default:
				fmt.Printf("**** PROBLEM: UNKOWN CTRL KNOB -- %v - %v - %v\n", bayopt_m.id, sheep.id, ctrl.id)
			}
		}
	}
	bayopt_m.init_log_files(bayopt_m.logs_dir)
}

/*
   reads one full memory buffer of log entries
*/
func do_bayopt_log(shared_log *log, reader *csv.Reader) error {
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

func (bayopt_m *bayopt_muster) assign_log_files(sheep_id string) {
//	c_str := strconv.Itoa(int(bayopt_m.pasture[sheep_id].core))
//	log_fname := bayopt_m.logs_dir + "/" + c_str
//	if bayopt_m.log_f_map[sheep_id] != nil {
//		bayopt_m.log_f_map[sheep_id].Close()
//	}
//	f, err := os.Create(log_fname)
//	if err != nil { panic(err) }
//	bayopt_m.log_f_map[sheep_id] = f
}


/* The following functions associate each sheep with its
   intlog data in /proc/ixgbe_stats/core/sheep.core
*/
func (bayopt_m *bayopt_muster) attach_native_logger(sheep_id string, log_id string) {
	c_str := strconv.Itoa(int(bayopt_m.pasture[sheep_id].core))
	src_fname := "/proc/ixgbe_stats/core/" + c_str
	log_fname := bayopt_m.logs_dir + "/" + c_str

	cmd := exec.Command("bash", "-c", "cat " + src_fname)
	if err := cmd.Run(); err != nil { panic(err) }

	go func() {
		sheep_id := sheep_id
		log_id := log_id
		for {
			select {
			case <- bayopt_m.pasture[sheep_id].detach_native_logger:
				return
			default:
				cmd = exec.Command("bash", "-c", "cat " + src_fname + " >> " + log_fname)
				if err := cmd.Run(); err != nil { panic(err) }
				time.Sleep(time.Second * bayopt_m.pasture[sheep_id].logs[log_id].log_wait_factor)
			}
		}
	} ()
}


func (bayopt_m *bayopt_muster) ctrl_manage(sheep_id string) {
	fmt.Printf("\033[36m-- MUSTER %v -- SHEEP %v - STARTING CONTROL MANAGER\n\033[0m", bayopt_m.id, sheep_id)
	sheep := bayopt_m.pasture[sheep_id]
	var err error
	for {
		select {
		case new_ctrls := <- sheep.new_ctrl_chan:
			for ctrl_id, ctrl_val := range(new_ctrls) {
				switch {
				case sheep.controls[ctrl_id].knob == "dvfs":
					err = sheep.controls[ctrl_id].setter(sheep.core, ctrl_val)
				case sheep.controls[ctrl_id].knob == "itr-delay":
					err = sheep.controls[ctrl_id].setter(sheep.core, ctrl_val)
				default:
				}
				if err != nil { panic(err) }
				sheep.controls[ctrl_id].value = ctrl_val
			}
			bayopt_m.pasture[sheep.id].ready_ctrl_chan <- control_reply{done: true, ctrls: new_ctrls}
		}
	}
}


func (bayopt_m *bayopt_muster) log_manage(sheep_id string) {
	fmt.Printf("\033[34m-- MUSTER %v --SHEEP %v - STARTING LOG MANAGER\n\033[0m", bayopt_m.id, sheep_id)
	for {
		select {
		case req := <- bayopt_m.pasture[sheep_id].request_log_chan:
			log_id := req[0]
			cmd := req[1]
			switch {
			case cmd == "start":
				// start communication with native logger
				bayopt_m.assign_log_files(sheep_id)
				bayopt_m.attach_native_logger(sheep_id, log_id)
				bayopt_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
			case cmd == "stop":
				// stop communication with native logger
				bayopt_m.pasture[sheep_id].detach_native_logger <- true
				bayopt_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
			case cmd == "first":
				// get first instance from native logger
				go func() {
					sheep_id := sheep_id
					log_id := log_id
					f := bayopt_m.pasture[sheep_id].log_f_map[log_id]
					reader := bayopt_m.pasture[sheep_id].log_reader_map[log_id]
//					f := bayopt_m.log_f_map[sheep_id]
//					reader := csv.NewReader(f)
//					reader.Comma = ' '
					f.Seek(0, io.SeekStart)
					err := bayopt_m.sync_with_logger(sheep_id, log_id, reader, do_bayopt_log, 1)
					if err == io.EOF {
						fmt.Println("************** FILE IS EMPTY *************", log_id) 
						bayopt_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
						return
					}
					if err != nil { panic(err) }
					bayopt_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
				} ()
			case cmd == "last":
				// get last instance from native logger
				go func() {
					sheep_id := sheep_id
					log_id := log_id
					f := bayopt_m.pasture[sheep_id].log_f_map[log_id]
//					f := bayopt_m.log_f_map[sheep_id]
					f.Seek(0, io.SeekStart)

					// get length of log file
					reader1 := csv.NewReader(f)
					reader1.Comma = ' '
					rows, err := reader1.ReadAll() 
					if err != nil { panic(err) }
					len_rows := len(rows)
					if len_rows == 0 {
						fmt.Println("************** FILE IS EMPTY *************", log_id) 
						bayopt_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
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

					err = bayopt_m.sync_with_logger(sheep_id, log_id, reader2, do_bayopt_log, 1)
					if err != nil { panic(err) }
					bayopt_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
				} ()
			default:
				fmt.Println("************ UNKNOWN BAYOPT_LOG COMMAND: ", cmd)
			}
		}
	}
}






