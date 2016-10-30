package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
)

// Plugin is the internal data structure
type Plugin struct {
	reporter *Reporter
}

func main() {
	const socket = "/var/run/scope/plugins/scope-debugger/scope-debugger.sock"

	// Handle the exit signal
	setupSignals(socket)

	listener, err := setupSocket(socket)
	if err != nil {
		log.Fatalf("Failed to setup socket: %v", err)
	}

	plugin, err := NewPlugin()
	if err != nil {
		log.Fatalf("Failed to create a plugin: %v", err)
	}


	debuggerServeMux := http.NewServeMux()

	// Report request handler
	reportHandler := http.HandlerFunc(plugin.report)
	debuggerServeMux.Handle("/report", reportHandler)

	// Control request handler
	controlHandler := http.HandlerFunc(plugin.control)
	debuggerServeMux.Handle("/control", controlHandler)

	log.Println("Listening...")
	if err = http.Serve(listener, debuggerServeMux); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func setupSignals(socket string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		os.Remove(socket)
		os.Exit(0)
	}()
}

func setupSocket(socket string) (net.Listener, error) {
	os.Remove(socket)
	if err := os.MkdirAll(filepath.Dir(socket), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %q: %v", filepath.Dir(socket), err)
	}
	listener, err := net.Listen("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %q: %v", socket, err)
	}

	log.Printf("Listening on: unix://%s", socket)
	return listener, nil
}

// NewPlugin instantiates a new plugin
func NewPlugin() (*Plugin, error) {
	reporter := NewReporter()
	plugin := &Plugin{
		reporter: reporter,
	}
	return plugin, nil
}

func (p *Plugin) report(w http.ResponseWriter, r *http.Request) {
	raw, err := p.reporter.RawReport()
	if err != nil {
		msg := fmt.Sprintf("error: failed to get raw report: %v", err)
		log.Print(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

type response struct {
	Error string `json:"error,omitempty"`
}

func (p *Plugin) control(w http.ResponseWriter, r *http.Request) {
	sendResponse(w, fmt.Errorf("Not implemented"))
}

func sendResponse(w http.ResponseWriter, err error) {
	res := response{}
	if err != nil {
		res.Error = err.Error()
	}
	raw, err := json.Marshal(res)
	if err != nil {
		log.Printf("Internal server error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}
