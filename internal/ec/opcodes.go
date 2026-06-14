package ec

type OpCode uint8

const (
	OpNoop           OpCode = 0x01
	OpAuthReq        OpCode = 0x02
	OpAuthFail       OpCode = 0x03
	OpAuthOK         OpCode = 0x04
	OpFailed         OpCode = 0x05
	OpStrings        OpCode = 0x06
	OpMiscData       OpCode = 0x07
	OpShutdown       OpCode = 0x08
	OpAddLink        OpCode = 0x09
	OpStatReq        OpCode = 0x0A
	OpGetConnState   OpCode = 0x0B
	OpStats          OpCode = 0x0C
	OpGetDloadQueue  OpCode = 0x0D
	OpGetUloadQueue  OpCode = 0x0E
	OpGetSharedFiles OpCode = 0x10
	OpSharedSetPrio  OpCode = 0x11
	OpDloadQueue     OpCode = 0x1F
	OpUloadQueue     OpCode = 0x20
	OpSharedFiles    OpCode = 0x22
	OpSearchStart    OpCode = 0x26
	OpSearchStop     OpCode = 0x27
	OpSearchResults  OpCode = 0x28
	OpGetServerList  OpCode = 0x2C
	OpServerList     OpCode = 0x2D
	OpAddLogLine     OpCode = 0x33
	OpAddDebugLogLine OpCode = 0x34
	OpGetLog         OpCode = 0x35
	OpGetDebugLog    OpCode = 0x36
	OpLog            OpCode = 0x38
	OpDebugLog       OpCode = 0x39
	OpGetPreferences OpCode = 0x3F
	OpSetPreferences OpCode = 0x40
	OpGetStatsTree   OpCode = 0x46
	OpStatsTree      OpCode = 0x47
	OpKadStart        OpCode = 0x48
	OpKadStop         OpCode = 0x49
	OpConnect         OpCode = 0x4A
	OpDisconnect      OpCode = 0x4B
	OpAuthSalt        OpCode = 0x4F
	OpAuthPasswd      OpCode = 0x50
	OpFriend          OpCode = 0x57
	OpServerDisconnect OpCode = 0x2E
	OpServerConnect   OpCode = 0x2F
	OpServerRemove    OpCode = 0x30
	OpServerAdd       OpCode = 0x31
)

var opcodeNames = map[OpCode]string{
	OpNoop:           "Noop",
	OpAuthReq:        "AuthReq",
	OpAuthFail:       "AuthFail",
	OpAuthOK:         "AuthOK",
	OpFailed:         "Failed",
	OpStrings:        "Strings",
	OpMiscData:       "MiscData",
	OpShutdown:       "Shutdown",
	OpAddLink:        "AddLink",
	OpStatReq:        "StatReq",
	OpGetConnState:   "GetConnState",
	OpStats:          "Stats",
	OpGetDloadQueue:  "GetDloadQueue",
	OpGetUloadQueue:  "GetUloadQueue",
	OpGetSharedFiles: "GetSharedFiles",
	OpSharedSetPrio:  "SharedSetPrio",
	OpDloadQueue:     "DloadQueue",
	OpUloadQueue:     "UloadQueue",
	OpSharedFiles:    "SharedFiles",
	OpSearchStart:    "SearchStart",
	OpSearchStop:     "SearchStop",
	OpSearchResults:  "SearchResults",
	OpGetServerList:  "GetServerList",
	OpServerList:     "ServerList",
	OpAddLogLine:     "AddLogLine",
	OpAddDebugLogLine: "AddDebugLogLine",
	OpGetLog:         "GetLog",
	OpGetDebugLog:    "GetDebugLog",
	OpLog:            "Log",
	OpDebugLog:       "DebugLog",
	OpGetPreferences: "GetPreferences",
	OpSetPreferences: "SetPreferences",
	OpGetStatsTree:   "GetStatsTree",
	OpStatsTree:      "StatsTree",
	OpKadStart:        "KadStart",
	OpKadStop:         "KadStop",
	OpConnect:         "Connect",
	OpDisconnect:      "Disconnect",
	OpAuthSalt:        "AuthSalt",
	OpAuthPasswd:      "AuthPasswd",
	OpFriend:          "Friend",
	OpServerDisconnect: "ServerDisconnect",
	OpServerConnect:   "ServerConnect",
	OpServerRemove:    "ServerRemove",
	OpServerAdd:       "ServerAdd",
}

func (o OpCode) String() string {
	if s, ok := opcodeNames[o]; ok {
		return s
	}
	return "Unknown"
}
