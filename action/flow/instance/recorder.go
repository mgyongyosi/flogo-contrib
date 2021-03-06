package instance

import (
	"strings"
	"encoding/json"
	"net/http"
	"bytes"

	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/TIBCOSoftware/flogo-lib/flow/service"
	"github.com/TIBCOSoftware/flogo-lib/util"
)

// StateRecorder is the interface that describes a service that can record
// snapshots and steps of a Flow Instance
type StateRecorder interface {
	// RecordSnapshot records a Snapshot of the FlowInstance
	RecordSnapshot(instance *Instance)

	// RecordStep records the changes for the current Step of the Flow Instance
	RecordStep(instance *Instance)
}

// RemoteStateRecorder is an implementation of StateRecorder service
// that can access flows via URI
type RemoteStateRecorder struct {
	host    string
	enabled bool
}

// NewRemoteStateRecorder creates a new RemoteStateRecorder
func NewRemoteStateRecorder(config *util.ServiceConfig) *RemoteStateRecorder {

	recorder := &RemoteStateRecorder{enabled: config.Enabled}
	recorder.init(config.Settings)

	return recorder
}

func (sr *RemoteStateRecorder) Name() string {
	return service.ServiceStateRecorder
}

func (sr *RemoteStateRecorder) Enabled() bool {
	return sr.enabled
}

// Start implements util.Managed.Start()
func (sr *RemoteStateRecorder) Start() error {
	// no-op
	return nil
}

// Stop implements util.Managed.Stop()
func (sr *RemoteStateRecorder) Stop() error {
	// no-op
	return nil
}

// Init implements services.StateRecorderService.Init()
func (sr *RemoteStateRecorder) init(settings map[string]string) {

	host, set := settings["host"]
	port, set := settings["port"]

	if !set {
		panic("RemoteStateRecorder: requried setting 'host' not set")
	}

	if strings.Index(host, "http") != 0 {
		sr.host = "http://" + host + ":" + port
	} else {
		sr.host = host + ":" + port
	}

	logger.Debugf("RemoteStateRecorder: StateRecoder Server = %s", sr.host)
}

// RecordSnapshot implements instance.StateRecorder.RecordSnapshot
func (sr *RemoteStateRecorder) RecordSnapshot(instance *Instance) {

	storeReq := &RecordSnapshotReq{
		ID:           instance.StepID(),
		FlowID:       instance.ID(),
		State:        instance.State(),
		Status:       int(instance.Status()),
		SnapshotData: instance,
	}

	uri := sr.host + "/instances/snapshot"

	logger.Debugf("POST Snapshot: %s\n", uri)

	jsonReq, _ := json.Marshal(storeReq)

	logger.Debug("JSON: ", string(jsonReq))

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	logger.Debug("response Status:", resp.Status)

	if resp.StatusCode >= 300 {
		//error
	}
}

// RecordStep implements instance.StateRecorder.RecordStep
func (sr *RemoteStateRecorder) RecordStep(instance *Instance) {

	storeReq := &RecordStepReq{
		ID:       instance.StepID(),
		FlowID:   instance.ID(),
		State:    instance.State(),
		Status:   int(instance.Status()),
		StepData: instance.ChangeTracker,
	}

	uri := sr.host + "/instances/steps"

	logger.Debugf("POST Snapshot: %s\n", uri)

	jsonReq, _ := json.Marshal(storeReq)

	logger.Debug("JSON: ", string(jsonReq))

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	logger.Debug("response Status:", resp.Status)

	if resp.StatusCode >= 300 {
		//error
	}
}

// RecordSnapshotReq serializable representation of the RecordSnapshot request
type RecordSnapshotReq struct {
	ID     int    `json:"id"`
	FlowID string `json:"flowID"`
	State  int    `json:"state"`
	Status int    `json:"status"`

	SnapshotData *Instance `json:"snapshotData"`
}

// RecordStepReq serializable representation of the RecordStep request
type RecordStepReq struct {
	ID     int    `json:"id"`
	FlowID string `json:"flowID"`
	State  int    `json:"state"`
	Status int    `json:"status"`

	StepData *InstanceChangeTracker `json:"stepData"`
}

func DefaultConfig() *util.ServiceConfig {
	return &util.ServiceConfig{Name: service.ServiceStateRecorder, Enabled: true, Settings: map[string]string{"host": ""}}
}
