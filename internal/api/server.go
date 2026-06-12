package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/neonpc/go-amule-webui/internal/ec"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Server struct {
	client     *ec.Client
	mu         sync.RWMutex
	amuleHost  string
	amulePort  int
	amulePass  string
	listenAddr string
	clients    map[chan []byte]struct{}
	register   chan chan []byte
	unregister chan chan []byte
}

func NewServer(host string, port int, pass string, listen string) *Server {
	return &Server{
		amuleHost:  host,
		amulePort:  port,
		amulePass:  pass,
		listenAddr: listen,
		clients:    make(map[chan []byte]struct{}),
		register:   make(chan chan []byte),
		unregister: make(chan chan []byte),
	}
}

func (s *Server) Connect() error {
	client, err := ec.NewClient(s.amuleHost, s.amulePort, s.amulePass)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	s.mu.Lock()
	if s.client != nil {
		s.client.Close()
	}
	s.client = client
	s.mu.Unlock()
	return nil
}

func (s *Server) Run() error {
	if err := s.Connect(); err != nil {
		return fmt.Errorf("initial connect: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/downloads", s.handleDownloads)
	mux.HandleFunc("/api/uploads", s.handleUploads)
	mux.HandleFunc("/api/shared", s.handleShared)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/search/results", s.handleSearchResults)
	mux.HandleFunc("/api/search/stop", s.handleSearchStop)
	mux.HandleFunc("/api/servers", s.handleServers)
	mux.HandleFunc("/api/kad", s.handleKad)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/api/log", s.handleLog)
	mux.HandleFunc("/ws", s.handleWS)

	go s.runWSHub()
	go s.periodicPush()

	log.Printf("Listening on %s", s.listenAddr)
	log.Printf("Connected to aMule at %s:%d", s.amuleHost, s.amulePort)
	return http.ListenAndServe(s.listenAddr, corsMiddleware(mux))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func sendJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func sendError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (s *Server) getClient() (*ec.Client, error) {
	s.mu.RLock()
	c := s.client
	s.mu.RUnlock()
	if c == nil {
		return nil, fmt.Errorf("not connected to aMule")
	}
	return c, nil
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	st, err := c.GetStatus()
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	stats, _ := c.GetStats()
	if stats != nil {
		if v, ok := stats["ul_speed"]; ok {
			st.ULSpeed = uint32(v.(uint64))
		}
		if v, ok := stats["dl_speed"]; ok {
			st.DLSpeed = uint32(v.(uint64))
		}
	}
	sendJSON(w, st)
}

func (s *Server) handleDownloads(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if r.Method == "POST" {
		hash := r.URL.Query().Get("hash")
		action := r.URL.Query().Get("action")
		switch action {
		case "pause":
			err = c.PauseDownload(hash)
		case "resume":
			err = c.ResumeDownload(hash)
		case "cancel":
			err = c.CancelDownload(hash)
		default:
			sendError(w, http.StatusBadRequest, "unknown action: "+action)
			return
		}
		if err != nil {
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		sendJSON(w, map[string]string{"status": "ok"})
		return
	}
	dl, err := c.GetDownloads()
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if dl == nil {
		dl = []ec.DownloadEntry{}
	}
	sendJSON(w, dl)
}

func (s *Server) handleUploads(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	ul, err := c.GetUploads()
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if ul == nil {
		ul = []ec.UploadEntry{}
	}
	sendJSON(w, ul)
}

func (s *Server) handleShared(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	sf, err := c.GetSharedFiles()
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sf == nil {
		sf = []ec.SharedFile{}
	}
	sendJSON(w, sf)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		Query      string `json:"query"`
		SearchType string `json:"type"`
		Avail      string `json:"avail"`
		MinSize    uint64 `json:"min_size"`
		MaxSize    uint64 `json:"max_size"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.SearchType == "" {
		req.SearchType = "global"
	}
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if err := c.Search(req.Query, req.SearchType, req.Avail, req.MinSize, req.MaxSize); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, map[string]string{"status": "search started"})
}

func (s *Server) handleSearchResults(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	res, err := c.GetSearchResults()
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if res == nil {
		res = []ec.SearchResult{}
	}
	sendJSON(w, res)
}

func (s *Server) handleSearchStop(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if err := c.StopSearch(); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleServers(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	sv, err := c.GetServers()
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sv == nil {
		sv = []ec.ServerEntry{}
	}
	sendJSON(w, sv)
}

func (s *Server) handleKad(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if r.Method == "POST" {
		action := r.URL.Query().Get("action")
		switch action {
		case "start":
			err = c.KadStart()
		case "stop":
			err = c.KadStop()
		default:
			sendError(w, http.StatusBadRequest, "unknown action")
			return
		}
		if err != nil {
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		sendJSON(w, map[string]string{"status": "ok"})
		return
	}
	st, err := c.GetStatus()
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, map[string]interface{}{
		"connected":  st.KadConnected,
		"firewalled": st.KadFirewalled,
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if r.URL.Query().Get("tree") == "1" {
		nodes, err := c.GetStatsTree()
		if err != nil {
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		sendJSON(w, nodes)
		return
	}
	stats, err := c.GetStats()
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, stats)
}

func (s *Server) handleLog(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	lines, err := c.GetLog()
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if lines == nil {
		lines = []string{}
	}
	sendJSON(w, lines)
}

func (s *Server) runWSHub() {
	for {
		select {
		case ch := <-s.register:
			s.mu.Lock()
			s.clients[ch] = struct{}{}
			s.mu.Unlock()
		case ch := <-s.unregister:
			s.mu.Lock()
			delete(s.clients, ch)
			s.mu.Unlock()
		}
	}
}

func (s *Server) broadcast(data []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.clients {
		select {
		case ch <- data:
		default:
		}
	}
}

func (s *Server) periodicPush() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		c, err := s.getClient()
		if err != nil {
			continue
		}
		stats, err := c.GetStats()
		if err != nil {
			continue
		}
		data, _ := json.Marshal(map[string]interface{}{
			"type":  "speed",
			"stats": stats,
		})
		s.broadcast(data)
	}
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade: %v", err)
		return
	}

	ch := make(chan []byte, 16)
	s.register <- ch

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	defer func() {
		s.unregister <- ch
		conn.Close()
	}()

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	for {
		select {
		case msg := <-ch:
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
