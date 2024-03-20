package main

import (
	"fmt"
	"time"
	"strconv"
	"os"
	"encoding/csv"
	"math/rand"
)

/************************************/

/****************************/
/****** INIT & CLEANUP ******/
/****************************/

/* 
   This function initializes a specialized shepherd for energy-and-performance 
   supervision. Each muster under this shepherd's supervision logs energy and
   performance metrics - i.e. joules and timestamp counters - and controls
   energy and performance settings - i.e. interrupt delay 
   and dynamic-voltage-frequency-scaling - for each core under its supervision.
*/
func (ep_s ep_shepherd) init() {
	for m_id, m := range(ep_s.musters) {
		var c uint8
		for c = 0; c < m.ncores; c++ {
			var max_size uint64 = 4096 * 8
			c_str := strconv.Itoa(int(c))
			log_id := "log-" + c_str + "-" + m.ip 
			log_c := log{id: log_id, n_ip: m.ip,
					metrics: []string{"joules", "timestamp"}, 
					max_size: max_size, 
					ready_buff_chan: make(chan bool, 1)}
			l_buff := make([][]uint64, log_c.max_size)
			log_c.l_buff = &l_buff
			ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + m.ip
			ctrl_itr_id := "ctrl-itr-" + c_str + "-" + m.ip
			ctrl_dvfs_c := control{id: ctrl_dvfs_id, n_ip: m.ip, knob: "dvfs", value: 0x1100, dirty: false}
			ctrl_itr_c := control{id: ctrl_itr_id, n_ip: m.ip, knob: "itr-delay", value: 100, dirty: false}

			sheep_id := c_str + "-" + m.ip
			ep_s.musters[m_id].pasture[sheep_id].logs[log_id] = &log_c
			ep_s.musters[m_id].pasture[sheep_id].controls[ctrl_dvfs_c.id] = &ctrl_dvfs_c
			ep_s.musters[m_id].pasture[sheep_id].controls[ctrl_itr_c.id] = &ctrl_itr_c
			ep_s.musters[m_id].pasture[sheep_id].logs[log_c.id].ready_buff_chan <- true
		}
	}
}

func (ep_s ep_shepherd) complete_runs() {
	for {
		select {
		case ids := <- ep_s.complete_run_chan:
			muster_id := ids[0]
			sheep_id := ids[1]
			// TODO confirm no log processing or ctrl computation is still active
			fmt.Printf("*** COMPLETED RUN :  muster %v - sheep %v\n", muster_id, sheep_id)
			ep_s.musters[muster_id].pasture[sheep_id].finish_run_chan <- true
		}
	}
}

func (ep_s *ep_shepherd) init_out_files(out_dir string) {
	ep_s.out_f_map = make(map[string](map[string]*os.File))
	for _, m := range(ep_s.musters) {
		ep_s.out_f_map[m.id] = make(map[string]*os.File)
		for _, sheep := range(m.pasture) {
			c_str := strconv.Itoa(int(sheep.core))
			ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + m.ip
			ctrl_itr_id := "ctrl-itr-" + c_str + "-" + m.ip
			ctrl_dvfs := fmt.Sprintf("0x%x", sheep.controls[ctrl_dvfs_id].value)
			ctrl_itr := strconv.Itoa(int(sheep.controls[ctrl_itr_id].value))
			out_fname := out_dir + "linux.mcd.dmesg.0_" + c_str + "_" + ctrl_itr + "_" + ctrl_dvfs + "_135_200000_" + m.id
			f, err := os.Create(out_fname)
			if err != nil { panic(err) }
			ep_s.out_f_map[m.id][sheep.id] = f
		}
	}
}

//func (ep_s *ep_shepherd) update_out_f_map(m_id string, sheep_id string, out_dir string) {
//	// close current open out_file for this muster/sheep pair
//	ep_s.out_f_map[m_id][sheep_id].Close()
//	sheep := ep_s.musters[m_id].pasture[sheep_id]
//	c_str := strconv.Itoa(int(sheep.core))
//	ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + ep_s.musters[m_id].ip
//	ctrl_itr_id := "ctrl-itr-" + c_str + "-" + ep_s.musters[m_id].ip
//	ctrl_dvfs := fmt.Sprintf("0x%x", sheep.controls[ctrl_dvfs_id].value)
//	ctrl_itr := strconv.Itoa(int(sheep.controls[ctrl_itr_id].value))
//	out_fname := out_dir + "linux.mcd.dmesg.0_" + c_str + "_" + ctrl_itr + "_" + ctrl_dvfs + "_135_200000_" + m_id
//	var f *os.File
//	_, err := os.Stat(out_fname)
//	if err != nil { 
//		if os.IsNotExist(err) {
//			f, err = os.Create(out_fname)
//			if err != nil { panic(err) }
//		}
//	} else {
//		f, err = os.Open(out_fname)
//		if err != nil { panic(err) }
//	}
//	ep_s.out_f_map[m_id][sheep_id] = f
//}

/**************************/
/***** LOG PROCESSING *****/
/**************************/

