package main

import (
	"fmt"
	"time"
	"strconv"
	"context"
	"flag"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
)

/************************************/

/* 
   This function initializes 1) a generic shepherd with a muster representation
   for each node under the shepherd's supervision and 2) a general log and control 
   representation for each core under a muster's supervision. 
*/
func (s *shepherd) init(nodes []node) {
	s.musters = make(map[string]*muster)
	s.local_musters = make(map[string]*local_muster)
	s.hb_chan = make(chan *pb.HeartbeatReply)
	for n := 0; n < len(nodes); n++ {
		/* init 1 muster for each node */
		m_id := "muster-" + nodes[n].ip
		m_n := muster{id: m_id, node: nodes[n],
//				logs: make(map[string]*log), 
//				controls: make(map[string]*control),
				pasture: make(map[string]*sheep),
				hb_chan: make(chan *pb.HeartbeatReply),
				process_buff_chan: make(chan []string),
				compute_ctrl_chan: make(chan string),
				ready_ctrl_chan: make(chan string),
				log_sync_port: flag.Int("log_sync_port_" + m_id, 
							nodes[n].log_sync_port,
							"local muster log syncing server port"),
				remote_ctrl_addr: flag.String("remote_ctrl_addr_" + m_id,
							      "localhost:" + strconv.Itoa(nodes[n].ctrl_port),
							      "address of one remote muster control server"),
				remote_muster_addr: flag.String("remote_muster_addr_" + m_id,
								"localhost:" + strconv.Itoa(nodes[n].pulse_port),
								"address of one remote muster pulse server")}
		/* init 1 sheep for each muster core */
		var c uint8
		for c = 0; c < m_n.ncores; c++ {
			sheep_id := strconv.Itoa(int(c)) + "-" + m_n.ip
			sheep_c := sheep{id: sheep_id, core: c,
					 logs: make(map[string]*log), 
					 controls: make(map[string]*control)}
			m_n.pasture[sheep_id] = &sheep_c
		}
		s.musters[m_id] = &m_n
	}
}


/* This function receives and processes incoming messages
   on a shepherd's heartbeat channel. This channel is unbuffered.
   It receives a message each time a remote muster responds
   to a heartbeat RPC request from the shepherd.
*/
func (s *shepherd) listen_heartbeats() {
	fmt.Printf("-- STARTING HEARTBEAT LISTENER :  %v ... ... ...\n", s.id)
	for {
		for _, m := range(s.musters) {
			select {
			case r := <- s.musters[m.id].hb_chan:
				m_id := r.GetMusterReply()
				fmt.Println("------HB-REP --", m_id, r.GetShepRequest())
			default:
			}
		}
	}
}


/* This function establishes a connection between local
   muster 'm' and its remote muster mirror.
*/
func (s *shepherd) start_local_pulser(l_m local_muster) {
	fmt.Printf("-- STARTING LOCAL PULSER :  %v\n", l_m.id)
	conn, err := grpc.Dial(*l_m.remote_muster_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("** ** ** ERROR: %v could not create local connection to remote muster %s:\n** ** ** %v\n", s.id, l_m.id, err)
	}
	c := pb.NewPulseClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	fmt.Printf("-- %v -- Initialized pulse client \n", l_m.id)
	go l_m.pulse(conn, c, ctx, cancel)
}


func (s *shepherd) start_local_logger(l_m local_muster) {
	fmt.Printf("-- STARTING LOCAL LOGGER :  %v\n", l_m.id)
	go l_m.log()
}

//func (s *shepherd) start_local_controller(l_m local_muster) {
//	fmt.Println("-------------------------------------------------------------")
//	fmt.Printf("-- %v -- STARTING LOCAL CONTROLLER %v\n", s.id, l_m.id)
//	fmt.Println("-------------------------------------------------------------")
//
//	conn, err := grpc.Dial(*l_m.remote_ctrl_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
//	if err != nil {
//		fmt.Printf("** ** ** ERROR: %v could not create local connection to remote controller %s:\n** ** ** %v\n", s.id, l_m.id, err)
//	}
//	c := pb.NewControlClient(conn)
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
//
//	go l_m.control(conn, c, ctx, cancel)
//}


/* This function starts all local threads relevant to 
   a shepherd's musters. It also establishes a connection
   between every local-remote muster pair. These connections
   are used by the heartbeat protocol between the shepherd
   and each of its musters.
*/
func (s *shepherd) deploy_musters() {
	flag.Parse()
	s.pulsers = make(map[string]pb.PulseClient)
	s.conn_remotes = make(map[string]*grpc.ClientConn)
	s.ctx_remotes = make(map[string]context.Context)
	s.cancel_remotes = make(map[string]context.CancelFunc)
	for _, m := range(s.musters) {	
		l_m := local_muster{muster: *m}
		s.start_local_pulser(l_m)
		s.start_local_logger(l_m)
//		go l_m.log()
//		s.start_local_controller(l_m)
	}
	go s.listen_heartbeats()
}




