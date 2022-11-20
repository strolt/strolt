package telegram

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/logger"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Token  string `yaml:"token"`
	ChatID string `yaml:"chatId"`
}

type Params struct {
	apiPrefix string // changed only in tests
}

type Telegram struct {
	Params

	logger *logger.Logger
	config Config
}

const telegramAPIPrefix = "https://api.telegram.org/bot"

func New(params Params) *Telegram {
	res := Telegram{
		Params: params,
	}

	if res.Params.apiPrefix == "" {
		res.Params.apiPrefix = telegramAPIPrefix
	}

	return &res
}

func (i *Telegram) SetConfig(config interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &i.config); err != nil {
		return err
	}

	return validateConfig(i.config)
}

func (i *Telegram) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

func (i *Telegram) getURL() string {
	chatID := i.config.ChatID

	if _, err := strconv.ParseInt(chatID, 10, 64); err != nil {
		chatID = "@" + chatID // if chatID not a number enforce @ prefix
	}

	return fmt.Sprintf("%s%s/sendMessage?chat_id=%s&no_webpage=true", i.Params.apiPrefix, i.config.Token, chatID)
}

func (i *Telegram) Send(ctx context.Context) {
	body, err := getTemplate(ctx)
	if err != nil {
		i.logger.Error(err)
		return
	}

	resp, err := http.Post(i.getURL(), "application/json", body) //nolint:noctx
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
	if config.Token == "" {
		return fmt.Errorf("token is empty")
	}

	if config.ChatID == "" {
		return fmt.Errorf("chatId is empty")
	}

	return nil
}