/* This function implements the log processing loop of an ep-shepherd.
   Currently, an ep-shepherd processing loop re-generates logs read
   by all of its remote musters. This is a test implementation to
   confirm the correctness of the log syncing mechanism.
*/
func (ep_s ep_shepherd) process_full_buffers(m_id string, out_f_map map[string]*os.File) {
	m := ep_s.musters[m_id]
	for {
		select {
		case ids := <- m.process_buff_chan:
			sheep_id := ids[0]
			log_id := ids[1]
			fmt.Printf("-------------- PROCESS LOG SIGNAL :  %v - %v\n", sheep_id, log_id)
			log := *(m.pasture[sheep_id].logs[log_id])
			mem_buff := *(log.l_buff)
			str_mem_buff := make([][]string,0)
			for _, row := range(mem_buff) {
				if len(row) == 0 { break }
				str_row := []string{strconv.Itoa(int(row[0])),
						     strconv.Itoa(int(row[1]))}
				str_mem_buff = append(str_mem_buff, str_row)
			}
			f := out_f_map[sheep_id]
			writer := csv.NewWriter(f)
			writer.Comma = ' '
			writer.WriteAll(str_mem_buff)
			fmt.Printf("-------------- COMPLETED PROCESS LOG :  %v - %v\n", sheep_id, log_id)
			// NOTE: local muster can now start new sync request for this log
			m.pasture[sheep_id].logs[log_id].ready_buff_chan <- true
			m.compute_ctrl_chan <- []string{sheep_id, log_id}
		}
	}
}

/***************************/
/********* CONTROL *********/
/***************************/

func (ep_s *ep_shepherd) run_bayopt(dvfs_val uint64, itr_val uint64 /* , joules uint64, latency float64 */) (uint64, uint64) {
//	dvfs_list := []uint64{0x1100, 0x1300, 0x1500, 0x1700, 0x1900}
	dvfs_list := []uint64{0x1100}
	itr_list := []uint64{100}
	dvfs_idx := rand.Intn(len(dvfs_list))
	itr_idx := rand.Intn(len(itr_list))
	return dvfs_list[dvfs_idx], itr_list[itr_idx]
}

func (ep_s *ep_shepherd) compute_ctrl_per_core(m *muster, sheep_id string) {
	core := strconv.Itoa(int(m.pasture[sheep_id].core))
	ctrl_dvfs_id := "ctrl-dvfs-" + core + "-" + m.ip
	ctrl_itr_id := "ctrl-itr-" + core + "-" + m.ip
	ctrl_dvfs_val := m.pasture[sheep_id].controls[ctrl_dvfs_id].value 
	ctrl_itr_val := m.pasture[sheep_id].controls[ctrl_itr_id].value 
	// TODO compute joules and latency
	new_dvfs, new_itr := ep_s.run_bayopt(ctrl_dvfs_val, ctrl_itr_val /*, joules, latency */)
	m.pasture[sheep_id].controls[ctrl_dvfs_id].value = new_dvfs
	m.pasture[sheep_id].controls[ctrl_dvfs_id].dirty = true
	m.pasture[sheep_id].controls[ctrl_itr_id].value = new_itr
	m.pasture[sheep_id].controls[ctrl_itr_id].dirty = true
}

//func (ep_s *ep_shepherd) compute_ctrl_all_cores(m muster) {
////			n_dirty := 0
////			for _, ctrl := range(m.controls) {
////				if ctrl.dirty { n_dirty ++ }
////			}
////			if n_dirty == len(m.controls) {
////				fmt.Printf("-- -- -- -- NEW CONTROL DECISION -- -- -- --\n", m.controls)
////				m.ready_ctrl_chan <- log_id
////			}
//}

func (ep_s ep_shepherd) compute_control(m_id string, out_f_map map[string]*os.File) {
	m := ep_s.musters[m_id]
	for {
		select {
		case ids := <- m.compute_ctrl_chan:
			sheep_id := ids[0]
			log_id := ids[1]
			fmt.Printf("------------------- COMPUTE CONTROL SIGNAL :  %v - %v\n", sheep_id, log_id)
			ep_s.compute_ctrl_per_core(m, sheep_id)
			fmt.Printf("------------------- NEW CONTROL :  ")
			for _, ctrl := range(m.pasture[sheep_id].controls) { fmt.Printf(" %v ", ctrl) }
			fmt.Println()
			m.ready_ctrl_chan <- sheep_id
//			ep_s.update_out_f_map(m.id, sheep_id, "/home/tanneen/shepherd_muster/shep_reproduced_mcd_logs/")
		}
	}
}

/************************************/

func main() {
	// assume that a list of nodes is known apriori
	nodes := []node{{ip: "10.0.0.1", ncores: 8, pulse_port: 50051, log_sync_port:50061, ctrl_port: 50071},
			{ip: "10.0.0.2", ncores: 8, pulse_port: 50052, log_sync_port:50062, ctrl_port: 50072},
			{ip: "10.0.0.3", ncores: 16, pulse_port: 50053, log_sync_port:50063, ctrl_port: 50073},
			{ip: "10.0.0.4", ncores: 16, pulse_port: 50054, log_sync_port:50064, ctrl_port: 50074}}
	// initialize generic shepherd
	s := shepherd{id: "sheperd-ep"}
	s.init(nodes)
	go s.complete_runs()

	// initialize specialized energy-performance shepherd
	ep_s := ep_shepherd{shepherd:s}
	ep_s.init()
	ep_s.deploy_musters()
	go ep_s.complete_runs()
	out_dir := "/home/tanneen/shepherd_muster/shep_reproduced_mcd_logs/" 
	ep_s.init_out_files(out_dir)

	// start shepherd log processing and control loop
	for m_id, _ := range(ep_s.musters) {
		go ep_s.process_full_buffers(m_id, ep_s.out_f_map[m_id])
		go ep_s.compute_control(m_id, ep_s.out_f_map[m_id])
		for _, f := range(ep_s.out_f_map[m_id]) { defer f.Close() }
	}

	time.Sleep(time.Second*60*2)
}





