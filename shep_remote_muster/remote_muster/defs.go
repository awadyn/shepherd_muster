package main

import (
	"fmt"
	"context"
	"os"
	"encoding/csv"

	"google.golang.org/grpc"
	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
)

/************************************/

var intlog_cols []string = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
				    "instructions", "cycles", "ref_cycles", "llc_miss", 
				    "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
var max_rows uint64 = 1
var max_bytes uint64 = uint64(len(intlog_cols) * 64) * max_rows


type node struct {
	ncores uint8
	ip string
}

type log struct {
	ready_buff_chan chan bool

	mem_buff *[][]uint64
	max_size uint64
	metrics []string

	n_ip string
	id string
}

type control struct {
	dirty bool
	value uint64
	knob string
	n_ip string
	id string
}

type ctrl_req struct {
	sheep_id string
	ctrls map[string]uint64
}

type sheep struct {
	core uint8
	logs map[string]*log
	controls map[string]*control

	ready_ctrl_chan chan bool
	done_ctrl_chan chan bool

	kill_log_chan chan bool
	done_kill_chan chan bool

	id string
}

type muster struct {
	node
	pasture map[string]*sheep

	hb_chan chan bool
	full_buff_chan chan []string
	new_ctrl_chan chan ctrl_req

	done_chan chan []string
	exit_chan chan bool

	id string
}

type remote_muster struct {	// i.e. 1st level specialization of a muster
	muster

	pulse_server_port *int
	ctrl_server_port *int
	log_server_addr *string
	coordinate_server_addr *string

	pb.UnimplementedPulseServer

	logger pb.LogClient
	conn_local *grpc.ClientConn
	ctx_local context.Context
	cancel_local context.CancelFunc 
}

type test_muster struct {	// 2nd level specialization of a muster
	remote_muster

	pb.UnimplementedControlServer

	log_f_map map[string](map[string]*os.File)
	log_reader_map map[string](map[string]*csv.Reader)

	done_log_map map[string](map[string]chan bool)
}

type intlog_muster struct {	// 2nd level specialization of a muster
	remote_muster

	pb.UnimplementedControlServer

	log_f_map map[string]*os.File
	log_reader_map map[string]*csv.Reader
}
/*****************************************/

func (l_ptr *log) show() {
	fmt.Printf("    ADDR %p ", l_ptr)
	fmt.Println("ID:", l_ptr.id, "  --  MAX_SIZE:", l_ptr.max_size, "  --  METRICS:", l_ptr.metrics)
	fmt.Printf("    -- %p R_BUFF:", l_ptr.mem_buff)
	fmt.Println(*l_ptr.mem_buff)

}

func (m *muster) show() {
	fmt.Printf("-- MUSTER %v --\n", m.id)
	fmt.Printf("-- -- NODE -- -- %v\n", m.node)
	fmt.Printf("-- -- PASTURE -- -- %v\n", m.pasture)
}

func (r_m *remote_muster) show() {
	fmt.Printf("-- REMOTE MUSTER :  %v \n", r_m.id)
	fmt.Printf("-- NODE :  %v \n", r_m.node)
	fmt.Printf("   -- PULSE SERVE PORT :  %v \n", *r_m.pulse_server_port)
	fmt.Printf("   -- LOG CLIENT PORT :  %v \n", *r_m.log_server_addr)
	fmt.Printf("   -- CONTROL SERVE PORT :  %v \n", *r_m.ctrl_server_port)
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

