package main

import (
	"fmt"
	"time"
	"strconv"
	"os"
	"encoding/csv"
//	"math/rand"
)

/**************************************/

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/

var logs_dir string = "/users/awadyn/shepherd_muster/logs/"

var intlog_cols []string = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes", "instructions", "cycles", "ref_cycles", "llc_miss", "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}

var buff_max_size uint64 = 4096 * 4

/* 
   This function initializes a specialized shepherd for energy-and-performance 
   supervision. Each muster under this shepherd's supervision logs energy and
   performance metrics - i.e. joules and timestamp counters - and controls
   energy and performance settings - i.e. interrupt delay 
   and dynamic-voltage-frequency-scaling - for each core under its supervision.
*/
func (intlog_s intlog_shepherd) init() {
	for m_id, m := range(intlog_s.musters) {
		var c uint8
		for c = 0; c < m.ncores; c++ {
			c_str := strconv.Itoa(int(c))
			log_id := "log-" + c_str + "-" + m.ip 
			log_c := log{id: log_id, n_ip: m.ip,
					metrics: intlog_cols, 
					max_size: buff_max_size, 
					ready_buff_chan: make(chan bool, 1),
					done_process_chan: make(chan bool, 1)}
			mem_buff := make([][]uint64, log_c.max_size)
			log_c.mem_buff = &mem_buff
			ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + m.ip
			ctrl_itr_id := "ctrl-itr-" + c_str + "-" + m.ip
			ctrl_dvfs_c := control{id: ctrl_dvfs_id, n_ip: m.ip, knob: "dvfs", value: 0xc00, dirty: false}
			ctrl_itr_c := control{id: ctrl_itr_id, n_ip: m.ip, knob: "itr-delay", value: 1, dirty: false}

			sheep_id := c_str + "-" + m.ip
			intlog_s.musters[m_id].pasture[sheep_id].logs[log_id] = &log_c
			intlog_s.musters[m_id].pasture[sheep_id].controls[ctrl_dvfs_c.id] = &ctrl_dvfs_c
			intlog_s.musters[m_id].pasture[sheep_id].controls[ctrl_itr_c.id] = &ctrl_itr_c
			intlog_s.musters[m_id].pasture[sheep_id].logs[log_c.id].ready_buff_chan <- true
		}
	}

	intlog_s.init_log_files()
}

/* This function assigns a map of log files to each sheep/core.
   There can then be a separate sheep log for different controls.
*/
func (intlog_s *intlog_shepherd) init_log_files() {
	err := os.Mkdir(logs_dir, 0750)
	if err != nil && !os.IsExist(err) { panic(err) }
	for _, l_m := range(intlog_s.local_musters) {
		l_m.out_f_map = make(map[string](map[string]*os.File))
		l_m.out_writer_map = make(map[string](map[string]*csv.Writer))
		l_m.out_f = make(map[string]*os.File)
		l_m.out_writer = make(map[string]*csv.Writer)

		/* initializing log files */
		for _, sheep := range(l_m.pasture) {
			l_m.out_f_map[sheep.id] = make(map[string]*os.File)
			l_m.out_writer_map[sheep.id] = make(map[string]*csv.Writer)
			c_str := strconv.Itoa(int(sheep.core))
			ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + l_m.ip
			ctrl_itr_id := "ctrl-itr-" + c_str + "-" + l_m.ip
			ctrl_dvfs := fmt.Sprintf("0x%x", sheep.controls[ctrl_dvfs_id].value)
			ctrl_itr := strconv.Itoa(int(sheep.controls[ctrl_itr_id].value))
			out_fname := logs_dir + l_m.id + "_" + c_str + "_" + ctrl_itr + "_" + ctrl_dvfs + ".intlog"
			f, err := os.Create(out_fname)
			if err != nil { panic(err) }
			writer := csv.NewWriter(f)
			writer.Comma = ' '
			l_m.out_f_map[sheep.id][out_fname] = f
			l_m.out_writer_map[sheep.id][out_fname] = writer
			l_m.out_f[sheep.id] = f
			l_m.out_writer[sheep.id] = writer
		}
	}
}

