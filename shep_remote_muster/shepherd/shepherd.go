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
		m_id := "muster-" + nodes[n].ip
		m_n := muster{id: m_id, node: nodes[n],
				logs: make(map[string]*log), 
				controls: make(map[string]*control),
				//hb_chan: make(chan bool),
				hb_chan: make(chan *pb.HeartbeatReply),
				process_chan: make(chan string)}
		s.musters[m_id] = &m_n
	}
}

/* This function receives and processes incoming messages
   on a shepherd's heartbeat channel. This channel is unbuffered.
   It receives a message each time a remote muster responds
   to a heartbeat RPC request from the shepherd.
*/
func (s *shepherd) listen_heartbeats() {
	fmt.Printf("-- %v -- START LISTENING TO HEARTBEATS\n", s.id)
	for {
		for _, m := range(s.musters) {
			select {
			case r := <- s.musters[m.id].hb_chan:
				m_id := r.GetMusterReply()
				//s.musters[m_id].hb_chan <- true
				fmt.Println("------HB-REP--", m_id, r.GetShepRequest())
			default:
			}
		}
	}
}

func (s *shepherd) heartbeat(m muster, conn *grpc.ClientConn, cancel context.CancelFunc) {
	// all of this shepherd's open connections (one per remote muster)
	// can be closed when the heartbeat protocol terminates
	defer conn.Close()
	defer cancel()

	var counter uint32 = 0
	for {
		counter += 1
		c := s.pulsers[m.id]
		ctx := s.ctx_remotes[m.id]
		r, err := c.HeartBeat(ctx, &pb.HeartbeatRequest{ShepRequest: counter})  
		if err != nil {
			//fmt.Printf("** ** ** ERROR: %v could not send heartbeat request to remote %v:\n** ** ** %v\n", s.id, m.id, err)
			time.Sleep(time.Second/5)
			continue
//			return
		}
//		s.hb_chan <- r
		select {
		case s.musters[m.id].hb_chan <- r:
		}
		time.Sleep(time.Second/5)
	}
}
/* This function sends heartbeat RPC requests in a round-robin
   fashion to all musters under this shepherd's supervision.
   Upon RPC reply, it signals the reply heartbeat back 
   to its calling shepherd.
*/
func (s *shepherd) send_heartbeats() {
	fmt.Printf("-- %v -- START SENDING HEARTBEAT RPC\n", s.id)
	for m_id, m := range(s.musters) {
		conn := s.conn_remotes[m_id]
		cancel := s.cancel_remotes[m_id]
		go s.heartbeat(*m, conn, cancel)
	}
}

/* This function establishes a connection between local
   muster 'm' and its remote muster mirror.
*/
func (s *shepherd) start_local_pulser(m local_muster) (*grpc.ClientConn, *pb.PulseClient, *context.Context, *context.CancelFunc) {
	fmt.Println("-------------------------------------------------------------")
	fmt.Printf("-- %v -- STARTING LOCAL MUSTER %v\n", s.id, m.id)
	fmt.Println("-------------------------------------------------------------")

	conn, err := grpc.Dial(*m.remote_muster_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("** ** ** ERROR: %v could not create local connection to remote muster %s:\n** ** ** %v\n", s.id, m.id, err)
	}
	c := pb.NewPulseClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	return conn, &c, &ctx, &cancel
}

/* This function starts all local threads relevant to 
   a shepherd's musters. It also establishes a connection
   between every local-remote muster pair. These connections
   are used by the heartbeat protocol between the shepherd
   and each of its musters.
*/
func (s *shepherd) deploy_musters() {
	flag.Parse()
	port_ctr := 1
	s.conn_remotes = make(map[string]*grpc.ClientConn)
	s.pulsers = make(map[string]pb.PulseClient)
	s.ctx_remotes = make(map[string]context.Context)
	s.cancel_remotes = make(map[string]context.CancelFunc)
	for _, m := range(s.musters) {	
		l_m := local_muster{muster: *m, 
					log_sync_port: flag.Int("sync_port_"+m.id, 50060 + port_ctr, "local muster log syncing port"),
					remote_muster_addr: flag.String("remote_muster_addr_"+m.id, 
									"localhost:5005" + strconv.Itoa(port_ctr), 
									"address of one remote muster")}
		conn, c, ctx, cancel := s.start_local_pulser(l_m)
		s.local_musters[l_m.id] = &l_m
		s.conn_remotes[l_m.id] = conn
		s.pulsers[l_m.id] = *c
		s.ctx_remotes[l_m.id] = *ctx
		s.cancel_remotes[l_m.id] = *cancel
		port_ctr ++
		go l_m.log()
		go l_m.control()
	}
	go s.send_heartbeats()	
	go s.listen_heartbeats()
}



