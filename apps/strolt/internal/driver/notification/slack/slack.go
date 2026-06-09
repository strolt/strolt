// Package slack implements a notification driver that posts messages to a Slack webhook.
package slack

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/shared/logger"

	"gopkg.in/yaml.v3"
)

// Config holds the Slack webhook identifiers.
type Config struct {
	TeamID string `yaml:"teamId"`
	BotID  string `yaml:"botId"`
	HookID string `yaml:"hookId"`
}

// Params holds optional driver parameters.
type Params struct {
	apiPrefix string // changed only in tests
}

// Slack is a notification driver that sends messages to a Slack webhook.
type Slack struct {
	Params

	logger *logger.Logger
	config Config
}

const slackAPIPrefix = "https://hooks.slack.com/services"

// New creates a Slack driver with the given parameters.
func New(params Params) *Slack {
	res := Slack{
		Params: params,
	}

	if res.apiPrefix == "" {
		res.apiPrefix = slackAPIPrefix
	}

	return &res
}

// SetConfig parses and validates the driver configuration.
func (i *Slack) SetConfig(config any) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := yaml.Unmarshal(data, &i.config); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	return validateConfig(i.config)
}

// SetLogger sets the logger used by the driver.
func (i *Slack) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

// Send posts a notification message built from the context to the Slack webhook.
func (i *Slack) Send(ctx context.Context) {
	body, err := getTemplate(ctx)
	if err != nil {
		i.logger.Error(err)
		return
	}

	resp, err := http.Post(i.getWebhook(), "application/json", body) //nolint:noctx
	if err != nil {
		i.logger.Error(err)
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

func (i *Slack) getWebhook() string {
	return fmt.Sprintf("%s/%s/%s/%s", i.apiPrefix, i.config.TeamID, i.config.BotID, i.config.HookID)
}

func validateConfig(config Config) error {
	if config.TeamID == "" {
		return errors.New("teamId is empty")
	}

	if config.BotID == "" {
		return errors.New("botId is empty")
	}

	if config.HookID == "" {
		return errors.New("hookId is empty")
	}

	return nil
}
