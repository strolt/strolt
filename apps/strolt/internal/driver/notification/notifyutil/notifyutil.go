// Package notifyutil provides shared helpers for notification drivers.
package notifyutil

import (
	"errors"
	"net/url"
)

// RedactURLError removes the request URL from an *url.Error so that secrets
// embedded in notification endpoints (Telegram bot tokens, Slack webhook paths,
// custom webhook URLs) are not written to logs. HTTP client calls such as
// http.Post return an *url.Error whose Error() includes the full request URL;
// logging it verbatim would leak the secret to any log sink.
func RedactURLError(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		urlErr.URL = "[redacted]"
	}

	return err
}
