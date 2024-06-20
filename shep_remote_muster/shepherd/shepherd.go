package main

import (
	"fmt"
//	"time"
	"strconv"
//	"context"
	"flag"
//	"net"

//	"google.golang.org/grpc"
//	"google.golang.org/grpc/credentials/insecure"
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
	s.hb_chan = make(chan *pb.HeartbeatReply)
	s.process_buff_chan = make(chan []string)
	s.compute_ctrl_chan = make(chan []string)

//	s.complete_run_chan = make(chan []string)
//	s.coordinate_port = flag.Int("coordinate_port", 50081,
//				     "shepherd coordination server port")

	for n := 0; n < len(nodes); n++ {
		/* init 1 muster for each node */
		m_id := "muster-" + nodes[n].ip
		m_n := muster{id: m_id, node: nodes[n],
				pasture: make(map[string]*sheep),
				hb_chan: make(chan *pb.HeartbeatReply),
				full_buff_chan: make(chan []string),
				new_ctrl_chan: make(chan control_request),
				ready_ctrl_chan: make(chan string)}
		/* init 1 sheep for each muster core */
		var c uint8
		for c = 0; c < m_n.ncores; c++ {
			sheep_id := strconv.Itoa(int(c)) + "-" + m_n.ip
			sheep_c := sheep{id: sheep_id, core: c,
					 logs: make(map[string]*log), 
					 controls: make(map[string]*control),
//					 done_ctrl_chan: make(chan bool, 1),
					 done_ctrl_chan: make(chan control_reply, 1),
					 finish_run_chan: make(chan bool, 1)}
			m_n.pasture[sheep_id] = &sheep_c
		}
		s.musters[m_id] = &m_n
	}
}

func (s *shepherd) init_local(nodes []node) {
	s.local_musters = make(map[string]*local_muster)

	for n := 0; n < len(nodes); n++ {
		/* init 1 local muster for each node */
		m_id := "muster-" + nodes[n].ip
		m := s.musters[m_id]
		s.local_musters[m_id] = &local_muster{muster: *m,
						     log_server_port: flag.Int("log_server_port_" + m_id, 
										nodes[n].log_sync_port,
										"local muster log syncing server port"),
						     ctrl_server_addr: flag.String("ctrl_server_addr_" + m_id,
										   "localhost:" + strconv.Itoa(nodes[n].ctrl_port),
										   "address of one remote muster control server"),
						     pulse_server_addr: flag.String("pulse_server_addr_" + m_id,
										    "10.10.1.2:" + strconv.Itoa(nodes[n].pulse_port),
										    "address of one remote muster pulse server")}
	}
}

/* This function starts all threads required by a shepherd 
   to manage all of its musters and their sheep.
*/
func (s *shepherd) deploy_musters() {
	for _, l_m := range(s.local_musters) {	
		l_m.start_pulser()		// general functionality: start pulse management thread per-muster
//		l_m.start_logger()		// * specialization: start log server
//		go s.log(l_m.id)		// general functionality: start log management thread per-muster
//		go l_m.start_controller()
	}
}

/********** PULSING **********/

/* shepherd pulsing: handles pulse signals received
   on shepherd's heartbeat channel. This channel is unbuffered.
   It receives a message each time a muster pulses.
*/
func (s *shepherd) listen_heartbeats() {
	fmt.Printf("-- STARTING HEARTBEAT LISTENER :  %v ... ... ...\n", s.id)
	for {
		for _, m := range(s.musters) {
			select {
			case r := <- s.musters[m.id].hb_chan:
				m_id := r.GetMusterReply()
				fmt.Println("------HB-REP --", m_id, r.GetShepRequest())
//				s.musters[m_id].pulsing = true
			default:
			}
		}
	}
}


/********** LOGGING **********/

/* shepherd logging: handles log management signals received 
   on a muster's full_buff channel. This channel is unbuffered.
   It receives a message each time a sheep requires log management.
*/
func (s *shepherd) log(m_id string) {
	m := s.musters[m_id]
	for {
		select {
		case ids := <- m.full_buff_chan:
			sheep_id := ids[0]
			log_id := ids[1]
			s.process_buff_chan <- []string{m.id, sheep_id, log_id}
			go func() {
				<- m.pasture[sheep_id].logs[log_id].done_process_chan
				select {
				case m.pasture[sheep_id].logs[log_id].ready_buff_chan <- true:
				default:
				}
			} ()
		}
	}
}


//func (s *shepherd) control(m_id string) {
//	m := s.musters[m_id]
//	for {
//		select {
//		case new_ctrl_req := <- m.new_ctrl_chan:
//			sheep_id := new_ctrl_req.sheep_id
//			new_ctrls := new_ctrl_req.ctrls
//			
//			
//		}
//	}
//}

/**/

//func (s *shepherd) CompleteRun(ctx context.Context, in *pb.CompleteRunRequest) (*pb.CompleteRunReply, error) {
//	sheep_id := in.GetSheepId()
//	muster_id := in.GetMusterId()
//	s.complete_run_chan <- []string{muster_id, sheep_id}
//	<- s.musters[muster_id].pasture[sheep_id].finish_run_chan
//	return &pb.CompleteRunReply{RunComplete: true}, nil
//}
//
//func (s *shepherd) complete_runs() {
//	flag.Parse()
//	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *s.coordinate_port))
//	if err != nil {
//		fmt.Printf("** ** ** ERROR: %v failed to listen at %v: %v\n", s.id, *s.coordinate_port, err)
//	}
//	server := grpc.NewServer()
//	pb.RegisterCoordinateServer(server, s)
//	fmt.Printf("-- %v -- Coordination server listening at %v ... ... ...\n", s.id, lis.Addr())
//	if err := server.Serve(lis); err != nil {
//		fmt.Printf("** ** ** ERROR: %v failed to start coordination server: %v\n", s.id, err)
//	}
//}







