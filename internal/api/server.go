package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/neonpc/go-amule-webui/internal/ec"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Server struct {
	client      *ec.Client
	mu          sync.RWMutex
	amuleHost   string
	amulePort   int
	amulePass   string
	listenAddr  string
	authToken   string
	clients     map[chan []byte]struct{}
	register    chan chan []byte
	unregister  chan chan []byte
	pausedHashes map[string]bool
}

func NewServer(host string, port int, pass string, listen string) *Server {
	buf := make([]byte, 32)
	rand.Read(buf)
	return &Server{
		amuleHost:  host,
		amulePort:  port,
		amulePass:  pass,
		listenAddr: listen,
		authToken:  hex.EncodeToString(buf),
		clients:    make(map[chan []byte]struct{}),
		register:   make(chan chan []byte),
		unregister:   make(chan chan []byte),
		pausedHashes: make(map[string]bool),
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

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/login", s.handleLogin)
	apiMux.HandleFunc("/status", s.authMiddleware(s.handleStatus))
	apiMux.HandleFunc("/downloads", s.authMiddleware(s.handleDownloads))
	apiMux.HandleFunc("/uploads", s.authMiddleware(s.handleUploads))
	apiMux.HandleFunc("/shared", s.authMiddleware(s.handleShared))
	apiMux.HandleFunc("/search", s.authMiddleware(s.handleSearch))
	apiMux.HandleFunc("/search/results", s.authMiddleware(s.handleSearchResults))
	apiMux.HandleFunc("/search/stop", s.authMiddleware(s.handleSearchStop))
	apiMux.HandleFunc("/servers", s.authMiddleware(s.handleServers))
	apiMux.HandleFunc("/servers/add", s.authMiddleware(s.handleServerAdd))
	apiMux.HandleFunc("/servers/connect", s.authMiddleware(s.handleServerConnect))
	apiMux.HandleFunc("/servers/remove", s.authMiddleware(s.handleServerRemove))
	apiMux.HandleFunc("/search/download", s.authMiddleware(s.handleSearchDownload))
	apiMux.HandleFunc("/ed2k", s.authMiddleware(s.handleED2K))
	apiMux.HandleFunc("/kad", s.authMiddleware(s.handleKad))
	apiMux.HandleFunc("/stats", s.authMiddleware(s.handleStats))
	apiMux.HandleFunc("/log", s.authMiddleware(s.handleLog))
	apiMux.HandleFunc("/fs/browse", s.authMiddleware(s.handleFSBrowse))
	mux.Handle("/api/", http.StripPrefix("/api", corsMiddleware(apiMux)))
	mux.HandleFunc("/ws", s.handleWS)

	dist := os.Getenv("AMULE_WEB_DIR")
	if dist == "" {
		dist = "web/dist"
	}
	if fi, err := os.Stat(dist); err == nil && fi.IsDir() {
		log.Printf("Serving web UI from %s", dist)
		fs := http.FileServer(http.Dir(dist))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				path := dist + r.URL.Path
				if _, err := os.Stat(path); os.IsNotExist(err) {
					http.ServeFile(w, r, dist+"/index.html")
					return
				}
			}
			fs.ServeHTTP(w, r)
		})
	} else {
		log.Print("No web UI found at web/dist — API only")
	}

	go s.runWSHub()
	go s.periodicPush()

	log.Printf("Listening on %s", s.listenAddr)
	log.Printf("Connected to aMule at %s:%d", s.amuleHost, s.amulePort)
	return http.ListenAndServe(s.listenAddr, mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendError(w, 405, "method not allowed")
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendError(w, 400, "invalid request body")
		return
	}
	if body.Password != s.amulePass {
		sendError(w, 401, "invalid password")
		return
	}
	sendJSON(w, map[string]string{"token": s.authToken})
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			// Also check query param for WebSocket
			auth = r.URL.Query().Get("token")
			if auth != "" {
				auth = "Bearer " + auth
			}
		}
		if auth != "Bearer "+s.authToken {
			sendError(w, 401, "unauthorized")
			return
		}
		next(w, r)
	}
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
			if err == nil {
				s.mu.Lock()
				s.pausedHashes[hash] = true
				s.mu.Unlock()
			}
		case "resume":
			err = c.ResumeDownload(hash)
			if err == nil {
				s.mu.Lock()
				delete(s.pausedHashes, hash)
				s.mu.Unlock()
			}
		case "cancel":
			err = c.CancelDownload(hash)
			if err == nil {
				s.mu.Lock()
				delete(s.pausedHashes, hash)
				s.mu.Unlock()
			}
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
	s.mu.RLock()
	for i := range dl {
		if s.pausedHashes[dl[i].Hash] {
			dl[i].Paused = true
			dl[i].Status = "paused"
		}
	}
	s.mu.RUnlock()
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

func (s *Server) handleED2K(w http.ResponseWriter, r *http.Request) {
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if r.Method != "POST" {
		sendError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	action := r.URL.Query().Get("action")
	switch action {
	case "connect":
		err = c.ConnectED2K()
	case "disconnect":
		err = c.DisconnectED2K()
	default:
		sendError(w, http.StatusBadRequest, "unknown action: "+action)
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleServerConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Address == "" {
		sendError(w, http.StatusBadRequest, "address required")
		return
	}
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if err := c.ConnectToServer(req.Address); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleServerAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		Address string `json:"address"`
		Name    string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Address == "" {
		sendError(w, http.StatusBadRequest, "address required")
		return
	}
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if err := c.AddServer(req.Address, req.Name); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleServerRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Address == "" {
		sendError(w, http.StatusBadRequest, "address required")
		return
	}
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if err := c.RemoveServer(req.Address); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleSearchDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		Hash string `json:"hash"`
		Name string `json:"name"`
		Size uint64 `json:"size"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Hash == "" || req.Name == "" {
		sendError(w, http.StatusBadRequest, "hash and name required")
		return
	}
	c, err := s.getClient()
	if err != nil {
		sendError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if err := c.DownloadSearchResult(req.Hash, req.Name, req.Size); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, map[string]string{"status": "ok"})
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

type FSEntry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

func (s *Server) handleFSBrowse(w http.ResponseWriter, r *http.Request) {
	root := r.URL.Query().Get("path")
	if root == "" {
		root = "/media"
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		sendError(w, http.StatusNotFound, "directory not found: "+err.Error())
		return
	}

	result := make([]FSEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		result = append(result, FSEntry{
			Name:    e.Name(),
			Path:    filepath.Join(root, e.Name()),
			IsDir:   e.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return result[i].Name < result[j].Name
	})

	sendJSON(w, result)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token != s.authToken {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
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
