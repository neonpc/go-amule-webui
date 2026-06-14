package ec

type TagName uint16

const (
	TagString          TagName = 0x0000
	TagPasswdHash      TagName = 0x0001
	TagProtocolVersion TagName = 0x0002
	TagDetailLevel     TagName = 0x0004
	TagPasswdSalt      TagName = 0x000B
	TagCanZLib         TagName = 0x000C
	TagCanUTF8Numbers  TagName = 0x000D
	TagCanLargeTagCount TagName = 0x0011

	TagConnState TagName = 0x0005
	TagEd2kID    TagName = 0x0006
	TagClientID  TagName = 0x000A
	TagKadID     TagName = 0x0010

	TagClientName    TagName = 0x0100
	TagClientVersion TagName = 0x0101

	TagStatsULSpeed      TagName = 0x0200
	TagStatsDLSpeed      TagName = 0x0201
	TagStatsULData       TagName = 0x0202
	TagStatsDLData       TagName = 0x0203
	TagStatsUpOverhead   TagName = 0x0204
	TagStatsDownOverhead TagName = 0x0205

	TagPartfile       TagName = 0x0300
	TagPartfileName   TagName = 0x0301
	TagPartfileSize   TagName = 0x0303
	TagPartfileDone   TagName = 0x0304
	TagPartfileSizeDone TagName = 0x0306
	TagPartfileSpeed  TagName = 0x0307
	TagPartfileHash   TagName = 0x031E
	TagPartfileStatus TagName = 0x0315
	TagPartfilePrio   TagName = 0x030F
	TagPartfileSources TagName = 0x030C
	TagPartfileCat    TagName = 0x0318
	TagPartfilePaused TagName = 0x031A
	TagPartfileSourceCount TagName = 0x030A
	TagPartfilePriority TagName = 0x0309
	TagPartfileStopped TagName = 0x0317

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
	TagServerVersion  TagName = 0x050B
	TagServerIP       TagName = 0x050C
	TagServerPort     TagName = 0x050D

	TagClient         TagName = 0x0600
	TagClientNameF    TagName = 0x0601
	TagClientUploaded TagName = 0x0602
	TagClientSpeed    TagName = 0x0603

	TagSearchfile       TagName = 0x0700
	TagSearchType       TagName = 0x0701
	TagSearchName       TagName = 0x0702
	TagSearchMinSize    TagName = 0x0703
	TagSearchMaxSize    TagName = 0x0704
	TagSearchFileType   TagName = 0x0705
	TagSearchExtension  TagName = 0x0706
	TagSearchAvail      TagName = 0x0707

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
	TagPasswdSalt:      "PasswdSalt",
	TagCanZLib:         "CanZLib",
	TagCanUTF8Numbers:  "CanUTF8Numbers",
	TagCanLargeTagCount: "CanLargeTagCount",
	TagConnState:       "ConnState",
	TagEd2kID:          "Ed2kID",
	TagClientID:        "ClientID",
	TagKadID:           "KadID",
	TagClientName:      "ClientName",
	TagClientVersion:   "ClientVersion",
	TagStatsULSpeed:    "StatsULSpeed",
	TagStatsDLSpeed:    "StatsDLSpeed",
	TagStatsULData:     "StatsULData",
	TagStatsDLData:     "StatsDLData",
	TagStatsUpOverhead:   "StatsUpOverhead",
	TagStatsDownOverhead: "StatsDownOverhead",
	TagPartfile:        "Partfile",
	TagPartfileName:    "PartfileName",
	TagPartfileSize:    "PartfileSize",
	TagPartfileDone:     "PartfileDone",
	TagPartfileSizeDone: "PartfileSizeDone",
	TagPartfileSpeed:    "PartfileSpeed",
	TagPartfileHash:     "PartfileHash",
	TagPartfileStatus:   "PartfileStatus",
	TagPartfilePrio:     "PartfilePrio",
	TagPartfileSources:  "PartfileSources",
	TagPartfileSourceCount: "PartfileSourceCount",
	TagPartfilePriority: "PartfilePriority",
	TagPartfileCat:      "PartfileCat",
	TagPartfilePaused:   "PartfilePaused",
	TagPartfileStopped:  "PartfileStopped",
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
	TagServerVersion:   "ServerVersion",
	TagServerIP:        "ServerIP",
	TagServerPort:      "ServerPort",
	TagClient:          "Client",
	TagClientNameF:     "ClientNameF",
	TagClientUploaded:  "ClientUploaded",
	TagClientSpeed:     "ClientSpeed",
	TagSearchfile:       "Searchfile",
	TagSearchType:       "SearchType",
	TagSearchName:       "SearchName",
	TagSearchMinSize:    "SearchMinSize",
	TagSearchMaxSize:    "SearchMaxSize",
	TagSearchFileType:   "SearchFileType",
	TagSearchExtension:  "SearchExtension",
	TagSearchAvail:      "SearchAvail",
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
