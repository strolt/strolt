package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/template"
)

// telegramMsg is used to send message trough Telegram bot API.
type telegramMsg struct {
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func getTemplate(ctx context.Context) (*bytes.Buffer, error) {
	t := template.New("telegram", ctx)

	msg := fmt.Sprintf("<b>%s</b>", t.Header)
	msg += fmt.Sprintf("\n\n%s", t.Body)
	msg += fmt.Sprintf("\n\n %s", t.CopyrightHTML)

	body := telegramMsg{Text: msg, ParseMode: "HTML"}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(data), nil
}
