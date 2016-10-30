package main

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	debuggerTablePrefix = "debugger-"
)

type report struct {
	Process   topology
	Container topology
	Plugins   []pluginSpec
}

type topology struct {
	Nodes             map[string]node             `json:"nodes"`
	Controls          map[string]control          `json:"controls"`
	MetadataTemplates map[string]metadataTemplate `json:"metadata_templates,omitempty"`
	TableTemplates    map[string]tableTemplate    `json:"table_templates,omitempty"`
}

type tableTemplate struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Prefix string `json:"prefix"`
}

type metadataTemplate struct {
	ID       string  `json:"id"`
	Label    string  `json:"label,omitempty"`    // Human-readable descriptor for this row
	Truncate int     `json:"truncate,omitempty"` // If > 0, truncate the value to this length.
	Datatype string  `json:"dataType,omitempty"`
	Priority float64 `json:"priority,omitempty"`
	From     string  `json:"from,omitempty"` // Defines how to get the value from a report node
}

type node struct {
	LatestControls map[string]controlEntry `json:"latestControls,omitempty"`
	Latest         map[string]stringEntry  `json:"latest,omitempty"`
}

type controlEntry struct {
	Timestamp time.Time   `json:"timestamp"`
	Value     controlData `json:"value"`
}

type controlData struct {
	Dead bool `json:"dead"`
}

type control struct {
	ID    string `json:"id"`
	Human string `json:"human"`
	Icon  string `json:"icon"`
	Rank  int    `json:"rank"`

	AlwaysPropagated bool   `json:"always_propagated,omitempty"`
	StartImage       string `json:"start_image,omitempty"`
}

type stringEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Value     string    `json:"value"`
}

type pluginSpec struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Interfaces  []string `json:"interfaces"`
	APIVersion  string   `json:"api_version,omitempty"`
}

// Reporter internal data structure
type Reporter struct {
}

// NewReporter instantiates a new Reporter
func NewReporter() *Reporter {
	return &Reporter{}
}

// RawReport returns a report
func (r *Reporter) RawReport() ([]byte, error) {
	rpt := &report{
		Process: topology{
			Controls: getDebuggerControls(),
		},
		Container: topology{
			Controls: getDebuggerControls(),
		},
		Plugins: []pluginSpec{
			{
				ID:          "scope-debugger",
				Label:       "Debugger",
				Description: "Add buttons to run debugger tools: GDB, strace, delve",
				Interfaces:  []string{"reporter", "controller"},
				APIVersion:  "1",
			},
		},
	}
	raw, err := json.Marshal(rpt)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the report: %v", err)
	}
	return raw, nil
}

func getDebuggerControls() map[string]control {
	controls := map[string]control{}
	for _, c := range getControls() {
		controls[c.control.ID] = c.control
	}
	return controls
}

type extControl struct {
	control control
	handler func(pid int) error
}

func getControls() []extControl {
	return []extControl{
		{
			control: control{
				ID:    fmt.Sprintf("%s%s", debuggerTablePrefix, "gdb"),
				Human: "GDB",
				Icon:  "fa-bug",
				Rank:  24,

				AlwaysPropagated: true,
				StartImage:       "albanc/toolbox",
			},
			handler: func(pid int) error {
				return nil
			},
		},
	}
}
