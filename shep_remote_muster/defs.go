package main

import (
	"fmt"
	"context"
	"os"
	"encoding/csv"
	"time"

	"google.golang.org/grpc"
	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
	pb_opt "github.com/awadyn/shep_remote_muster/shep_log_processor"
	"google.golang.org/protobuf/types/known/anypb"
)
/************************************/

var exp_timeout time.Duration = time.Second * 1200

func setup_target_resources_cX220(ncores uint8) []resource {
	target_resources := make([]resource, 0)
	mid := ncores / 2
	var i uint8
	for i = 0; i < mid - 1; i++ {
		target_resources = append(target_resources, resource{label: "core", index: i})
	}
	for i = mid; i < ncores - 1; i++ {
		target_resources = append(target_resources, resource{label: "core", index: i})
	}
	target_resources = append(target_resources, resource{label: "node", index: 0})
	return target_resources
}

type resource struct {
	label string
	index uint8
}

type node struct {
	ncores uint8
	pulse_port int
	log_port int
	ctrl_port int
	coordinate_port int
	optimizer_server_ports []int
	optimizer_client_ports []int
	ip_idx int			//differentiates musters on the same node
	ip string
	resources []resource
}


type control struct {
	ready_request_chan chan bool	//syncs access to ctrl object 

	// TODO fix; temp
	ready_ctrl_chan chan bool

	value uint64
	dirty bool
	knob string
	n_ip string
	id string

	getter func(uint8, ...string)uint64
	setter func(uint8, uint64, ...string)error
}

type control_request struct {
	sheep_id string
	ctrls map[string]uint64
}

type control_reply struct {
	done bool
	ctrls map[string]uint64
}

type optimize_setting struct {
	knob string
	val uint64
}

type reward struct {
	id string
	val float32
}

type reward_reply struct {
	rewards []reward
}

type log struct {
	ready_request_chan chan bool	//syncs access to log object 
	ready_buff_chan chan bool	//syncs access to log memory buffer
	ready_file_chan chan bool	//syncs access to log file
	ready_process_chan chan bool	//..? 
	kill_log_chan chan bool

	mem_buff *[][]uint64
	max_size uint64
	metrics []string
	n_ip string
	id string

	log_wait_factor time.Duration		// seconds to wait before updating log filesystem representative
}


type sheep struct {
	resource

	new_ctrl_chan chan map[string]uint64	//signals set new ctrls 
	done_ctrl_chan chan control_reply	//syncs application of ctrl change
	ready_ctrl_chan chan bool

	ready_metadata_chan chan bool		//syncs access to metadata of a specialized sheep

	request_log_chan chan []string		//signals get current logs
	request_ctrl_chan chan string		//signals get current ctrls

	detach_native_logger chan bool		//signals stop logging

	// TODO remove core
	core uint8
	logs map[string]*log
	controls map[string]*control
	id string

	log_f_map map[string]*os.File		//file pointer for each log object
	log_reader_map map[string]*csv.Reader	//reader pointer for each log file 
	log_writer_map map[string]*csv.Writer	//writer pointer for each log file 

	perf_data map[string][]float32		//map of <perf-id, perf-val>
}

type muster struct {
	full_buff_chan chan []string
	process_buff_chan chan []string

	request_ctrl_chan chan map[string]map[string]uint64
	done_request_chan chan bool

	new_ctrl_chan chan control_request
	request_log_chan chan []string
//	request_ctrl_chan chan []string

	node
	pulsing bool
	role string
	pasture map[string]*sheep
	id string

	native_loggers map[string]func(*sheep, *log, string)
	logs_dir string
}

type local_muster struct {
	muster
	hb_chan chan *pb.HeartbeatReply

//	request_optimize_chan chan optimize_request
//	ready_reward_chan chan reward_reply

	log_server_port *int
	ctrl_server_addr *string
	pulse_server_addr *string
	coordinate_server_addr *string

	pb.UnimplementedLogServer
}

type remote_muster struct {	// i.e. 1st level specialization of a muster
	muster

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

type start_optimize_request struct {
	args []*anypb.Any
}

type optimize_request struct {
//	m_id string
//	sheep_id string
//	settings []optimize_setting 
	args []*anypb.Any
}

type optimizer struct {
	id string
	port *int
	addr *string

	start_optimize_chan chan start_optimize_request	
	ready_optimize_chan chan bool

	request_optimize_chan chan optimize_request
	done_optimize_chan chan bool
}

type shepherd struct {
	hb_chan chan *pb.HeartbeatReply
	new_ctrl_chan chan map[string]control_request

	//complete_run_chan chan []string

	musters map[string]*muster
	local_musters map[string]*local_muster
	id string

	optimizers map[string]*optimizer
	
	pb_opt.UnimplementedLogStatsMessengerServer
}

type Shepherd interface {
	init()
	deploy_musters()
	
	process_logs()
	process_control()

	compute_control()
	complete_run()
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
	fmt.Printf("   -- PULSE SERVER PORT :  %v \n", *r_m.pulse_server_port)
	fmt.Printf("   -- LOG CLIENT PORT :  %v \n", *r_m.log_server_addr)
	fmt.Printf("   -- CONTROL SERVER PORT :  %v \n", *r_m.ctrl_server_port)
	fmt.Printf("   -- COORDINATION SERVER PORT :  %v \n", *r_m.coordinate_server_port)
	fmt.Printf("   -- PASTURE :  \n")
	r_m.muster.show()
}

func (l_m *local_muster) show() {
	fmt.Printf("-- LOCAL MUSTER :  %v\n   %v \n", l_m.id, l_m)
	fmt.Printf("-- ROLE : %v \n", l_m.role)
	fmt.Printf("-- NODE :  %v \n", l_m.node)
	fmt.Printf("   -- PULSE CLIENT PORT :  %v \n", *l_m.pulse_server_addr)
	fmt.Printf("   -- LOG SERVER PORT :  %v \n", *l_m.log_server_port)
	fmt.Printf("   -- CONTROL CLIENT PORT :  %v \n", *l_m.ctrl_server_addr)
	fmt.Printf("   -- COORDINATION CLIENT PORT :  %v \n", *l_m.coordinate_server_addr)
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

func (s *shepherd) show() {
	for id, opt := range(s.optimizers) {
		fmt.Printf("** OPTIMIZER :  %v \n", id)
		port := opt.port
		addr := opt.addr
		fmt.Printf("** ** OPTIMIZER SERVER PORT:  %v \n", *port)
		fmt.Printf("** ** OPTIMIZER CLIENT PORT:  %v \n", *addr)
	}
}

