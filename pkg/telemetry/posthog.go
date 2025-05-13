package telemetry

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/posthog/posthog-go"
	"github.com/saasuke-labs/gengo/pkg/version"
)

const (
	envDisableKey = "GENGO_TELEMETRY"                                 // GENGO_TELEMETRY=off disables tracking
	projectKey    = "phc_7mEPoPEGtg6BbqetPhCiMZb0W762T5Eo0quW31Nu2El" // <-- replace with your PostHog project token
	cacheFileName = "00000000-0000-0000-0000-000000000000"            // UUID file name
)

var client posthog.Client
var distinctID string

func init() {
	if os.Getenv(envDisableKey) == "off" {
		log.Println("[telemetry] disabled via env var")
		return
	}

	var err error
	client, err = posthog.NewWithConfig(projectKey, posthog.Config{
		Endpoint:  "https://eu.i.posthog.com",
		BatchSize: 5,
		Interval:  5 * time.Second,
	})
	if err != nil {
		log.Println("[telemetry] failed to init:", err)
		return
	}

	distinctID, _ = getOrCreateDistinctID()
}

// getOrCreateDistinctID stores a persistent UUID in the user's config dir
func getOrCreateDistinctID() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	idPath := filepath.Join(cfgDir, "gengo", cacheFileName)

	// Read existing ID
	if data, err := os.ReadFile(idPath); err == nil {
		return string(bytes.TrimSpace(data)), nil
	}

	// Create new ID
	id := newUUID()
	if err := os.MkdirAll(filepath.Dir(idPath), 0755); err != nil {
		return id, err
	}
	if err := os.WriteFile(idPath, []byte(id), 0644); err != nil {
		return id, err
	}
	return id, nil
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Track sends a CLI event
func Track(event string, props map[string]interface{}) {
	if client == nil {
		return
	}
	// Add standard metadata
	props["os"] = runtime.GOOS
	props["arch"] = runtime.GOARCH
	props["go_version"] = runtime.Version()
	props["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	props["version"] = version.Version
	props["commit"] = version.Commit
	props["build_date"] = version.Date

	client.Enqueue(posthog.Capture{
		DistinctId: distinctID,
		Event:      event,

		Properties: props,
	})
}

// Close flushes any pending events (defer this in main)
func Close() {
	if client != nil {
		client.Close()
	}
}
