package main

import (
	"fmt"
	"context"
	"os"
	"encoding/csv"
	"time"

	"google.golang.org/grpc"
	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
)
/************************************/

var exp_timeout time.Duration = time.Second * 75

type node struct {
	ncores uint8
	pulse_port int
	log_port int
	ctrl_port int
	coordinate_port int
	ip_idx int			//differentiates musters on the same node
	ip string
}


type control struct {
	ready_request_chan chan bool	//syncs access to ctrl object 

	value uint64
	dirty bool
	knob string
	n_ip string
	id string

	getter func(uint8, ...string)uint64
	setter func(uint8, uint64)error
}

type control_request struct {
	sheep_id string
	ctrls map[string]uint64
}

type control_reply struct {
	done bool
	ctrls map[string]uint64
}


type log struct {
	ready_request_chan chan bool	//syncs access to log object 
	ready_buff_chan chan bool	//syncs access to log memory buffer
	ready_process_chan chan bool	//..? 

	mem_buff *[][]uint64
	max_size uint64
	metrics []string
	n_ip string
	id string

	log_wait_factor time.Duration		// seconds to wait before updating log filesystem representative
						// TODO: buff_wait_factor ??
}


type sheep struct {
	//finish_run_chan chan bool

	new_ctrl_chan chan map[string]uint64	//signals set new ctrls 
	ready_ctrl_chan chan control_reply	//syncs application of ctrl change
	request_log_chan chan []string		//signals get current logs
	request_ctrl_chan chan string		//signals get current ctrls
	detach_native_logger chan bool

	core uint8
	logs map[string]*log
	controls map[string]*control
	id string
}

type muster struct {
	full_buff_chan chan []string
	new_ctrl_chan chan control_request

	request_log_chan chan []string
	request_ctrl_chan chan []string

	node
	pulsing bool
	role string
	pasture map[string]*sheep
	id string

	log_f_map map[string]*os.File
	log_reader_map map[string]*csv.Reader
}

type local_muster struct {
	muster
	hb_chan chan *pb.HeartbeatReply

	log_server_port *int
	ctrl_server_addr *string
	pulse_server_addr *string
	coordinate_server_addr *string

	out_f_map map[string](map[string]*os.File)
	out_writer_map map[string](map[string]*csv.Writer)
	out_f map[string]*os.File
	out_writer map[string]*csv.Writer

	pb.UnimplementedLogServer
}

type remote_muster struct {	// i.e. 1st level specialization of a muster
	muster
	hb_chan chan bool

	log_server_addr *string
	ctrl_server_port *int
	pulse_server_port *int
	coordinate_server_port *int

	pb.UnimplementedPulseServer
	pb.UnimplementedControlServer
	pb.UnimplementedCoordinateServer

	logger pb.LogClient
	conn_local *grpc.ClientConn
	ctx_local context.Context
	cancel_local context.CancelFunc 
}

type cat struct {
	id string
	chaos uint8
}

type shepherd struct {
	hb_chan chan *pb.HeartbeatReply
	process_buff_chan chan []string
	compute_ctrl_chan chan []string

	//complete_run_chan chan []string

	musters map[string]*muster
	local_musters map[string]*local_muster
	id string
}

type Shepherd interface {
	init()
	deploy_musters()
	process_logs()
	compute_control()
	complete_run()
}

/* SPECIALIZATIONS */

type intlog_shepherd struct {
	shepherd
	logs_dir string
	intlog_metrics []string
	buff_max_size uint64
}

type intlog_muster struct {
	remote_muster
	logs_dir string
	intlog_metrics []string
	buff_max_size uint64
}


type flink_shepherd struct {
	shepherd
	flink_musters map[string]*flink_local_muster
}

type flink_local_muster struct {
	base_muster *local_muster
	logs_dir string
	flink_metrics []string
	buff_max_size uint64
}

type flink_worker_muster struct {
	bayopt_muster
}

type flink_source_muster struct {
	remote_muster
	logs_dir string
	flink_metrics []string
	buff_max_size uint64
}

type ep_shepherd struct {
	shepherd
}

type test_muster struct {
	remote_muster
	done_log_map map[string](map[string]chan bool)
	log_f_map map[string](map[string]*os.File)
	log_reader_map map[string](map[string]*csv.Reader)
}


/*****************************************/

func (l_ptr *log) show() {
	fmt.Printf("    ADDR %p ", l_ptr)
	fmt.Println("ID:", l_ptr.id, "  --  MAX_SIZE:", l_ptr.max_size, "  --  METRICS:", l_ptr.metrics)
	fmt.Printf("    -- %p R_BUFF:", l_ptr.mem_buff)
	fmt.Println(*l_ptr.mem_buff)

}

func (r_m *remote_muster) show() {
	fmt.Printf("-- REMOTE MUSTER :  %v\n   %v \n", r_m.id, r_m)
	fmt.Printf("-- ROLE : %v \n", r_m.role)
	fmt.Printf("-- NODE :  %v \n", r_m.node)
	fmt.Printf("   -- PULSE SERVE PORT :  %v \n", *r_m.pulse_server_port)
	fmt.Printf("   -- LOG CLIENT PORT :  %v \n", *r_m.log_server_addr)
	fmt.Printf("   -- CONTROL SERVE PORT :  %v \n", *r_m.ctrl_server_port)
	fmt.Printf("   -- COORDINATION SERVE PORT :  %v \n", *r_m.coordinate_server_port)
	fmt.Printf("   -- PASTURE :  \n")
	r_m.muster.show()
}

func (l_m *local_muster) show() {
	fmt.Printf("-- LOCAL MUSTER :  %v\n   %v \n", l_m.id, l_m)
	fmt.Printf("-- ROLE : %v \n", l_m.role)
	fmt.Printf("-- NODE :  %v \n", l_m.node)
	fmt.Printf("   -- PULSE SERVE PORT :  %v \n", *l_m.pulse_server_addr)
	fmt.Printf("   -- LOG CLIENT PORT :  %v \n", *l_m.log_server_port)
	fmt.Printf("   -- CONTROL SERVE PORT :  %v \n", *l_m.ctrl_server_addr)
	fmt.Printf("   -- COORDINATION SERVE PORT :  %v \n", *l_m.coordinate_server_addr)
	fmt.Printf("   -- PASTURE :  \n")
	l_m.muster.show()
}

func (m *muster) show() {
	for sheep_id, _ := range(m.pasture) {
		fmt.Printf("      -- SHEEP : %v \n", sheep_id)
		fmt.Printf("         -- LOGS :  \n")
		for log_id, _ := range(m.pasture[sheep_id].logs) {
			fmt.Printf("            %v \n", m.pasture[sheep_id].logs[log_id])
		} 
		fmt.Printf("         -- CONTROLS :  \n")
		for ctrl_id, _ := range(m.pasture[sheep_id].controls) {
			fmt.Printf("            %v \n", m.pasture[sheep_id].controls[ctrl_id])
		}
	}
	fmt.Println()
}



