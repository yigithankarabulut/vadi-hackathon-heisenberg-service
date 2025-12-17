package model

// TelemetryDTO defines the structure for incoming telemetry data from ingestion service
type TelemetryDTO struct {
	Timestamp   uint64  `json:"timestamp"`  // time of the telemetry data
	PlaneID     string  `json:"planeId"`    // unique identifier for the plane (MAC address)
	Latitude    float64 `json:"lat"`        // latitude of the plane
	Longitude   float64 `json:"lon"`        // longitude of the plane
	Altitude    float64 `json:"alt_baro"`   // barometric altitude
	GroundSpeed float64 `json:"gs"`         // ground speed
	Heading     float64 `json:"heading"`    // heading of the plane
	ClimbRate   float64 `json:"climb_rate"` // climb rate of the plane
}