//func (ep_s *ep_shepherd) switch_out_f(m_id string, sheep_id string, out_dir string) {
//	l_m := ep_s.local_musters[m_id]
//	sheep := l_m.pasture[sheep_id]
//	c_str := strconv.Itoa(int(sheep.core))
//	ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + l_m.ip
//	ctrl_itr_id := "ctrl-itr-" + c_str + "-" + l_m.ip
//	ctrl_dvfs := fmt.Sprintf("0x%x", sheep.controls[ctrl_dvfs_id].value)
//	ctrl_itr := strconv.Itoa(int(sheep.controls[ctrl_itr_id].value))
//	out_fname := out_dir + "linux.mcd.dmesg.0_" + c_str + "_" + ctrl_itr + "_" + ctrl_dvfs + "_135_200000_" + m_id
//	l_m.out_f[sheep_id] = l_m.out_f_map[sheep_id][out_fname]
//	l_m.out_writer[sheep_id] = l_m.out_writer_map[sheep_id][out_fname]
//}

//func (ep_s *ep_shepherd) update_out_f_map(m_id string, sheep_id string, out_dir string) {
//	l_m := ep_s.local_musters[m_id]
//	sheep := l_m.pasture[sheep_id]
//	// open new out_file
//	c_str := strconv.Itoa(int(sheep.core))
//	ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + ep_s.musters[m_id].ip
//	ctrl_itr_id := "ctrl-itr-" + c_str + "-" + ep_s.musters[m_id].ip
//	ctrl_dvfs := fmt.Sprintf("0x%x", sheep.controls[ctrl_dvfs_id].value)
//	ctrl_itr := strconv.Itoa(int(sheep.controls[ctrl_itr_id].value))
//	out_fname := out_dir + "linux.mcd.dmesg.0_" + c_str + "_" + ctrl_itr + "_" + ctrl_dvfs + "_135_200000_" + m_id
//	_, err := os.Stat(out_fname)
//	if err != nil { 
//		if os.IsNotExist(err) {
//			f, err := os.Create(out_fname)
//			if err != nil { panic(err) }
//			l_m.out_f_map[sheep_id][out_fname] = f
//			writer := csv.NewWriter(f)
//			writer.Comma = ' '
//			l_m.out_writer_map[sheep.id][out_fname] = writer
//		}
//	} 
//}

//func (ep_s *ep_shepherd) apply_control(m_id string, sheep_id string, ctrls map[string]uint64) {
////	sheep := ep_s.musters[m_id].pasture[sheep_id]
////	done_ctrl := <- sheep.done_ctrl_chan
////	if !done_ctrl { return }
////	for ctrl_id, ctrl_val := range(ctrls) {
////		sheep.controls[ctrl_id].value = ctrl_val
////	}
//	ep_s.update_out_f_map(m_id, sheep_id, "/home/tanneen/shepherd_muster/shep_reproduced_mcd_logs/")
//	ep_s.switch_out_f(m_id, sheep_id, "/home/tanneen/shepherd_muster/shep_reproduced_mcd_logs/")
//	fmt.Println("!!!!!!!!!!!!!!!!!!! SWITCHED OUTFILES")
//}

//func (ep_s ep_shepherd) complete_runs() {
//	for {
//		select {
//		case ids := <- ep_s.complete_run_chan:
//			muster_id := ids[0]
//			sheep_id := ids[1]
//			// TODO confirm no log processing or ctrl computation is still active
//			fmt.Printf("*** COMPLETED RUN :  muster %v - sheep %v\n", muster_id, sheep_id)
//			ep_s.musters[muster_id].pasture[sheep_id].finish_run_chan <- true
//		}
//	}
//}

