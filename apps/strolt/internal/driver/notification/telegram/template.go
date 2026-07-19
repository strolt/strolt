package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"

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

	// Header and Body carry dynamic content (service/task names, snapshot IDs
	// and the operation error text). Escape them so characters like <, > and &
	// do not form invalid HTML entities that make Telegram reject the message
	// with HTTP 400 and silently drop the alert. CopyrightHTML is a fixed,
	// intentional <a> tag and must stay literal.
	msg := fmt.Sprintf("<b>%s</b>", html.EscapeString(t.Header))
	msg += "\n\n" + html.EscapeString(t.Body)
	msg += "\n\n " + t.CopyrightHTML

	body := telegramMsg{Text: msg, ParseMode: "HTML"}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal telegram message: %w", err)
	}

	return bytes.NewBuffer(data), nil
}
