package main

import (
	"fmt"
	"strconv"
)

/*********************************************/


func (flink_m *flink_muster) init() {
	flink_m.logs_dir = "/users/awadyn/shepherd_muster/shep_remote_muster/flink_logs/"
	flink_m.flink_metrics = []string{"joules","backpressure"}
	flink_m.buff_max_size = 1
	var core uint8
	ctrl_itr_shared := control{n_ip: flink_m.ip, knob: "itr-delay", dirty: false}  
	rx_usecs_shared := ctrl_itr_shared.getter(0, read_rx_usecs)  

	for sheep_id, sheep := range(flink_m.pasture) {
		mem_buff := make([][]uint64, flink_m.buff_max_size)
		c_str := strconv.Itoa(int(sheep.core))
		log_id := "log-" + c_str + "-" + flink_m.ip
		log_c := log{id: log_id,
			     metrics: flink_m.flink_metrics,
			     max_size: flink_m.buff_max_size,
			     mem_buff: &mem_buff}
		log_c.init()

		ctrl_dvfs := control{n_ip: flink_m.ip, knob: "dvfs", dirty: false}
		ctrl_dvfs.id = "ctrl-dvfs-" + c_str + "-" + flink_m.ip
		ctrl_dvfs.value = ctrl_dvfs.getter(core, read_dvfs)
		ctrl_itr := ctrl_itr_shared
		ctrl_itr.id = "ctrl-itr-" + c_str + "-" + flink_m.ip
		ctrl_itr.value = rx_usecs_shared

		flink_m.pasture[sheep_id].logs[log_id] = &log_c
		flink_m.pasture[sheep_id].controls[ctrl_dvfs.id] = &ctrl_dvfs
		flink_m.pasture[sheep_id].controls[ctrl_itr.id] = &ctrl_itr
	}

	flink_m.init_log_files(flink_m.logs_dir)
}

func (flink_m *flink_muster) log_manage(sheep_id string, log_id string) {
	fmt.Printf("-- -- STARTING LOG MANAGER FOR SHEEP %v - LOG %v \n", sheep_id, log_id)
	for {
		select {
		case cmd := <- flink_m.pasture[sheep_id].logs[log_id].request_log_chan:
			switch {
			case cmd == "start":
				// start communication with native logger
//				bayopt_m.assign_log_files(sheep_id)
//				go bayopt_m.attach_native_logger(sheep_id)
				flink_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
			case cmd == "stop":
				// stop communication with native logger
//				bayopt_m.pasture[sheep_id].detach_native_logger <- true
				flink_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
			case cmd == "first":
				// get first instance from native logger
				go func() {
//					sheep_id := sheep_id
//					log_id := log_id
//					f := bayopt_m.log_f_map[sheep_id]
//					reader := csv.NewReader(f)
//					reader.Comma = ' '
//					f.Seek(0, io.SeekStart)
//					err := bayopt_m.sync_with_logger(sheep_id, log_id, reader, do_bayopt_log, 1)
//					if err == io.EOF {
//						fmt.Println("************** FILE IS EMPTY *************", log_id) 
//						bayopt_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
//						return
//					}
//					if err != nil { panic(err) }
					flink_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
				} ()
			case cmd == "last":
				// get last instance from native logger
				go func() {
					flink_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
				} ()
			default:
				fmt.Println("************ UNKNOWN BAYOPT_LOG COMMAND: ", cmd)
			}
		}
	}
}


