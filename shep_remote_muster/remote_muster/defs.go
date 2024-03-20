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

type remote_logger struct {
	stop_log_chan chan bool		// signal stop current logging ==> stop appending to r_buff
	done_log_chan chan bool		// signal logging is complete
}

type log struct {
	remote_logger

	ready_buff_chan chan bool
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
	dirty bool
	value uint64
	knob string
	n_ip string
	id string
	/* e.g. { "ctrl-dvfs-i", "10.0.0.1", "dvfs", 0x1234, i }*/
}

type sheep struct {
	core uint8
	logs map[string]*log
	controls map[string]*control
	id string
}

type muster struct {
	node
	logger pb.LogClient
	conn_local *grpc.ClientConn
	ctx_local context.Context
	cancel_local context.CancelFunc 
	hb_chan chan bool
	full_buff_chan chan []string
	pasture map[string]*sheep
	id string
}

type remote_muster struct {
	muster
	pulse_port *int
	ctrl_port *int
	local_muster_addr *string
	coordinate_addr *string
	pb.UnimplementedPulseServer
	pb.UnimplementedControlServer
}

/*****************************************/

func (l_ptr *log) show() {
	fmt.Printf("    ADDR %p ", l_ptr)
	fmt.Println("ID:", l_ptr.id, "  --  MAX_SIZE:", l_ptr.max_size, "  --  METRICS:", l_ptr.metrics)
	fmt.Printf("    -- %p R_BUFF:", l_ptr.r_buff)
	fmt.Println(*l_ptr.r_buff)

}

func (m *muster) show() {
	fmt.Printf("-- MUSTER %v --\n", m.id)
	fmt.Printf("-- -- NODE -- -- %v\n", m.node)
	fmt.Printf("-- -- PASTURE -- -- %v\n", m.pasture)
}

func (r_m *remote_muster) show() {
	fmt.Printf("-- REMOTE MUSTER :  %v \n", r_m.id)
	fmt.Printf("-- NODE :  %v \n", r_m.node)
	fmt.Printf("   -- PULSE SERVE PORT :  %v \n", *r_m.pulse_port)
	fmt.Printf("   -- LOG CLIENT PORT :  %v \n", *r_m.local_muster_addr)
	fmt.Printf("   -- CONTROL SERVE PORT :  %v \n", *r_m.ctrl_port)
	fmt.Printf("   -- PASTURE :  \n")
	for sheep_id, _ := range(r_m.pasture) {
		fmt.Printf("      -- SHEEP %v \n", sheep_id)
		fmt.Printf("         -- LOGS :  \n")
		for log_id, _ := range(r_m.pasture[sheep_id].logs) {
			fmt.Printf("            %v \n", r_m.pasture[sheep_id].logs[log_id])
		} 
		fmt.Printf("         -- CONTROLS :  \n")
		for ctrl_id, _ := range(r_m.pasture[sheep_id].controls) {
			fmt.Printf("            %v \n", r_m.pasture[sheep_id].controls[ctrl_id])
		}
	}
	fmt.Println()
}

