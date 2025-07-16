package main

import "time"

type ErrorReport struct {
	ID             string    `json:"id,omitempty"`
	Symptom        string    `json:"symptom"`
	Date           time.Time `json:"date"`
	Program        string    `json:"program"`
	ProgramVersion string    `json:"program_version"`
	Distro         string    `json:"distro"`
	DistroVersion  string    `json:"distro_version"`
	Resources      []string  `json:"resources"`
	Solution       string    `json:"solution"`
}

type Filter struct {
	Q              string     `json:"q,omitempty"`               // General search query
	Symptom        string     `json:"symptom,omitempty"`         // Filter by symptom
	Program        string     `json:"program,omitempty"`         // Filter by program
	ProgramVersion string     `json:"program_version,omitempty"` // Filter by program version
	Distro         string     `json:"distro,omitempty"`          // Filter by distro
	DistroVersion  string     `json:"distro_version,omitempty"`  // Filter by distro version
	DateFrom       *time.Time `json:"date_from,omitempty"`       // Filter by date range (from)
	DateTo         *time.Time `json:"date_to,omitempty"`         // Filter by date range (to)
	ResourcesAny   []string   `json:"resources_any,omitempty"`   // Filter by any of these resources
	Solution       string     `json:"solution,omitempty"`        // Filter by solution text
}
