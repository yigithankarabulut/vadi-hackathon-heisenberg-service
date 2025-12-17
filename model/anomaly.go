package model

// AnomalyType represents the type of anomaly detected
type AnomalyType string

const (
	AnomalyTypeThreshold AnomalyType = "threshold"
	AnomalyTypeGeofence  AnomalyType = "geofence"
	AnomalyTypeBoth      AnomalyType = "both"
)

// Anomaly represents detected anomaly information
type Anomaly struct {
	HasAnomaly  bool        `json:"has_anomaly"`
	AnomalyType AnomalyType `json:"anomaly_type,omitempty"`
	Details     string      `json:"details,omitempty"`
}
