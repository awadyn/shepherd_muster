package main

import (
	"fmt"
	"context"

	"google.golang.org/grpc"
	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
)


type node struct {
	ncores uint8
	ip string
}

type log struct {
	core uint8
	ready_buff_chan chan bool
	l_buff *[][]uint64
	r_buff *[][]uint64
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
	pulsing chan bool

	logs map[string]*log
	controls map[string]*control

	hb_chan chan bool
	full_buff_chan chan string

	logger pb.LogClient
	conn_local *grpc.ClientConn
	ctx_local context.Context
	cancel_local context.CancelFunc 

	id string
}

type remote_muster struct {
	muster
	pulse_port *int
	ctrl_port *int
	local_muster_addr *string
	pb.UnimplementedPulseServer
	pb.UnimplementedControlServer
}

/*****************************************/
func (l_ptr *log) show() {
	fmt.Printf("    ADDR %p ", l_ptr)
	fmt.Println("ID:", l_ptr.id, "  --  MAX_SIZE:", l_ptr.max_size, "  --  METRICS:", l_ptr.metrics)
	//fmt.Printf("    -- %p L_BUFF:", l_ptr.l_buff)
	//fmt.Println(*l_ptr.l_buff)
	fmt.Printf("    -- %p R_BUFF:", l_ptr.r_buff)
	fmt.Println(*l_ptr.r_buff)

}
func (m_ptr *muster) show() {
	fmt.Println()
	fmt.Printf("ADDR %p ", m_ptr)
	fmt.Println("ID:", m_ptr.id, "HB_CHAN:", m_ptr.hb_chan)
	fmt.Println("------ LOGS:", m_ptr.logs)
	//fmt.Println("-- CONTROLS:", m_ptr.controls) 
}


