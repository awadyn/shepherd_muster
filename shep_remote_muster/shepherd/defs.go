package main

import (
	"fmt"
	"context"

	"google.golang.org/grpc"
	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
)
/************************************/

type node struct {
	ncores uint8
	ip string
}

type log struct {
	core uint8
	ready_chan chan bool
	l_buff *[][]uint64
	max_size uint64
	metrics []string
	n_ip string
	id string
	/* e.g. { "log-ep-i", "10.0.0.1", ["joules", "timestamp"], 64KB, 0xdeadbeef, 0x12345678:PORT(i):10.0.0.1, i}
		0xdeadbeef: l_buff  ->  [  [x, 0]
					   [y, 1]
		   			   [z, 2], ...]  */
}

type control struct {
	core uint8
	value uint64
	knob string
	n_ip string
	id string
	/* e.g. { "ctrl-dvfs-i", "10.0.0.1", "dvfs", 0x1234, i }*/
}

type muster struct {
	node
	//hb_chan chan bool
	hb_chan chan *pb.HeartbeatReply
	process_chan chan string
	logs map[string]*log
	controls map[string]*control
	id string
	/* e.g. {"muster_n", {"ctrl-dvfs-i": {..}, "ctrl-itr-i": {..} ...}, {"log-ep-i": {..}, "log-ep-j": {..}, ...}, node{"10.0.0.1", 24}} */
}

type local_muster struct {
	muster
	log_sync_port *int
	remote_muster_addr *string
	pb.UnimplementedLogServer
}

type cat struct {
	id string
	chaos uint8
}

type shepherd struct {
	musters map[string]*muster
	local_musters map[string]*local_muster

	hb_chan chan *pb.HeartbeatReply
	process_chan chan string

	pulsers map[string]pb.PulseClient
	conn_remotes map[string]*grpc.ClientConn
	ctx_remotes map[string]context.Context
	cancel_remotes map[string]context.CancelFunc

	id string
	/* e.g. {"sheperd-ep", {"muster-10.0.0.1": &muster{..}, "muster-10.0.0.2": &muster{..} ...}} */
}

type ep_shepherd struct {
	shepherd
}

type Shepherd interface {
	init()
	process_logs()
}



func (l_ptr *log) show() {
	fmt.Printf("    ADDR %p ", l_ptr)
	fmt.Println("ID:", l_ptr.id, "  --  MAX_SIZE:", l_ptr.max_size, "  --  METRICS:", l_ptr.metrics)
	fmt.Printf("    -- %p L_BUFF:", l_ptr.l_buff)
	fmt.Println(*l_ptr.l_buff)
	//fmt.Printf("    -- %p R_BUFF:", l_ptr.r_buff)
	//fmt.Println(*l_ptr.r_buff)

}
func (c_ptr *control) show() {
	fmt.Printf("    ADDR %p ", c_ptr)
	fmt.Println("ID:", c_ptr.id, "  --  KNOB:", c_ptr.knob, "  --  VALUE:", c_ptr.value)
}
func (m_ptr *muster) show() {
	fmt.Println()
	fmt.Printf("ADDR %p ", m_ptr)
	fmt.Println("ID:", m_ptr.id, "HB_CHAN:", m_ptr.hb_chan)
	fmt.Println("------ LOGS:", m_ptr.logs)
	fmt.Println("-- CONTROLS:", m_ptr.controls) 
}
func (s_ptr *shepherd) show() {
	fmt.Printf("ADDR %p ", s_ptr)
	fmt.Println("ID:", s_ptr.id, "HB_CHAN:", s_ptr.hb_chan)
	fmt.Println("-- MUSTERS:", s_ptr.musters)
	for _, m := range(s_ptr.musters) {
		m.show()
		for _, l := range(m.logs) {l.show()}
		for _, c := range(m.controls) {c.show()}
	}
	fmt.Println()
}
