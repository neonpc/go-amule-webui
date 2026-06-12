package ec

type TagName uint16

const (
	TagString          TagName = 0x0000
	TagPasswdHash      TagName = 0x0001
	TagProtocolVersion TagName = 0x0002
	TagDetailLevel     TagName = 0x0004
	TagCanZLib         TagName = 0x000C
	TagCanUTF8Numbers  TagName = 0x000D
	TagClientName      TagName = 0x0100
	TagClientVersion   TagName = 0x0101
	TagServerVersion   TagName = 0x0102

	TagStatsULSpeed  TagName = 0x0200
	TagStatsDLSpeed  TagName = 0x0201
	TagStatsULData   TagName = 0x0202
	TagStatsDLData   TagName = 0x0203
	TagStatsConnState TagName = 0x0204
	TagStatsKadState TagName = 0x0205

	TagPartfile       TagName = 0x0300
	TagPartfileName   TagName = 0x0301
	TagPartfileSize   TagName = 0x0303
	TagPartfileDone   TagName = 0x0304
	TagPartfileSpeed  TagName = 0x0307
	TagPartfileHash   TagName = 0x031E
	TagPartfileStatus TagName = 0x0315
	TagPartfilePrio   TagName = 0x030F
	TagPartfileSources TagName = 0x030C
	TagPartfileCat    TagName = 0x0318
	TagPartfilePaused TagName = 0x031A

	TagKnownfile      TagName = 0x0400
	TagKnownfileName  TagName = 0x0401
	TagKnownfileSize  TagName = 0x0402
	TagKnownfileHash  TagName = 0x0403

	TagServer         TagName = 0x0500
	TagServerName     TagName = 0x0501
	TagServerDesc     TagName = 0x0502
	TagServerAddress  TagName = 0x0503
	TagServerUsers    TagName = 0x0504
	TagServerFiles    TagName = 0x0505

	TagClient         TagName = 0x0600
	TagClientNameF    TagName = 0x0601
	TagClientUploaded TagName = 0x0602
	TagClientSpeed    TagName = 0x0603

	TagSearchfile     TagName = 0x0700
	TagSearchfileName TagName = 0x0701
	TagSearchfileSize TagName = 0x0702
	TagSearchfileHash TagName = 0x0703
	TagSearchfileSources TagName = 0x0704

	TagSelectPrefs      TagName = 0x1000
	TagPrefsConns       TagName = 0x1300
	TagPrefsFiles       TagName = 0x1400
	TagPrefsWebserver   TagName = 0x1C00

	TagStatsTreeNode    TagName = 0x1B06
	TagStatsNodeValue   TagName = 0x1B07
)

var tagNames = map[TagName]string{
	TagString:          "String",
	TagPasswdHash:      "PasswdHash",
	TagProtocolVersion: "ProtocolVersion",
	TagDetailLevel:     "DetailLevel",
	TagCanZLib:         "CanZLib",
	TagCanUTF8Numbers:  "CanUTF8Numbers",
	TagClientName:      "ClientName",
	TagClientVersion:   "ClientVersion",
	TagServerVersion:   "ServerVersion",
	TagStatsULSpeed:    "StatsULSpeed",
	TagStatsDLSpeed:    "StatsDLSpeed",
	TagStatsULData:     "StatsULData",
	TagStatsDLData:     "StatsDLData",
	TagStatsConnState:  "StatsConnState",
	TagStatsKadState:   "StatsKadState",
	TagPartfile:        "Partfile",
	TagPartfileName:    "PartfileName",
	TagPartfileSize:    "PartfileSize",
	TagPartfileDone:    "PartfileDone",
	TagPartfileSpeed:   "PartfileSpeed",
	TagPartfileHash:    "PartfileHash",
	TagPartfileStatus:  "PartfileStatus",
	TagPartfilePrio:    "PartfilePrio",
	TagPartfileSources: "PartfileSources",
	TagPartfileCat:     "PartfileCat",
	TagPartfilePaused:  "PartfilePaused",
	TagKnownfile:       "Knownfile",
	TagKnownfileName:   "KnownfileName",
	TagKnownfileSize:   "KnownfileSize",
	TagKnownfileHash:   "KnownfileHash",
	TagServer:          "Server",
	TagServerName:      "ServerName",
	TagServerDesc:      "ServerDesc",
	TagServerAddress:   "ServerAddress",
	TagServerUsers:     "ServerUsers",
	TagServerFiles:     "ServerFiles",
	TagClient:          "Client",
	TagClientNameF:     "ClientNameF",
	TagClientUploaded:  "ClientUploaded",
	TagClientSpeed:     "ClientSpeed",
	TagSearchfile:      "Searchfile",
	TagSearchfileName:  "SearchfileName",
	TagSearchfileSize:  "SearchfileSize",
	TagSearchfileHash:  "SearchfileHash",
	TagSearchfileSources: "SearchfileSources",
	TagSelectPrefs:     "SelectPrefs",
	TagPrefsConns:      "PrefsConns",
	TagPrefsFiles:      "PrefsFiles",
	TagPrefsWebserver:  "PrefsWebserver",
	TagStatsTreeNode:   "StatsTreeNode",
	TagStatsNodeValue:  "StatsNodeValue",
}

func (t TagName) String() string {
	if s, ok := tagNames[t]; ok {
		return s
	}
	return "Unknown"
}
