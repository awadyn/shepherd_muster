package main

// formatting package for printing
import "fmt"


////////////////////////////////

// different musters must each implement the muster interface
//	for example, one muster can log latency metrics and control for processor frequency
//	while another muster can log NIC buffer stats and control for interrupt delay
//	and another muster can log page fault stats and control for memory placement/reservation
//	and so on.
type muster interface {
	log()
	control()
}
// this is a function run_muster that uses an interface muster (which can be any type of muster)
func run_muster(m muster) {
	fmt.Println("RUN_MUSTER: this function should start a muster e.g. in a loop of log-and-control.")
	m.log()
	m.control()
}

type node struct {
	id int
}

// a muster that is responsible for its node's performance and energy 
//	it has a nodeid (an id of the node it is responsible for)
type muster_perf_energy struct {
	node
	//nodeid int;
}

func (n node) get_id() int {
	return n.id
}
func (n *node) set_id(num int) {
	n.id = num
}

// this is a method log implemented by muster_perf_struct
func (m muster_perf_energy) log() {
	fmt.Println("LOG: this function should implement a muster's logging of performance and energy metrics.")
}
// this is a method control implemented by muster_perf_struct
func (m muster_perf_energy) control() {
	fmt.Println("CONTROL: this function should implement a muster's control of performance and energy.")
}


func main() {
	fmt.Println("hello world.. I am the shepherd..")
	m1 := muster_perf_energy {node: node {id: 123}}
	fmt.Println("this is a muster m1 responsible for energy and performance:", m1)
	m1.set_id(12345)
	fmt.Println("changing node id for muster m1 to ", m1.get_id())
	run_muster(m1)
}
