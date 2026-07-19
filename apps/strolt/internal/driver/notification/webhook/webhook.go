// Package webhook implements a notification driver that posts the raw operation
// context as JSON to a configured HTTP endpoint.
package webhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/driver/notification/notifyutil"
	"github.com/strolt/strolt/shared/logger"

	"gopkg.in/yaml.v3"
)

// requestTimeout bounds a single webhook delivery so a slow or unresponsive
// endpoint cannot block the task pipeline indefinitely.
const requestTimeout = 30 * time.Second

// Config holds the target webhook URL.
type Config struct {
	URL string `yaml:"url"`
}

// Webhook is a notification driver that posts the operation context to a URL.
type Webhook struct {
	logger *logger.Logger
	config Config
	client *http.Client
}

// New creates a Webhook driver.
func New() *Webhook {
	return &Webhook{
		client: &http.Client{Timeout: requestTimeout},
	}
}

// SetLogger sets the logger used by the driver.
func (i *Webhook) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

// SetConfig parses and validates the driver configuration.
func (i *Webhook) SetConfig(config any) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := yaml.Unmarshal(data, &i.config); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	return validateConfig(i.config)
}

// Send posts the operation context as JSON to the configured webhook URL.
func (i *Webhook) Send(ctx context.Context) {
	data, err := json.Marshal(ctx)
	if err != nil {
		i.logger.Error(err)
		return
	}

	resp, err := i.client.Post(i.config.URL, "application/json", bytes.NewReader(data)) //nolint:noctx
	if err != nil {
		// The transport error embeds the request URL, which may carry a secret
		// token; redact it before logging.
		i.logger.Error(notifyutil.RedactURLError(err))
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	b, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= http.StatusBadRequest {
		i.logger.WithField("status", resp.Status).Error(string(b))
	}
}

func validateConfig(config Config) error {
	if config.URL == "" {
		return errors.New("url is empty")
	}

	u, err := url.Parse(config.URL)
	if err != nil {
		return fmt.Errorf("url is invalid: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("url must use the http or https scheme")
	}

	if u.Host == "" {
		return errors.New("url has no host")
	}

	return nil
}
