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

type node struct {
	ncores uint8
	ip string
}

type log struct {
	ready_buff_chan chan bool
	kill_log_chan chan bool
	//do_log_chan chan bool
	request_log_chan chan string
	done_log_chan chan bool

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

type Control interface {
	getter()
	setter()
}

type ctrl_req struct {
	sheep_id string
	ctrls map[string]uint64
}

type sheep struct {
	core uint8
	logs map[string]*log
	controls map[string]*control

	request_ctrl_chan chan map[string]uint64 
	ready_ctrl_chan chan bool
	done_ctrl_chan chan bool

	detach_native_logger chan bool
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

	log_f_map map[string]*os.File
	log_reader_map map[string]*csv.Reader
}

type remote_muster struct {	// i.e. 1st level specialization of a muster
	muster

	pulse_server_port *int
	ctrl_server_port *int
	log_server_addr *string
	coordinate_server_port *int

	pb.UnimplementedPulseServer
	pb.UnimplementedControlServer
	pb.UnimplementedCoordinateServer

	logger pb.LogClient
	conn_local *grpc.ClientConn
	ctx_local context.Context
	cancel_local context.CancelFunc 
}

type test_muster struct {	// 2nd level specialization of a muster
	remote_muster

	done_log_map map[string](map[string]chan bool)

	log_f_map map[string](map[string]*os.File)
	log_reader_map map[string](map[string]*csv.Reader)
}

type intlog_muster struct {	// 2nd level specialization of a muster
	remote_muster

	logs_dir string
	intlog_metrics []string
	buff_max_size uint64
}

type bayopt_muster struct {	// 2nd level specialization of a muster
	remote_muster

	intlog_metrics []string
	bayopt_metrics []string
	buff_max_size uint64
}

type flink_muster struct {	// 2nd level specialization of a muster
	remote_muster

	logs_dir string
	flink_metrics []string
	buff_max_size uint64

}

type flink_energy_muster struct {
	flink_muster
}

type flink_backpressure_muster struct {
	flink_muster
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

