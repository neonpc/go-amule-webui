package ec

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	conn *Conn
}

func NewClient(host string, port int, password string) (*Client, error) {
	conn, err := Dial(host, port, password, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	if err := conn.Authenticate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("auth: %w", err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

type DownloadEntry struct {
	Hash     string  `json:"hash"`
	Name     string  `json:"name"`
	Size     uint64  `json:"size"`
	Done     uint64  `json:"done"`
	Speed    uint32  `json:"speed"`
	Progress float64 `json:"progress"`
	Status   string  `json:"status"`
	Sources  uint32  `json:"sources"`
	Priority uint8   `json:"priority"`
	Category string  `json:"category"`
	Paused   bool    `json:"paused"`
}

type UploadEntry struct {
	Name     string `json:"name"`
	Client   string `json:"client"`
	Speed    uint32 `json:"speed"`
	Uploaded uint64 `json:"uploaded"`
}

type SharedFile struct {
	Hash        string  `json:"hash"`
	Name        string  `json:"name"`
	Size        uint64  `json:"size"`
	Requests    uint32  `json:"requests"`
	Transfers   uint32  `json:"transfers"`
	Priority    uint8   `json:"priority"`
	LastXfer    uint64  `json:"last_xfer"`
	AllXfer     uint64  `json:"all_xfer"`
}

type ServerEntry struct {
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Address  string `json:"address"`
	IP       string `json:"ip"`
	Port     uint16 `json:"port"`
	Users    uint32 `json:"users"`
	Files    uint32 `json:"files"`
}

type SearchResult struct {
	Hash    string `json:"hash"`
	Name    string `json:"name"`
	Size    uint64 `json:"size"`
	Sources uint32 `json:"sources"`
}

type StatusInfo struct {
	ED2KConnected   bool   `json:"ed2k_connected"`
	ED2KServer      string `json:"ed2k_server"`
	ED2KID          string `json:"ed2k_id"`
	KadConnected    bool   `json:"kad_connected"`
	KadFirewalled   bool   `json:"kad_firewalled"`
	DLSpeed         uint32 `json:"dl_speed"`
	ULSpeed         uint32 `json:"ul_speed"`
	QueueCount      uint32 `json:"queue_count"`
	SourceCount     uint32 `json:"source_count"`
}

func (c *Client) GetStatus() (*StatusInfo, error) {
	resp, err := c.conn.SendRequest(&Packet{
		OpCode: OpGetConnState,
		Tags: []Tag{
			newUint32Tag(TagDetailLevel, uint32(DetailWEB)),
		},
	})
	if err != nil {
		return nil, err
	}
	if resp.OpCode == OpFailed {
		return nil, fmt.Errorf("server returned failed")
	}

	status := &StatusInfo{}

	for _, tag := range resp.Tags {
		switch tag.Name {
		case TagStatsConnState:
			status.ED2KConnected = tag.UintValue() == 1
		case TagStatsKadState:
			status.KadConnected = tag.UintValue() == 1 || tag.UintValue() == 3
			status.KadFirewalled = tag.UintValue() == 2 || tag.UintValue() == 3
		case TagClientName:
			status.ED2KServer = tag.StringValue()
		}
	}
	_ = status.ED2KID

	return status, nil
}

func (c *Client) GetDownloads() ([]DownloadEntry, error) {
	resp, err := c.conn.SendRequest(&Packet{
		OpCode: OpGetDloadQueue,
		Tags: []Tag{
			newUint32Tag(TagDetailLevel, uint32(DetailWEB)),
		},
	})
	if err != nil {
		return nil, err
	}
	if resp.OpCode == OpFailed {
		return nil, fmt.Errorf("get downloads failed")
	}

	if resp.OpCode != OpDloadQueue {
		return nil, fmt.Errorf("unexpected response: %s", resp.OpCode)
	}

	var downloads []DownloadEntry
	partfiles := resp.ChildrenByName(TagPartfile)

	for _, pf := range partfiles {
		var entry DownloadEntry

		if name := pf.ChildByName(TagPartfileName); name != nil {
			entry.Name = name.StringValue()
		}
		if hash := pf.ChildByName(TagPartfileHash); hash != nil {
			entry.Hash = fmt.Sprintf("%x", hash.HashValue())
		}
		if size := pf.ChildByName(TagPartfileSize); size != nil {
			entry.Size = size.UintValue()
		}
		if done := pf.ChildByName(TagPartfileDone); done != nil {
			entry.Done = done.UintValue()
		}
		if speed := pf.ChildByName(TagPartfileSpeed); speed != nil {
			entry.Speed = uint32(speed.UintValue())
		}
		if src := pf.ChildByName(TagPartfileSources); src != nil {
			entry.Sources = uint32(src.UintValue())
		}
		if prio := pf.ChildByName(TagPartfilePrio); prio != nil {
			entry.Priority = uint8(prio.UintValue())
		}
		if cat := pf.ChildByName(TagPartfileCat); cat != nil {
			entry.Category = cat.StringValue()
		}
		if paused := pf.ChildByName(TagPartfilePaused); paused != nil {
			entry.Paused = paused.UintValue() == 1
		}

		if entry.Size > 0 {
			entry.Progress = float64(entry.Done) / float64(entry.Size) * 100.0
		}

		if entry.Paused {
			entry.Status = "paused"
		} else if entry.Speed > 0 {
			entry.Status = "downloading"
		} else {
			entry.Status = "waiting"
		}

		downloads = append(downloads, entry)
	}

	return downloads, nil
}

func (c *Client) CancelDownload(hash string) error {
	return c.downloadAction(hash, 4)
}

func (c *Client) PauseDownload(hash string) error {
	return c.downloadAction(hash, 5)
}

func (c *Client) ResumeDownload(hash string) error {
	return c.downloadAction(hash, 2)
}

func (c *Client) downloadAction(hash string, action uint8) error {
	var h [16]byte
	fmt.Sscanf(hash, "%16x", &h)

	resp, err := c.conn.SendRequest(&Packet{
		OpCode: OpMiscData,
		Tags: []Tag{
			newContainerTag(TagPartfile,
				newUint8Tag(TagPartfilePrio, action),
				Tag{Name: TagPartfileHash, Type: TagTypeHash16, Data: h},
			),
		},
	})
	if err != nil {
		return err
	}
	if resp.OpCode == OpFailed {
		return fmt.Errorf("action failed")
	}
	return nil
}

func (c *Client) GetUploads() ([]UploadEntry, error) {
	resp, err := c.conn.SendRequest(&Packet{
		OpCode: OpGetUloadQueue,
		Tags: []Tag{
			newUint32Tag(TagDetailLevel, uint32(DetailWEB)),
		},
	})
	if err != nil {
		return nil, err
	}
	if resp.OpCode == OpFailed {
		return nil, fmt.Errorf("get uploads failed")
	}

	var uploads []UploadEntry
	clients := resp.ChildrenByName(TagClient)

	for _, cl := range clients {
		var entry UploadEntry
		if name := cl.ChildByName(TagClientNameF); name != nil {
			entry.Client = name.StringValue()
		}
		if uploaded := cl.ChildByName(TagClientUploaded); uploaded != nil {
			entry.Uploaded = uploaded.UintValue()
		}
		if speed := cl.ChildByName(TagClientSpeed); speed != nil {
			entry.Speed = uint32(speed.UintValue())
		}
		uploads = append(uploads, entry)
	}

	return uploads, nil
}

func (c *Client) GetSharedFiles() ([]SharedFile, error) {
	resp, err := c.conn.SendRequest(&Packet{
		OpCode: OpGetSharedFiles,
		Tags: []Tag{
			newUint32Tag(TagDetailLevel, uint32(DetailWEB)),
		},
	})
	if err != nil {
		return nil, err
	}
	if resp.OpCode == OpFailed {
		return nil, fmt.Errorf("get shared files failed")
	}

	var files []SharedFile
	knownfiles := resp.ChildrenByName(TagKnownfile)

	for _, kf := range knownfiles {
		var entry SharedFile
		if name := kf.ChildByName(TagKnownfileName); name != nil {
			entry.Name = name.StringValue()
		}
		if hash := kf.ChildByName(TagKnownfileHash); hash != nil {
			entry.Hash = fmt.Sprintf("%x", hash.HashValue())
		}
		if size := kf.ChildByName(TagKnownfileSize); size != nil {
			entry.Size = size.UintValue()
		}
		files = append(files, entry)
	}

	return files, nil
}

func (c *Client) GetServers() ([]ServerEntry, error) {
	resp, err := c.conn.SendRequest(&Packet{
		OpCode: OpGetServerList,
		Tags: []Tag{
			newUint32Tag(TagDetailLevel, uint32(DetailWEB)),
		},
	})
	if err != nil {
		return nil, err
	}
	if resp.OpCode == OpFailed {
		return nil, fmt.Errorf("get servers failed")
	}

	var servers []ServerEntry
	serverTags := resp.ChildrenByName(TagServer)

	for _, st := range serverTags {
		var entry ServerEntry
		if name := st.ChildByName(TagServerName); name != nil {
			entry.Name = name.StringValue()
		}
		if desc := st.ChildByName(TagServerDesc); desc != nil {
			entry.Desc = desc.StringValue()
		}
		if addr := st.ChildByName(TagServerAddress); addr != nil {
			entry.Address = addr.StringValue()
			if parts := strings.Split(entry.Address, ":"); len(parts) == 2 {
				entry.IP = parts[0]
				p, _ := strconv.ParseUint(parts[1], 10, 16)
				entry.Port = uint16(p)
			}
		}
		if users := st.ChildByName(TagServerUsers); users != nil {
			entry.Users = uint32(users.UintValue())
		}
		if files := st.ChildByName(TagServerFiles); files != nil {
			entry.Files = uint32(files.UintValue())
		}
		servers = append(servers, entry)
	}

	return servers, nil
}

func (c *Client) Search(query string, searchType string, avail string, minSize, maxSize uint64) error {
	var stype uint32
	switch strings.ToLower(searchType) {
	case "global":
		stype = 1
	case "kad":
		stype = 2
	default:
		stype = 0
	}

	_, err := c.conn.SendRequest(&Packet{
		OpCode: OpSearchStart,
		Tags: []Tag{
			newStringTag(TagString, query),
			newUint32Tag(TagDetailLevel, uint32(DetailWEB)),
			newUint32Tag(TagPartfileStatus, stype),
		},
	})
	if err != nil {
		return err
	}

	_ = avail
	_ = minSize
	_ = maxSize
	return nil
}

func (c *Client) GetSearchResults() ([]SearchResult, error) {
	resp, err := c.conn.SendRequest(&Packet{
		OpCode: OpSearchResults,
		Tags: []Tag{
			newUint32Tag(TagDetailLevel, uint32(DetailWEB)),
		},
	})
	if err != nil {
		return nil, err
	}
	if resp.OpCode == OpFailed {
		return nil, fmt.Errorf("get search results failed")
	}

	var results []SearchResult
	searchfiles := resp.ChildrenByName(TagSearchfile)

	for _, sf := range searchfiles {
		var entry SearchResult
		if name := sf.ChildByName(TagSearchfileName); name != nil {
			entry.Name = name.StringValue()
		}
		if hash := sf.ChildByName(TagSearchfileHash); hash != nil {
			entry.Hash = fmt.Sprintf("%x", hash.HashValue())
		}
		if size := sf.ChildByName(TagSearchfileSize); size != nil {
			entry.Size = size.UintValue()
		}
		if src := sf.ChildByName(TagSearchfileSources); src != nil {
			entry.Sources = uint32(src.UintValue())
		}
		results = append(results, entry)
	}

	return results, nil
}

func (c *Client) StopSearch() error {
	_, err := c.conn.SendRequest(&Packet{
		OpCode: OpSearchStop,
	})
	return err
}

func (c *Client) DownloadSearchResult(hash string, categoryIdx int) error {
	var h [16]byte
	fmt.Sscanf(hash, "%16x", &h)

	_, err := c.conn.SendRequest(&Packet{
		OpCode: OpMiscData,
		Tags: []Tag{
			Tag{Name: TagSearchfileHash, Type: TagTypeHash16, Data: h},
			newUint32Tag(TagPartfileCat, uint32(categoryIdx)),
		},
	})
	return err
}

func (c *Client) GetStats() (map[string]interface{}, error) {
	resp, err := c.conn.SendRequest(&Packet{
		OpCode: OpStatReq,
		Tags: []Tag{
			newUint32Tag(TagDetailLevel, uint32(DetailWEB)),
		},
	})
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	for _, tag := range resp.Tags {
		switch tag.Name {
		case TagStatsULSpeed:
			stats["ul_speed"] = tag.UintValue()
		case TagStatsDLSpeed:
			stats["dl_speed"] = tag.UintValue()
		case TagStatsULData:
			stats["ul_data"] = tag.UintValue()
		case TagStatsDLData:
			stats["dl_data"] = tag.UintValue()
		}
	}

	return stats, nil
}

func (c *Client) GetLog() ([]string, error) {
	resp, err := c.conn.SendRequest(&Packet{
		OpCode: OpGetLog,
	})
	if err != nil {
		return nil, err
	}

	var lines []string
	for _, tag := range resp.Tags {
		if tag.Name == TagString {
			lines = append(lines, tag.StringValue())
		}
	}
	return lines, nil
}

func (c *Client) GetStatsTree() ([]Tag, error) {
	resp, err := c.conn.SendRequest(&Packet{
		OpCode: OpGetStatsTree,
	})
	if err != nil {
		return nil, err
	}

	return resp.ChildrenByName(TagStatsTreeNode), nil
}

func (c *Client) DisconnectED2K() error {
	_, err := c.conn.SendRequest(&Packet{
		OpCode: OpMiscData,
		Tags: []Tag{
			newUint32Tag(TagPartfileStatus, 0),
		},
	})
	return err
}

func (c *Client) KadStart() error {
	_, err := c.conn.SendRequest(&Packet{
		OpCode: OpKadStart,
	})
	return err
}

func (c *Client) KadStop() error {
	_, err := c.conn.SendRequest(&Packet{
		OpCode: OpKadStop,
	})
	return err
}
