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
		case TagConnState:
			v := tag.UintValue()
			status.ED2KConnected = (v & 0x01) != 0
			status.KadConnected = (v & 0x04) != 0
			status.KadFirewalled = (v & 0x08) != 0
			if serverName := tag.ChildByName(TagServerName); serverName != nil {
				status.ED2KServer = serverName.StringValue()
			} else if serverTag := tag.ChildByName(TagServer); serverTag != nil {
				if serverName := serverTag.ChildByName(TagServerName); serverName != nil {
					status.ED2KServer = serverName.StringValue()
				}
			}
		}
	}

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

	for _, t := range resp.Tags {
		if t.Name != TagPartfile {
			continue
		}
		pf := t
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
		if done := pf.ChildByName(TagPartfileSizeDone); done != nil {
			entry.Done = done.UintValue()
		} else if done := pf.ChildByName(TagPartfileDone); done != nil {
			entry.Done = done.UintValue()
		}
		if speed := pf.ChildByName(TagPartfileSpeed); speed != nil {
			entry.Speed = uint32(speed.UintValue())
		}
		if src := pf.ChildByName(TagPartfileSourceCount); src != nil {
			entry.Sources = uint32(src.UintValue())
		} else if src := pf.ChildByName(TagPartfileSources); src != nil {
			entry.Sources = uint32(src.UintValue())
		}
		if prio := pf.ChildByName(TagPartfilePriority); prio != nil {
			entry.Priority = uint8(prio.UintValue())
		} else if prio := pf.ChildByName(TagPartfilePrio); prio != nil {
			entry.Priority = uint8(prio.UintValue())
		}
		if cat := pf.ChildByName(TagPartfileCat); cat != nil {
			if s := cat.StringValue(); s != "" {
				entry.Category = s
			} else {
				entry.Category = fmt.Sprintf("%d", cat.UintValue())
			}
		}
		if paused := pf.ChildByName(TagPartfileStopped); paused != nil {
			entry.Paused = paused.UintValue() == 1
		} else if paused := pf.ChildByName(TagPartfilePaused); paused != nil {
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

	for _, t := range resp.Tags {
		if t.Name != TagClient {
			continue
		}
		cl := t
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
	for _, t := range resp.Tags {
		if t.Name != TagKnownfile {
			continue
		}
		var entry SharedFile
		if name := t.ChildByName(TagPartfileName); name != nil {
			entry.Name = name.StringValue()
		}
		if hash := t.ChildByName(TagPartfileHash); hash != nil {
			entry.Hash = fmt.Sprintf("%x", hash.HashValue())
		}
		if size := t.ChildByName(TagPartfileSize); size != nil {
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

	for _, st := range resp.Tags {
		if st.Name != TagServer {
			continue
		}
		var entry ServerEntry
		if name := st.ChildByName(TagServerName); name != nil {
			entry.Name = name.StringValue()
		}
		if desc := st.ChildByName(TagServerDesc); desc != nil {
			entry.Desc = desc.StringValue()
		}
		if st.Type == TagTypeIPV4 {
			entry.Address = st.StringValue()
			if parts := strings.Split(entry.Address, ":"); len(parts) == 2 {
				entry.IP = parts[0]
				p, _ := strconv.ParseUint(parts[1], 10, 16)
				entry.Port = uint16(p)
			}
		} else if addr := st.ChildByName(TagServerAddress); addr != nil {
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

	searchTag := Tag{Name: TagSearchType, Type: TagTypeUint32, Data: stype}
	searchTag.Children = []Tag{
		newStringTag(TagSearchName, query),
	}
	if minSize > 0 {
		searchTag.Children = append(searchTag.Children, newUint64Tag(TagSearchMinSize, minSize))
	}
	if maxSize > 0 {
		searchTag.Children = append(searchTag.Children, newUint64Tag(TagSearchMaxSize, maxSize))
	}

	_, err := c.conn.SendRequest(&Packet{
		OpCode: OpSearchStart,
		Tags: []Tag{
			searchTag,
		},
	})
	if err != nil {
		return err
	}

	_ = avail
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

	for _, t := range resp.Tags {
		if t.Name != TagSearchfile {
			continue
		}
		sf := t
		var entry SearchResult
		if name := sf.ChildByName(TagPartfileName); name != nil {
			entry.Name = name.StringValue()
		}
		if hash := sf.ChildByName(TagPartfileHash); hash != nil {
			entry.Hash = fmt.Sprintf("%x", hash.HashValue())
		}
		if size := sf.ChildByName(TagPartfileSize); size != nil {
			entry.Size = size.UintValue()
		}
		if src := sf.ChildByName(TagPartfileSourceCount); src != nil {
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

	searchTag := Tag{Name: TagSearchfile, Type: TagTypeHash16, Data: h}
	searchTag.Children = []Tag{newUint32Tag(TagPartfileCat, uint32(categoryIdx))}
	_, err := c.conn.SendRequest(&Packet{
		OpCode: OpMiscData,
		Tags: []Tag{
			searchTag,
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
		OpCode: OpDisconnect,
	})
	return err
}

func (c *Client) ConnectED2K() error {
	_, err := c.conn.SendRequest(&Packet{
		OpCode: OpConnect,
	})
	return err
}

func (c *Client) AddServer(address, name string) error {
	tags := []Tag{newStringTag(TagServerAddress, address)}
	if name != "" {
		tags = append(tags, newStringTag(TagServerName, name))
	}
	_, err := c.conn.SendRequest(&Packet{
		OpCode: OpServerAdd,
		Tags:   tags,
	})
	return err
}

func (c *Client) ConnectToServer(address string) error {
	_, err := c.conn.SendRequest(&Packet{
		OpCode: OpServerConnect,
		Tags: []Tag{
			{Name: TagServer, Type: TagTypeIPV4, Data: address},
		},
	})
	return err
}

func (c *Client) RemoveServer(address string) error {
	_, err := c.conn.SendRequest(&Packet{
		OpCode: OpServerRemove,
		Tags: []Tag{
			{Name: TagServer, Type: TagTypeIPV4, Data: address},
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
