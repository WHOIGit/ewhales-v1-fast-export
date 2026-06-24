package models

// Logbook represents a logbook entity.
// In the EAV model, it is identified when a post_id has a 'logbook_id'
// meta_value that contains text (like a name or year range).
type Logbook struct {
	PostID           uint   `json:"post_id"`
	LogbookID        string `json:"logbook_id"`
	Researcher       string `json:"researcher"`
	Repository       string `json:"repository"`
	Vessel           string `json:"vessel"`
	Flag             string `json:"flag"`
	HomePort         string `json:"home_port"`
	Master           string `json:"master"`
	HomeMeridian     string `json:"home_meridian"`
	ServiceType      string `json:"service_type"`
	MetInstruments   string `json:"met_instruments"`
	ShermanCode      string `json:"sherman_code"`
	RepositoryCallNo string `json:"repository_call_no"`
}

// LogbookEntry represents an entry within a logbook.
// In the EAV model, it is identified when a post_id has a 'logbook_id'
// meta_value that is an unsigned integer, which links back to the Logbook's PostID.
type LogbookEntry struct {
	PostID        uint   `json:"post_id"`
	LogbookID     uint   `json:"logbook_id"`
	Bottom        string `json:"bottom"`
	CloudCover    string `json:"cloud_cover"`
	Depth         string `json:"depth"`
	DepthUnit     string `json:"depth_unit"`
	EntryDate     string `json:"entry_date"`
	Landmark      string `json:"landmark"`
	Latitude      string `json:"latitude"`
	LocalTime     string `json:"local_time"`
	Longitude     string `json:"longitude"`
	Page          string `json:"page"`
	SeaState      string `json:"sea_state"`
	ShipHeading   string `json:"ship_heading"`
	ShipSightings string `json:"ship_sightings"`
	Weather	      string `json:"weather"`
	WindDirection string `json:"wind_direction"`
	WindForce     string `json:"wind_force"`
}

// PivotData acts as the central in-memory structure holding all parsed entities.
// It is the resulting data structure after querying the EAV database and parsing the properties.
type PivotData struct {
	Logbooks       []Logbook
	LogbookEntries []LogbookEntry
}
