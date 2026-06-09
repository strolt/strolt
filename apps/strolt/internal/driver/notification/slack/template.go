package slack

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/template"
)

type slackMsgBlockText struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji,omitempty"`
}

type slackMsgBlock struct {
	Type string            `json:"type"`
	Text slackMsgBlockText `json:"text"`
}

type slackMsgBlockElements struct {
	Type     string              `json:"type"`
	Elements []slackMsgBlockText `json:"elements"`
}

type slackMsg struct {
	Blocks []any `json:"blocks"`
}

func getTemplate(ctx context.Context) (*bytes.Buffer, error) {
	t := template.New("slack", ctx)

	body := slackMsg{
		Blocks: []any{
			slackMsgBlock{
				Type: "header",
				Text: slackMsgBlockText{
					Type:  "plain_text",
					Text:  t.Header,
					Emoji: true,
				},
			},
			slackMsgBlock{
				Type: "section",
				Text: slackMsgBlockText{
					Type: "plain_text",
					Text: t.Body,
				},
			},
			slackMsgBlockElements{
				Type: "context",
				Elements: []slackMsgBlockText{
					{
						Type: "mrkdwn",
						Text: t.CopyrightMarkdown,
					},
				},
			},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal slack message: %w", err)
	}

	return bytes.NewBuffer(data), nil
}
