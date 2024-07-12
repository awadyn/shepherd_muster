package main

import ( "time"
	 "fmt"
//	 "os/exec" 
)

/**************************************/

func (bayopt_s *bayopt_shepherd) run_workload(m_id string) {
//	time.Sleep(time.Second)

	l_m := bayopt_s.local_musters[m_id]
	<- l_m.hb_chan

	fmt.Println("Running Workload...")

	fmt.Println("Requesting start logs..")
	for sheep_id, sheep := range(l_m.pasture) {
		for log_id, _ := range(sheep.logs) {
			l_m.request_log_chan <- []string{sheep_id, log_id, "start"}
//			go func() {
//				sheep := sheep
//				<- sheep.done_request_chan
//			} ()
		}
	}

//	time.Sleep(time.Second * 5)

	for sheep_id, sheep := range(l_m.pasture) {
		for log_id, _ := range(sheep.logs) {
			l_m.request_log_chan <- []string{sheep_id, log_id, "first"}
		}
	}

	time.Sleep(time.Second * 10)

//	cmd := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s " + l_m.ip + " --noload --agent={10.10.1.3,10.10.1.4} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --connections=16 --measure_connections=32 --measure_qps=2000 --qps=100000 --time=10")
//	if err := cmd.Run(); err != nil { panic(err) }
//	fmt.Println("**** DONE WORKLOAD **** ", l_m.ip, " **** ")

	fmt.Println("Requesting end logs..")
	for sheep_id, sheep := range(l_m.pasture) {
		for log_id, _ := range(sheep.logs) {
			l_m.request_log_chan <- []string{sheep_id, log_id, "stop"}
//			go func() {
//				sheep := sheep
//				<- sheep.done_request_chan
//			} ()
		}
	}
	for sheep_id, sheep := range(l_m.pasture) {
		for log_id, _ := range(sheep.logs) {
			l_m.request_log_chan <- []string{sheep_id, log_id, "last"}
		}
	}

}

func main() {
	// assume that a list of nodes is known apriori
	nodes := []node{{ip: "10.10.1.2", ncores: 16, pulse_port: 50051, log_sync_port:50061, ctrl_port: 50071, coordinate_port: 50081}}

	// initialize generic shepherd
	s := shepherd{id: "sheperd-intlog"}
	s.init(nodes)

	// initialize specialized energy-performance shepherd
	bayopt_s := bayopt_shepherd{shepherd:s}
	bayopt_s.init()
	fmt.Println(bayopt_s.joules_diff)

	// for each muster, start pulse + log + control threads for a total
	// of num_musters * [1(pulse client) + 2(log server + coordinator) + 1(ctrl client)]
	// = 4 * num_musters
	bayopt_s.deploy_musters()

	// 1 thread listening for muster pulses
	go bayopt_s.listen_heartbeats()

	// 1 thread managing process signals + (0 <= threads <= muster.ncores) 
	// doing actual log processing 
	go bayopt_s.process_logs()
	//go bayopt_s.compute_control()

	for _, l_m := range(bayopt_s.local_musters) {
		go bayopt_s.run_workload(l_m.id)
	}

	time.Sleep(exp_timeout)
	for _, l_m := range(bayopt_s.local_musters) {
		for sheep_id, _ := range(l_m.pasture) {
			for _, f := range(l_m.out_f_map[sheep_id]) { f.Close() }
		}
	}
}





