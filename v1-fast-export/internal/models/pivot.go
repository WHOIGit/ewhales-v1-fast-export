package models

import pb "github.com/agl/ewhales-v1-exporter/proto"

// PivotData acts as the central in-memory structure holding all parsed entities.
// It is the resulting data structure after querying the EAV database and parsing the properties.
type PivotData struct {
	Logbooks       []*pb.Logbook
	LogbookEntries []*pb.LogbookEntry
}
