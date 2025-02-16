package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-tg-support-ticket/form"
	"go-tg-support-ticket/logger"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Config holds the webhook settings
type Config struct {
	Enabled      bool   `mapstructure:"enabled"`
	URL          string `mapstructure:"url"`
	Auth         Auth   `mapstructure:"auth"`
	WorkersCount int    `mapstructure:"workers_count"`
	QueueSize    int    `mapstructure:"queue_size"`
}

type Auth struct {
	Type     string `yaml:"type"`     // "none", "bearer", "basic"
	Token    string `yaml:"token"`    // Bearer token
	Username string `yaml:"username"` // Basic auth username
	Password string `yaml:"password"` // Basic auth password
}

// Event represents the event data to be sent
type Event struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// worker manages concurrent webhook requests
type worker struct {
	cfg    *Config
	queue  chan Event
	wg     sync.WaitGroup
	client *http.Client
}

type WorkerInterface interface {
	Enqueue(form *form.Form)
}

var Workers WorkerInterface

// NewWebhookWorker initializes a worker pool
func NewWebhookWorker(cfg *Config) {
	w := worker{
		cfg: cfg,
	}
	if cfg.Enabled {
		w.queue = make(chan Event, cfg.QueueSize)
		w.client = &http.Client{Timeout: 10 * time.Second}

		// Start workers
		for i := 0; i < cfg.WorkersCount; i++ {
			w.wg.Add(1)
			go w.processQueue()
		}
		Workers = &w
	}
}

// Enqueue adds a webhook request to the queue
func (w *worker) Enqueue(form *form.Form) {
	if w != nil {
		event := buildEvent(form)
		w.queue <- event
	}
}

func buildEvent(form *form.Form) Event {
	data := make(map[string]interface{})
	for _, field := range form.Fields {
		data[field.Name] = field.UserValue
	}
	return Event{Event: form.FormName, Data: data}
}

// processQueue processes webhook requests in background workers
func (w *worker) processQueue() {
	defer w.wg.Done()
	for req := range w.queue {
		err := w.SendWebhook(req)
		if err != nil {
			logger.PrintLog(0, "failed to send webhook", err)
		}
	}
}

// Shutdown waits for all workers to finish
func (w *worker) Shutdown() {
	close(w.queue)
	w.wg.Wait()
}

// SendWebhook sends event data if webhook is enabled
func (w *worker) SendWebhook(e Event) error {
	// Check if webhook is enabled
	if !w.cfg.Enabled || w.cfg.URL == "" {
		return nil // Webhook is disabled, do nothing
	}

	payload, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook data: %w", err)
	}

	req, err := http.NewRequest("POST", w.cfg.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Handle authentication if enabled
	if strings.ToLower(w.cfg.Auth.Type) == "bearer" {
		req.Header.Set("Authorization", "Bearer "+w.cfg.Auth.Token)
	} else if strings.ToLower(w.cfg.Auth.Type) == "basic" {
		req.SetBasicAuth(w.cfg.Auth.Username, w.cfg.Auth.Password)
	}

	// Send request using worker's HTTP client
	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook failed: status=%d, response=%s", resp.StatusCode, body)
	}

	return nil
}
