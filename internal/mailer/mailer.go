package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"intelhorizons/internal/store"
)

const resendEndpoint = "https://api.resend.com/emails"

type Mailer struct {
	apiKey  string
	from    string
	siteURL string
	client  *http.Client
}

func New(apiKey, from, siteURL string) *Mailer {
	return &Mailer{
		apiKey:  apiKey,
		from:    from,
		siteURL: strings.TrimRight(siteURL, "/"),
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (m *Mailer) SendConfirmation(ctx context.Context, email, token string) error {
	link := m.siteURL + "/api/subscribe/confirm?token=" + token
	html, err := renderConfirmation(link)
	if err != nil {
		return fmt.Errorf("render confirmation: %w", err)
	}
	return m.send(ctx, email, "Confirm your IntelHorizons subscription", html)
}

func (m *Mailer) SendDigest(ctx context.Context, subscribers []store.Subscriber, articles []store.Article, date time.Time) error {
	subject := "AI News Digest — " + date.Format("January 2, 2006")
	for _, sub := range subscribers {
		unsubLink := m.siteURL + "/api/unsubscribe?token=" + sub.Token
		html, err := renderDigest(articles, date, m.siteURL, unsubLink)
		if err != nil {
			return fmt.Errorf("render digest for %s: %w", sub.Email, err)
		}
		if err := m.send(ctx, sub.Email, subject, html); err != nil {
			return fmt.Errorf("send to %s: %w", sub.Email, err)
		}
	}
	return nil
}

func (m *Mailer) send(ctx context.Context, to, subject, html string) error {
	payload, err := json.Marshal(map[string]any{
		"from":    m.from,
		"to":      []string{to},
		"subject": subject,
		"html":    html,
	})
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendEndpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var resErr struct {
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&resErr) //nolint:errcheck
		return fmt.Errorf("resend API %d: %s", resp.StatusCode, resErr.Message)
	}
	return nil
}

const confirmHTML = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>Confirm subscription</title></head>
<body style="margin:0;padding:32px;background:#0f172a;font-family:ui-sans-serif,system-ui,sans-serif">
  <div style="max-width:560px;margin:0 auto;background:#1e293b;border-radius:12px;overflow:hidden">
    <div style="padding:32px;text-align:center;border-bottom:1px solid #334155">
      <p style="margin:0;font-size:13px;letter-spacing:.1em;text-transform:uppercase;color:#64748b">IntelHorizons</p>
      <h1 style="margin:8px 0 0;font-size:22px;color:#e2e8f0">Confirm your subscription</h1>
    </div>
    <div style="padding:32px;text-align:center">
      <p style="margin:0 0 24px;color:#94a3b8;line-height:1.6">
        Click the button below to start receiving the daily AI news digest.
      </p>
      <a href="{{.ConfirmLink}}"
         style="display:inline-block;padding:12px 28px;background:#6366f1;color:#fff;
                border-radius:8px;text-decoration:none;font-weight:600;font-size:15px">
        Confirm subscription
      </a>
      <p style="margin:24px 0 0;font-size:12px;color:#475569">
        If you didn't request this, you can safely ignore this email.
      </p>
    </div>
  </div>
</body>
</html>`

const digestHTML = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>AI News Digest</title></head>
<body style="margin:0;padding:32px;background:#0f172a;font-family:ui-sans-serif,system-ui,sans-serif">
  <div style="max-width:640px;margin:0 auto;background:#1e293b;border-radius:12px;overflow:hidden">
    <div style="padding:28px 32px;border-bottom:1px solid #334155">
      <p style="margin:0;font-size:13px;letter-spacing:.1em;text-transform:uppercase;color:#64748b">IntelHorizons</p>
      <h1 style="margin:6px 0 0;font-size:20px;color:#e2e8f0">AI News Digest</h1>
      <p style="margin:4px 0 0;font-size:13px;color:#64748b">{{.DateStr}}</p>
    </div>
    <div style="padding:24px 32px">
      {{range $i, $a := .Articles}}
      <div style="{{if $i}}margin-top:20px;padding-top:20px;border-top:1px solid #334155{{end}}">
        <div style="margin-bottom:6px">
          <span style="font-size:12px;font-weight:700;color:#475569">{{inc $i}}.</span>
          <span style="font-size:11px;color:#6366f1;text-transform:uppercase;letter-spacing:.05em;margin-left:6px">{{$a.SourceName}}</span>
        </div>
        <a href="{{$a.URL}}"
           style="display:block;margin-left:20px;font-size:15px;font-weight:600;
                  color:#e2e8f0;text-decoration:none;line-height:1.4">
          {{$a.Title}}
        </a>
        {{if $a.Summary}}
        <p style="margin:6px 0 0 20px;font-size:13px;color:#94a3b8;line-height:1.5">
          {{truncate $a.Summary 200}}
        </p>
        {{end}}
      </div>
      {{end}}
    </div>
    <div style="padding:20px 32px;border-top:1px solid #334155;text-align:center">
      <a href="{{.SiteURL}}"
         style="font-size:13px;color:#6366f1;text-decoration:none">
        View full feed at intelhorizons.com
      </a>
      <span style="color:#475569;margin:0 8px">·</span>
      <a href="{{.UnsubLink}}"
         style="font-size:13px;color:#475569;text-decoration:none">
        Unsubscribe
      </a>
    </div>
  </div>
</body>
</html>`

var (
	confirmTmpl = template.Must(template.New("confirm").Parse(confirmHTML))
	digestTmpl  = template.Must(
		template.New("digest").Funcs(template.FuncMap{
			"inc": func(i int) int { return i + 1 },
			"truncate": func(s string, n int) string {
				if len(s) <= n {
					return s
				}
				return s[:n] + "…"
			},
		}).Parse(digestHTML),
	)
)

func renderConfirmation(link string) (string, error) {
	var buf bytes.Buffer
	if err := confirmTmpl.Execute(&buf, map[string]string{"ConfirmLink": link}); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderDigest(articles []store.Article, date time.Time, siteURL, unsubLink string) (string, error) {
	var buf bytes.Buffer
	if err := digestTmpl.Execute(&buf, map[string]any{
		"Articles":  articles,
		"DateStr":   date.Format("Monday, January 2, 2006"),
		"SiteURL":   siteURL,
		"UnsubLink": unsubLink,
	}); err != nil {
		return "", err
	}
	return buf.String(), nil
}
