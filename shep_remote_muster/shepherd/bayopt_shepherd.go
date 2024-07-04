package main

import (
	"fmt"
	"strconv"
	"math/rand"
)

/**************************************/

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/

var logs_dir string = "/users/awadyn/shepherd_muster/logs/"

var bayopt_cols []string = []string{"joules","latency"}

var buff_max_size uint64 = 1

/* 
   This function initializes a specialized shepherd for energy-and-performance 
   supervision. Each muster under this shepherd's supervision logs energy and
   performance metrics - i.e. joules and timestamp counters - and controls
   energy and performance settings - i.e. interrupt delay 
   and dynamic-voltage-frequency-scaling - for each core under its supervision.
*/
func (bayopt_s bayopt_shepherd) init() {
	for m_id, m := range(bayopt_s.musters) {
		var c uint8
		for c = 0; c < m.ncores; c++ {
			c_str := strconv.Itoa(int(c))
			log_id := "log-" + c_str + "-" + m.ip 
			log_c := log{id: log_id, n_ip: m.ip,
					metrics: bayopt_cols, 
					max_size: buff_max_size, 
					ready_buff_chan: make(chan bool, 1),
					done_process_chan: make(chan bool, 1)}
			mem_buff := make([][]uint64, 0)
			log_c.mem_buff = &mem_buff
			ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + m.ip
			ctrl_dvfs_c := control{id: ctrl_dvfs_id, n_ip: m.ip, knob: "dvfs", value: 0xc00, dirty: false}
			//ctrl_itr_id := "ctrl-itr-" + c_str + "-" + m.ip
			//ctrl_itr_c := control{id: ctrl_itr_id, n_ip: m.ip, knob: "itr-delay", value: 1, dirty: false}

			sheep_id := c_str + "-" + m.ip
			bayopt_s.musters[m_id].pasture[sheep_id].logs[log_id] = &log_c
			bayopt_s.musters[m_id].pasture[sheep_id].controls[ctrl_dvfs_c.id] = &ctrl_dvfs_c
			//bayopt_s.musters[m_id].pasture[sheep_id].controls[ctrl_itr_c.id] = &ctrl_itr_c
			bayopt_s.musters[m_id].pasture[sheep_id].logs[log_c.id].ready_buff_chan <- true
		}
	}

//	bayopt_s.init_log_files()
}

/* This function assigns a map of log files to each sheep/core.
   There can then be a separate sheep log for different controls.
*/
//func (intlog_s *intlog_shepherd) init_log_files() {
//	err := os.Mkdir(logs_dir, 0750)
//	if err != nil && !os.IsExist(err) { panic(err) }
//	for _, l_m := range(intlog_s.local_musters) {
//		l_m.out_f_map = make(map[string](map[string]*os.File))
//		l_m.out_writer_map = make(map[string](map[string]*csv.Writer))
//		l_m.out_f = make(map[string]*os.File)
//		l_m.out_writer = make(map[string]*csv.Writer)
//
//		/* initializing log files */
//		for _, sheep := range(l_m.pasture) {
//			l_m.out_f_map[sheep.id] = make(map[string]*os.File)
//			l_m.out_writer_map[sheep.id] = make(map[string]*csv.Writer)
//			c_str := strconv.Itoa(int(sheep.core))
//			ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + l_m.ip
//			ctrl_itr_id := "ctrl-itr-" + c_str + "-" + l_m.ip
//			ctrl_dvfs := fmt.Sprintf("0x%x", sheep.controls[ctrl_dvfs_id].value)
//			ctrl_itr := strconv.Itoa(int(sheep.controls[ctrl_itr_id].value))
//			out_fname := logs_dir + l_m.id + "_" + c_str + "_" + ctrl_itr + "_" + ctrl_dvfs + ".intlog"
//			f, err := os.Create(out_fname)
//			if err != nil { panic(err) }
//			writer := csv.NewWriter(f)
//			writer.Comma = ' '
//			l_m.out_f_map[sheep.id][out_fname] = f
//			l_m.out_writer_map[sheep.id][out_fname] = writer
//			l_m.out_f[sheep.id] = f
//			l_m.out_writer[sheep.id] = writer
//		}
//	}
//}


