package main

/************************************/
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
type node struct {
	ncores uint8
	ip string
}

type log struct {
	core uint8
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
	// CHANNELS
	logs map[string]*log
	controls map[string]*control
	hb_chan chan bool
	id string
	/* e.g. {"muster_n", {"ctrl-dvfs-i": {..}, "ctrl-itr-i": {..} ...}, {"log-ep-i": {..}, "log-ep-j": {..}, ...}, node{"10.0.0.1", 24}} */
}

type local_muster struct {
	muster
	remote_muster_addr *string
}

type cat struct {
	id string
	chaos uint8
}

type shepherd struct {
	musters map[string]*muster
	hb_chan chan *pb.HeartbeatReply

	pulsers map[string]pb.PulseClient
	conn_remotes map[string]*grpc.ClientConn
	ctx_remotes map[string]context.Context
	cancel_remotes map[string]context.CancelFunc

	id string
	/* e.g. {"sheperd-ep", {"muster-10.0.0.1": &muster{..}, "muster-10.0.0.2": &muster{..} ...}} */
}

type ep_shepherd struct {
	shepherd
}

type Shepherd interface {
	init()
}
/************************************/
func (l_ptr *log) show() {
	fmt.Printf("    ADDR %p ", l_ptr)
	fmt.Println("ID:", l_ptr.id, "  --  MAX_SIZE:", l_ptr.max_size, "  --  METRICS:", l_ptr.metrics)
	fmt.Printf("    -- %p L_BUFF:", l_ptr.l_buff)
	fmt.Println(*l_ptr.l_buff)
	fmt.Printf("    -- %p R_BUFF:", l_ptr.r_buff)
	fmt.Println(*l_ptr.r_buff)

}
func (c_ptr *control) show() {
	fmt.Printf("    ADDR %p ", c_ptr)
	fmt.Println("ID:", c_ptr.id, "  --  KNOB:", c_ptr.knob, "  --  VALUE:", c_ptr.value)
}
func (m_ptr *muster) show() {
	fmt.Println()
	fmt.Printf("ADDR %p ", m_ptr)
	fmt.Println("ID:", m_ptr.id, "HB_CHAN:", m_ptr.hb_chan)
	fmt.Println("------ LOGS:", m_ptr.logs)
	fmt.Println("-- CONTROLS:", m_ptr.controls) 
}
func (s_ptr *shepherd) show() {
	fmt.Printf("ADDR %p ", s_ptr)
	fmt.Println("ID:", s_ptr.id, "HB_CHAN:", s_ptr.hb_chan)
	fmt.Println("-- MUSTERS:", s_ptr.musters)
	for _, m := range(s_ptr.musters) {
		m.show()
		for _, l := range(m.logs) {l.show()}
		for _, c := range(m.controls) {c.show()}
	}
	fmt.Println()
}
/************************************/
/* 
   This function initializes 1) a general shepherd with a muster representation
   for each node under the shepherd's supervision and 2) a general log and control 
   representation for each core under a muster's supervision. 
*/
func (s *shepherd) init(nodes []node) {
	s.musters = make(map[string]*muster)
	s.hb_chan = make(chan *pb.HeartbeatReply)
	for n := 0; n < len(nodes); n++ {
		m_id := "muster-" + nodes[n].ip
		m_n := muster{id: m_id, node: nodes[n], 
				logs: make(map[string]*log), 
				controls: make(map[string]*control),
				hb_chan: make(chan bool)}
		s.musters[m_id] = &m_n
	}
}

func (s *shepherd) listen_heartbeats() {
	fmt.Println("-- Starting listen_heartbeats: ", s.id)
	for {
		select {
		case r := <- s.hb_chan:
			fmt.Println("-- Remote muster heartbeat reply: ", r.GetMusterReply(), r.GetShepRequest())
		}
	}
}