///**************************/
///***** LOG PROCESSING *****/
///**************************/
//
///* This function implements the log processing loop of an ep-shepherd.
//   Currently, an ep-shepherd processing loop re-generates logs generated
//   by all of its remote musters. This is a test implementation to
//   confirm the correctness of the general functionality.
//*/
func (intlog_s intlog_shepherd) process_logs() {
	for {
		select {
		case ids := <- intlog_s.process_buff_chan:
			m_id := ids[0]
			sheep_id := ids[1]
			log_id := ids[2]
			l_m := intlog_s.local_musters[m_id]
			sheep := l_m.pasture[sheep_id]
			log := *(sheep.logs[log_id])
			go func() {
				l_m := l_m
				sheep := sheep
				log := log
				fmt.Printf("-------------- PROCESS LOG SIGNAL :  %v - %v - %v\n", m_id, sheep_id, log_id)
				mem_buff := *(log.mem_buff)
				str_mem_buff := make([][]string,0)
				for _, row := range(mem_buff) {
					if len(row) == 0 { break }
					str_row := []string{}
					for i := range(len(intlog_cols)) {
						val := strconv.Itoa(int(row[i]))
						str_row = append(str_row, val)
					}
					str_mem_buff = append(str_mem_buff, str_row)
				}
				writer := l_m.out_writer[sheep.id]
				writer.WriteAll(str_mem_buff)
				fmt.Printf("-------------- COMPLETED PROCESS LOG :  %v - %v - %v\n", l_m.id, sheep.id, log.id)	
				// muster can now overwrite mem_buff for this log
				sheep.logs[log.id].done_process_chan <- true
//				select {
//				case ep_s.compute_ctrl_chan <- []string{l_m.id, sheep.id}:
//				default:
//				}
			} ()
		}
	}
}
//
///***************************/
///********* CONTROL *********/
///***************************/
//
//func (ep_s *ep_shepherd) rand_ctrl(m_id string, sheep_id string) map[string]uint64 {
//	new_ctrls := make(map[string]uint64)
//	m := ep_s.musters[m_id]
//	sheep := m.pasture[sheep_id]
//	c_str := strconv.Itoa(int(sheep.core))
//	ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + m.ip
//	ctrl_itr_id := "ctrl-itr-" + c_str + "-" + m.ip
////	dvfs_list := []uint64{0x1300, 0x1100}
////	itr_list := []uint64{100}
////	dvfs_idx := rand.Intn(len(dvfs_list))
////	itr_idx := rand.Intn(len(itr_list))
////	new_dvfs := dvfs_list[dvfs_idx]
////	new_itr := itr_list[itr_idx]
//	var new_dvfs uint64
//	var new_itr uint64
//	if sheep.controls[ctrl_dvfs_id].value == 0x1100 { new_dvfs = 0x1300  
//	} else { new_dvfs = 0x1100 }
////	new_dvfs = 0x1300
//	new_itr = 100
//	new_ctrls[ctrl_dvfs_id] = new_dvfs
//	new_ctrls[ctrl_itr_id] = new_itr
//	return new_ctrls
//}
//
///* This function implements the control computation loop of an ep-shepherd.
//*/
//func (ep_s ep_shepherd) compute_control() {
//	for {
//		select {
//		case ids := <- ep_s.compute_ctrl_chan:
//			m_id := ids[0]
//			sheep_id := ids[1]
//			l_m := ep_s.local_musters[m_id]
//			sheep := l_m.pasture[sheep_id]
//			go func() {
//				l_m := l_m
//				sheep := sheep
//				fmt.Printf("------------------- COMPUTE CONTROL SIGNAL :  %v - %v\n", l_m.id, sheep.id)
//				new_ctrls := ep_s.rand_ctrl(l_m.id, sheep.id)
//				fmt.Printf("------------------- NEW CONTROLS :  %v\n", new_ctrls)
//
//				l_m.new_ctrl_chan <- control_request{sheep_id: sheep.id, ctrls: new_ctrls}
//
//				ctrl_reply := <- sheep.done_ctrl_chan
//				done_ctrl := ctrl_reply.done
//				ctrls := ctrl_reply.ctrls
//				if done_ctrl {
//					for ctrl_id, ctrl_val := range(ctrls) {
//						sheep.controls[ctrl_id].value = ctrl_val
//					}
//					ep_s.apply_control(l_m.id, sheep.id, ctrls)
//				}
//			} ()
//		}
//	}
//}

/************************************/

func main() {
	// assume that a list of nodes is known apriori
	nodes := []node{{ip: "128.110.96.54", ncores: 16, pulse_port: 50051, log_sync_port:50061, ctrl_port: 50071}}

	// initialize generic shepherd
	s := shepherd{id: "sheperd-intlog"}
	s.init(nodes)
	s.init_local(nodes)

	// initialize specialized energy-performance shepherd
	intlog_s := intlog_shepherd{shepherd:s}
	intlog_s.init()

	// for each muster, start pulse + log + control threads
	intlog_s.deploy_musters()

	go intlog_s.listen_heartbeats()
	go intlog_s.process_logs()
//	go intlog_s.compute_control()

	time.Sleep(time.Second*60)
	for _, l_m := range(intlog_s.local_musters) {
		for sheep_id, _ := range(l_m.pasture) {
			for _, f := range(l_m.out_f_map[sheep_id]) { f.Close() }
		}
	}
}





