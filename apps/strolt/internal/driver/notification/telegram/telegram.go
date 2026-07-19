// Package telegram implements a notification driver that sends messages via the Telegram bot API.
package telegram

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/driver/notification/notifyutil"
	"github.com/strolt/strolt/shared/logger"

	"gopkg.in/yaml.v3"
)

// Config holds the Telegram bot token and target chat.
type Config struct {
	Token  string `yaml:"token"`
	ChatID string `yaml:"chatId"`
}

// Params holds optional driver parameters.
type Params struct {
	apiPrefix string // changed only in tests
}

// Telegram is a notification driver that sends messages through the Telegram bot API.
type Telegram struct {
	Params

	logger *logger.Logger
	config Config
}

const telegramAPIPrefix = "https://api.telegram.org/bot"

// New creates a Telegram driver with the given parameters.
func New(params Params) *Telegram {
	res := Telegram{
		Params: params,
	}

	if res.apiPrefix == "" {
		res.apiPrefix = telegramAPIPrefix
	}

	return &res
}

// SetConfig parses and validates the driver configuration.
func (i *Telegram) SetConfig(config any) error {
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
func (i *Telegram) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

// Send posts a notification message built from the context to the Telegram chat.
func (i *Telegram) Send(ctx context.Context) {
	body, err := getTemplate(ctx)
	if err != nil {
		i.logger.Error(err)
		return
	}

	resp, err := http.Post(i.getURL(), "application/json", body) //nolint:noctx
	if err != nil {
		// The transport error embeds the request URL, which contains the bot
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

func (i *Telegram) getURL() string {
	chatID := i.config.ChatID

	if _, err := strconv.ParseInt(chatID, 10, 64); err != nil {
		chatID = "@" + chatID // if chatID not a number enforce @ prefix
	}

	return fmt.Sprintf("%s%s/sendMessage?chat_id=%s&no_webpage=true", i.apiPrefix, i.config.Token, chatID)
}

func validateConfig(config Config) error {
	if config.Token == "" {
		return errors.New("token is empty")
	}

	if config.ChatID == "" {
		return errors.New("chatId is empty")
	}

	return nil
}