func (s *shepherd) send_heartbeats() {
	fmt.Println("-- Starting send_heartbeats: ", s.id)
	for m_id, _ := range(s.musters) {
		conn := s.conn_remotes[m_id]
		cancel := s.cancel_remotes[m_id]
		defer conn.Close()
		defer cancel()
	}
	var counter uint32 = 0
	for {
		for m_id, _ := range(s.musters) {
			counter += 1
			fmt.Println(counter, m_id)
			c := s.pulsers[m_id]
			ctx := s.ctx_remotes[m_id]
			go func(counter uint32) {
				r, err := c.HeartBeat(ctx, &pb.HeartbeatRequest{ShepRequest: counter})  
				if err != nil {
					fmt.Printf("** ** ** Shepherd could not send heartbeat request: %v\n", err)
				}
				s.hb_chan <- r
			} (counter)
			time.Sleep(time.Second/10)
		}	
	}
}

func (s *shepherd) start_local_muster(m local_muster) (*grpc.ClientConn, *pb.PulseClient, *context.Context, *context.CancelFunc) {
	fmt.Println("-------------------------------------------------------------")
	fmt.Printf("-- Starting local muster %p\n", &m)
	m.show()
	fmt.Println("-------------------------------------------------------------")

	conn, err := grpc.Dial(*m.remote_muster_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("** ** ** Shepherd could not connect to remote muster %s: %v\n", m.id, err)
	}
	c := pb.NewPulseClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	return conn, &c, &ctx, &cancel
}

func (s *shepherd) deploy_musters() {
	flag.Parse()
	port_ctr := 1
	s.conn_remotes = make(map[string]*grpc.ClientConn)
	s.pulsers = make(map[string]pb.PulseClient)
	s.ctx_remotes = make(map[string]context.Context)
	s.cancel_remotes = make(map[string]context.CancelFunc)
	for _, m := range(s.musters) {	
		l_m := local_muster{muster: *m, remote_muster_addr: flag.String("remote_muster_addr_"+m.id, 
					"localhost:5005" + strconv.Itoa(port_ctr), 
					"address of one remote muster for shepherd to connect to")}
		conn, c, ctx, cancel := s.start_local_muster(l_m)
		s.conn_remotes[l_m.id] = conn
		s.pulsers[l_m.id] = *c
		s.ctx_remotes[l_m.id] = *ctx
		s.cancel_remotes[l_m.id] = *cancel
		port_ctr ++
	}
	go s.send_heartbeats()	
	go s.listen_heartbeats()
}
/************************************/
/* 
   This function initializes a specialized shepherd for energy-and-performance 
   supervision. Each muster under this shepherd's supervision logs energy and
   performance metrics - i.e. joules and timestamp counters - and controls
   energy and performance settings - i.e. interrupt delay 
   and dynamic-voltage-frequency-scaling - for each core under its supervision.
*/
func (ep_s ep_shepherd) init() {
	for m_id, m := range(ep_s.musters) {
		var c uint8
		for c = 0; c < m.ncores; c++ {
			c_str := strconv.Itoa(int(c))
			log_id := "log-" + c_str + "-" + m.ip 
			log_c := log{id: log_id, n_ip: m.ip, core: uint8(c),
					metrics: []string{"joules", "timestamp"}, max_size: 64}
			l_buff := make([][]uint64, 1)
			r_buff := make([][]uint64, 1)
			log_c.l_buff = &l_buff
			log_c.r_buff = &r_buff
			ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + m.ip
			ctrl_itr_id := "ctrl-itr-" + c_str + "-" + m.ip
			ctrl_dvfs_c := control{id: ctrl_dvfs_id, n_ip: m.ip, core: c, knob: "dvfs", value: 0xffff}
			ctrl_itr_c := control{id: ctrl_itr_id, n_ip: m.ip, core: c, knob: "itr-delay", value: 1}

			ep_s.musters[m_id].logs[log_c.id] = &log_c
			ep_s.musters[m_id].controls[ctrl_dvfs_c.id] = &ctrl_dvfs_c
			ep_s.musters[m_id].controls[ctrl_itr_c.id] = &ctrl_itr_c
		}
	}
}
/************************************/
func main() {
	nodes := []node{{ip: "10.0.0.1", ncores: 4}, 
			{ip: "10.0.0.2", ncores: 4}}

	s := shepherd{id: "sheperd-ep"}
	s.init(nodes)
	ep_s := ep_shepherd{s}
	ep_s.init()
	ep_s.show()
	ep_s.deploy_musters()

	time.Sleep(time.Second)
}
