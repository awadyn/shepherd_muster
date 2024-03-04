package main

import (
	"fmt"
	"time"
	"strconv"
	"os"
	"encoding/csv"
)

/************************************/

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
			c_str := strconv.Itoa(int(c))
			var max_size uint64 = 1024
			log_id := "log-" + c_str + "-" + m.ip 
			log_c := log{id: log_id, n_ip: m.ip, core: uint8(c),
					metrics: []string{"joules", "timestamp"}, 
					max_size: max_size, 
					ready_chan: make(chan bool, 1)}
			l_buff := make([][]uint64, log_c.max_size)
			log_c.l_buff = &l_buff
			ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + m.ip
			ctrl_itr_id := "ctrl-itr-" + c_str + "-" + m.ip
			ctrl_dvfs_c := control{id: ctrl_dvfs_id, n_ip: m.ip, core: c, knob: "dvfs", value: 0xffff}
			ctrl_itr_c := control{id: ctrl_itr_id, n_ip: m.ip, core: c, knob: "itr-delay", value: 1}

			ep_s.musters[m_id].logs[log_c.id] = &log_c
			ep_s.musters[m_id].controls[ctrl_dvfs_c.id] = &ctrl_dvfs_c
			ep_s.musters[m_id].controls[ctrl_itr_c.id] = &ctrl_itr_c

			ep_s.musters[m_id].logs[log_c.id].ready_chan <- true
		}
	}
}

func (ep_s *ep_shepherd) init_out_files(out_dir string) map[string](map[string]*os.File) {
	out_f_map := make(map[string](map[string]*os.File))
	for _, m := range(ep_s.musters) {
		out_f_map[m.id] = make(map[string]*os.File)
		for _, log := range(m.logs) {
			out_fname := out_dir + "linux.mcd.dmesg.0_" + strconv.Itoa(int(log.core)) + "_100_0x1100_135_200000_" + m.id
			f, err := os.Create(out_fname)
			if err != nil { panic(err) }
			out_f_map[m.id][log.id] = f
		}
	}
	return out_f_map
}

/* This function implements the log processing loop of an ep-shepherd.
   Currently, an ep-shepherd processing loop re-generates logs read
   by all of its remote musters. This is a test implementation to
   confirm the correctness of the log syncing mechanism.
*/
func (ep_s ep_shepherd) process_logs() {
	// output log files that ep_s will populate with remote-log data
	out_dir := "/home/tanneen/shepherd_muster/shep_reproduced_mcd_logs/" 
	out_f_map := ep_s.init_out_files(out_dir)
	for m_id, _ := range(out_f_map) {
		for _, f := range(out_f_map[m_id]) { defer f.Close() }
	}

	for {
		for _, m := range(ep_s.musters) {
			select {
			case log_id := <- m.process_chan:
				fmt.Println("-- -- -- RECEIVED PROCESS LOG SIGNAL FOR ", log_id)
				log := *(m.logs[log_id])
				mem_buff := *(log.l_buff)
				str_mem_buff := make([][]string,0)
				for _, row := range(mem_buff) {
					if len(row) == 0 { break }
					str_row := []string{strconv.Itoa(int(row[0])),
							     strconv.Itoa(int(row[1]))}
					str_mem_buff = append(str_mem_buff, str_row)
				}
				f := out_f_map[m.id][log_id]
				writer := csv.NewWriter(f)
				writer.Comma = ' '
				writer.WriteAll(str_mem_buff)

				fmt.Println("-- -- -- COMPLETED PROCESSING LOG ", log_id)
				m.logs[log_id].ready_chan <- true
				// NOTE: local muster can now start new sync request for this log
			default:
			}
		}
	}
}

/************************************/

func main() {
	// assume that a list of nodes is known apriori
	nodes := []node{{ip: "10.0.0.1", ncores: 8},
			{ip: "10.0.0.2", ncores: 8},
			{ip: "10.0.0.3", ncores: 16},
			{ip: "10.0.0.4", ncores: 16}}
	// initialize generic shepherd
	s := shepherd{id: "sheperd-ep"}
	s.init(nodes)
	// initialize specialized energy-performance shepherd
	ep_s := ep_shepherd{s}
	ep_s.init()
	ep_s.show()
	// start local musters and heartbeats
	ep_s.deploy_musters()
	// start shepherd log processing loop
//	go ep_s.process_logs()

	time.Sleep(time.Second*60)
}
