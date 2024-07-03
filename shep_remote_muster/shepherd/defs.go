package main

import (
	"fmt"
//	"context"
	"os"
	"encoding/csv"
	"time"

//	"google.golang.org/grpc"
	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
)
/************************************/

var exp_timeout time.Duration = time.Second * 75

type node struct {
	ncores uint8
	pulse_port int
	log_sync_port int
	ctrl_port int
	coordinate_port int
	ip string
}

type log struct {
	ready_buff_chan chan bool
	done_process_chan chan bool

	mem_buff *[][]uint64
	max_size uint64
	metrics []string
	n_ip string
	id string
	/* e.g. { "log-ep-i", "10.0.0.1", ["joules", "timestamp"], 64KB, 0xdeadbeef, 0x12345678:PORT(i):10.0.0.1, i}
		0xdeadbeef: l_buff  ->  [  [x, 0]
					   [y, 1]
		   			   [z, 2], ...]  */
}

type control_request struct {
	sheep_id string
	ctrls map[string]uint64
}

type control_reply struct {
	done bool
	ctrls map[string]uint64
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
	finish_run_chan chan bool
	//done_ctrl_chan chan bool
	done_ctrl_chan chan control_reply
	logs map[string]*log
	controls map[string]*control
	id string
}

type muster struct {
	node

	pulsing bool

	hb_chan chan *pb.HeartbeatReply
	full_buff_chan chan []string
	new_ctrl_chan chan control_request
	//ready_ctrl_chan chan string

	pasture map[string]*sheep
	id string
	/* e.g. {"muster_n", {"ctrl-dvfs-i": {..}, "ctrl-itr-i": {..} ...}, {"log-ep-i": {..}, "log-ep-j": {..}, ...}, node{"10.0.0.1", 24}} */
}

type local_muster struct {
	muster

	log_server_port *int
	ctrl_server_addr *string
	pulse_server_addr *string

	out_f_map map[string](map[string]*os.File)
	out_writer_map map[string](map[string]*csv.Writer)
	out_f map[string]*os.File
	out_writer map[string]*csv.Writer

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
	process_buff_chan chan []string
	compute_ctrl_chan chan []string

	complete_run_chan chan []string

	coordinate_port *int
	pb.UnimplementedCoordinateServer

	id string
	/* e.g. {"sheperd-ep", {"muster-10.0.0.1": &muster{..}, "muster-10.0.0.2": &muster{..} ...}} */
}

type ep_shepherd struct {
	shepherd
}

type intlog_shepherd struct {
	shepherd
}

type Shepherd interface {
	init()
	process_logs()
	compute_control()
	complete_run()
}



func (l_ptr *log) show() {
	fmt.Printf("    ADDR %p ", l_ptr)
	fmt.Println("ID:", l_ptr.id, "  --  MAX_SIZE:", l_ptr.max_size, "  --  METRICS:", l_ptr.metrics)
	fmt.Printf("    -- %p L_BUFF:", l_ptr.mem_buff)
	fmt.Println(*l_ptr.mem_buff)
}
func (c_ptr *control) show() {
	fmt.Printf("    ADDR %p ", c_ptr)
	fmt.Println("ID:", c_ptr.id, "  --  KNOB:", c_ptr.knob, "  --  VALUE:", c_ptr.value)
}
func (m_ptr *muster) show() {
	fmt.Println()
	fmt.Printf("ADDR %p ", m_ptr)
	fmt.Println("ID:", m_ptr.id, "HB_CHAN:", m_ptr.hb_chan)
}
func (s_ptr *shepherd) show() {
	fmt.Printf("ADDR %p ", s_ptr)
	fmt.Println("ID:", s_ptr.id, "HB_CHAN:", s_ptr.hb_chan)
	fmt.Println("-- MUSTERS:", s_ptr.musters)
	for _, m := range(s_ptr.musters) { m.show() }
	fmt.Println()
}

