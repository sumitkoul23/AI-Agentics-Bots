package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	meshPort      = 9898          // UDP discovery port
	meshMulticast = "239.255.9.9" // LAN multicast group
	meshAnnounce  = "PRIYA_SWARM" // beacon prefix
	meshTTL       = 90 * time.Second
)

// Peer represents another Bodhi Hub node on the LAN.
type Peer struct {
	Addr      string    // "192.168.x.x:8080"
	Hostname  string
	LastSeen  time.Time
	Available bool
}

// Mesh handles peer discovery, heartbeats, and cross-device routing/sync.
type Mesh struct {
	self    string // "hostname:port"
	port    string // HTTP port of this node
	peers   map[string]*Peer
	mu      sync.RWMutex
	mem     *Memory
	http    *http.Client
	stopCh  chan struct{}
	swarm   *Swarm
}

func NewMesh(port string, mem *Memory, swarm *Swarm) *Mesh {
	hostname, _ := os.Hostname()
	return &Mesh{
		self:   fmt.Sprintf("%s:%s", hostname, port),
		port:   port,
		peers:  make(map[string]*Peer),
		mem:    mem,
		http:   &http.Client{Timeout: 5 * time.Second},
		stopCh: make(chan struct{}),
		swarm:  swarm,
	}
}

// Start launches discovery listener, broadcaster, and gossip ticker.
func (m *Mesh) Start() {
	go m.listenAnnouncements()
	go m.broadcastLoop()
	go m.gossipLoop()
	log.Printf("[Mesh] started on %s (multicast %s:%d)", m.self, meshMulticast, meshPort)
}

// Stop shuts down the mesh.
func (m *Mesh) Stop() {
	close(m.stopCh)
}

// ── Discovery ──────────────────────────────────────────────────────────────────

func (m *Mesh) broadcastLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	m.sendAnnounce() // announce immediately on start
	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.sendAnnounce()
			m.pruneDeadPeers()
		}
	}
}

func (m *Mesh) sendAnnounce() {
	addr := &net.UDPAddr{IP: net.ParseIP(meshMulticast), Port: meshPort}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return
	}
	defer conn.Close()
	msg := fmt.Sprintf("%s:%s", meshAnnounce, m.self)
	conn.Write([]byte(msg))
}

func (m *Mesh) listenAnnouncements() {
	group := net.ParseIP(meshMulticast)
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}

	// Try to join multicast on each interface
	var conn *net.UDPConn
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		c, err := net.ListenMulticastUDP("udp4", &iface, &net.UDPAddr{
			IP:   group,
			Port: meshPort,
		})
		if err == nil {
			conn = c
			break
		}
	}

	// Fallback to broadcast on any interface
	if conn == nil {
		c, err := net.ListenUDP("udp4", &net.UDPAddr{Port: meshPort})
		if err != nil {
			log.Printf("[Mesh] listen error: %v — peer discovery disabled", err)
			return
		}
		conn = c
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Time{}) // no deadline

	buf := make([]byte, 256)
	for {
		select {
		case <-m.stopCh:
			return
		default:
		}
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		m.handleAnnounce(string(buf[:n]), src)
	}
}

func (m *Mesh) handleAnnounce(msg string, src *net.UDPAddr) {
	if !strings.HasPrefix(msg, meshAnnounce+":") {
		return
	}
	peerSelf := strings.TrimPrefix(msg, meshAnnounce+":")
	if peerSelf == m.self {
		return // ignore our own announce
	}

	// Extract the HTTP address: use the source IP + port from the announce payload
	parts := strings.SplitN(peerSelf, ":", 2)
	if len(parts) != 2 {
		return
	}
	httpAddr := src.IP.String() + ":" + parts[1]

	m.mu.Lock()
	if existing, ok := m.peers[httpAddr]; ok {
		existing.LastSeen = time.Now()
		existing.Available = true
	} else {
		m.peers[httpAddr] = &Peer{
			Addr:      httpAddr,
			Hostname:  parts[0],
			LastSeen:  time.Now(),
			Available: true,
		}
		log.Printf("[Mesh] peer discovered: %s (%s)", parts[0], httpAddr)
	}
	m.mu.Unlock()
}

func (m *Mesh) pruneDeadPeers() {
	deadline := time.Now().Add(-meshTTL)
	m.mu.Lock()
	for addr, p := range m.peers {
		if p.LastSeen.Before(deadline) {
			log.Printf("[Mesh] peer lost: %s", addr)
			delete(m.peers, addr)
		}
	}
	m.mu.Unlock()
}

// ── Gossip / memory sync ───────────────────────────────────────────────────────

func (m *Mesh) gossipLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.syncMemoryFromPeers()
		}
	}
}

func (m *Mesh) syncMemoryFromPeers() {
	m.mu.RLock()
	addrs := make([]string, 0, len(m.peers))
	for addr := range m.peers {
		addrs = append(addrs, addr)
	}
	m.mu.RUnlock()

	for _, addr := range addrs {
		m.mergeFromPeer(addr)
	}
}

func (m *Mesh) mergeFromPeer(addr string) {
	resp, err := m.http.Get("http://" + addr + "/memory")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return
	}

	var remote MemoryData
	if err := json.Unmarshal(body, &remote); err != nil {
		return
	}

	merged := 0
	for k, v := range remote.Facts {
		if existing := m.mem.GetFacts()[k]; existing == "" {
			m.mem.Learn(k, v)
			merged++
		}
	}
	for k, v := range remote.Preferences {
		if existing := m.mem.GetPreferences()[k]; existing == "" {
			m.mem.Set(k, v)
			merged++
		}
	}
	if merged > 0 {
		m.mem.Save()
		log.Printf("[Mesh] synced %d entries from %s", merged, addr)
	}
}

// ── Distributed routing ────────────────────────────────────────────────────────

// Delegate sends a chat message to a peer node and returns its reply.
// Used when local agent queues are saturated.
func (m *Mesh) Delegate(agentID, input string) (string, bool) {
	peer := m.pickPeer()
	if peer == nil {
		return "", false
	}

	payload, _ := json.Marshal(map[string]string{
		"message": input,
		"agent":   agentID,
	})
	resp, err := m.http.Post(
		"http://"+peer.Addr+"/chat",
		"application/json",
		bytes.NewReader(payload),
	)
	if err != nil {
		m.mu.Lock()
		peer.Available = false
		m.mu.Unlock()
		return "", false
	}
	defer resp.Body.Close()

	var result struct {
		Reply string `json:"reply"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", false
	}
	log.Printf("[Mesh] delegated %s to %s", agentID, peer.Addr)
	return result.Reply, true
}

func (m *Mesh) pickPeer() *Peer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.peers {
		if p.Available {
			return p
		}
	}
	return nil
}

// Peers returns a snapshot of all known peers.
func (m *Mesh) Peers() []*Peer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Peer, 0, len(m.peers))
	for _, p := range m.peers {
		out = append(out, p)
	}
	return out
}

// Status returns a human-readable mesh summary.
func (m *Mesh) Status() string {
	peers := m.Peers()
	if len(peers) == 0 {
		return fmt.Sprintf("Mesh: this node only (%s)\nNo other Bodhi Hub nodes found on the LAN yet.", m.self)
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Mesh: %s  +  %d peer(s)\n\n", m.self, len(peers)))
	for _, p := range peers {
		age := time.Since(p.LastSeen).Round(time.Second)
		sb.WriteString(fmt.Sprintf("  %s (%s) — last seen %s ago\n", p.Hostname, p.Addr, age))
	}
	return sb.String()
}
