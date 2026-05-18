package statuspage

import "time"

type StatusPage struct {
	ID                    string
	Slug                  string
	Title                 string
	Description           string
	Icon                  string
	Theme                 string
	Published             bool
	ShowTags              bool
	ShowPoweredBy         bool
	ShowCertificateExpiry bool
	CustomCSS             string
	FooterText            string
	GoogleAnalyticsTagID  string
}

type Group struct {
	ID           string
	Name         string
	Weight       int64
	StatusPageID string
	TagIDs       string
}

type MonitorGroup struct {
	MonitorID string
	GroupID   string
	Weight    int64
}

type Incident struct {
	ID              string
	Title           string
	Content         string
	Style           string
	Pin             bool
	Active          bool
	StatusPageID    string
	CreatedDate     time.Time
	LastUpdatedDate *time.Time
}
