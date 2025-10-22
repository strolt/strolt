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

type Config struct {
	TeamID string `yaml:"teamId"`
	BotID  string `yaml:"botId"`
	HookID string `yaml:"hookId"`
}

type Params struct {
	apiPrefix string // changed only in tests
}

type Slack struct {
	Params

	logger *logger.Logger
	config Config
}

const slackAPIPrefix = "https://hooks.slack.com/services"

func New(params Params) *Slack {
	res := Slack{
		Params: params,
	}

	if res.Params.apiPrefix == "" {
		res.Params.apiPrefix = slackAPIPrefix
	}

	return &res
}

func (i *Slack) SetConfig(config interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &i.config); err != nil {
		return err
	}

	return validateConfig(i.config)
}

func (i *Slack) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

func (i *Slack) getWebhook() string {
	return fmt.Sprintf("%s/%s/%s/%s", i.Params.apiPrefix, i.config.TeamID, i.config.BotID, i.config.HookID)
}

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

	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= http.StatusBadRequest {
		i.logger.WithField("status", resp.Status).Error(string(b))
	}
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