/**************************/
/***** LOG PROCESSING *****/
/**************************/

/* 
   This function implements the log processing loop of a Baysian optimization shepherd.
   - bayopt_shepherd expects logs to consist of 99th tail latency + total joules consumed 
     to represent execution for some period of time
*/
func (bayopt_s bayopt_shepherd) process_logs() {
	for {
		select {
		case ids := <- bayopt_s.process_buff_chan:
			m_id := ids[0]
			sheep_id := ids[1]
			log_id := ids[2]
			l_m := bayopt_s.local_musters[m_id]
			sheep := l_m.pasture[sheep_id]
			log := *(sheep.logs[log_id])
			go func() {
				l_m := l_m
				sheep := sheep
				log := log
				fmt.Printf("-------------- PROCESS LOG SIGNAL :  %v - %v - %v\n", m_id, sheep_id, log_id)
				//mem_buff := *(log.mem_buff)
				fmt.Printf("-------------- COMPLETED PROCESS LOG :  %v - %v - %v\n", l_m.id, sheep.id, log.id)	
				// muster can now overwrite mem_buff for this log
				sheep.logs[log.id].done_process_chan <- true
				select {
				case bayopt_s.compute_ctrl_chan <- []string{l_m.id, sheep.id}:
				default:
				}
			} ()
		}
	}
}

/***************************/
/********* CONTROL *********/
/***************************/

func (bayopt_s *bayopt_shepherd) bayopt_ctrl(m_id string, sheep_id string) map[string]uint64 {
	new_ctrls := make(map[string]uint64)
	m := bayopt_s.musters[m_id]
	sheep := m.pasture[sheep_id]
	c_str := strconv.Itoa(int(sheep.core))
	ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + m.ip
	//ctrl_itr_id := "ctrl-itr-" + c_str + "-" + m.ip

	dvfs_list := []uint64{0xc00, 0xe00, 0x1100, 0x1300, 0x1500, 0x1700, 0x1900, 0x1a00}
	dvfs_idx := rand.Intn(len(dvfs_list))
	new_dvfs := dvfs_list[dvfs_idx]
	//itr_list := []uint64{100}
	//itr_idx := rand.Intn(len(itr_list))
	//new_itr := itr_list[itr_idx]

	new_ctrls[ctrl_dvfs_id] = new_dvfs
	//new_ctrls[ctrl_itr_id] = new_itr
	return new_ctrls
}

/* 
  This function implements the control computation loop of a Bayesian optimization shepherd.
*/
func (bayopt_s bayopt_shepherd) compute_control() {
	for {
		select {
		case ids := <- bayopt_s.compute_ctrl_chan:
			m_id := ids[0]
			sheep_id := ids[1]
			l_m := bayopt_s.local_musters[m_id]
			sheep := l_m.pasture[sheep_id]
			go func() {
				l_m := l_m
				sheep := sheep
				fmt.Printf("------------------- COMPUTE CONTROL SIGNAL :  %v - %v\n", l_m.id, sheep.id)
				new_ctrls := bayopt_s.bayopt_ctrl(l_m.id, sheep.id)
				fmt.Printf("----------------------- NEW CONTROLS :  %v\n", new_ctrls)

				l_m.new_ctrl_chan <- control_request{sheep_id: sheep.id, ctrls: new_ctrls}
				ctrl_reply := <- sheep.done_ctrl_chan
				ctrls := ctrl_reply.ctrls
				done_ctrl := ctrl_reply.done
				if done_ctrl { 
					for ctrl_id, ctrl_val := range(ctrls) {
						sheep.controls[ctrl_id].value = ctrl_val
					}
				}
			} ()
		}
	}
}





