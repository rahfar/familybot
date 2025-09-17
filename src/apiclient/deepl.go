package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/rahfar/familybot/src/db"
)

type DeeplAPI struct {
	DBClient *db.Client
	HttpClient  *http.Client
	BaseURL     string
	ApiKey      string
}

type TranslationIn struct {
	Text       []string `json:"text"`
	TargetLang string   `json:"target_lang"`
}
type TranslationOut struct {
	Translations []*Translation `json:"translations"`
}
type Translation struct {
	SourceLang string `json:"detected_source_language"`
	Text       string `json:"text"`
}

func (a *DeeplAPI) CallDeeplAPI(text []string) (string, error) {
	const maxRetry = 3

	ctx := context.Background()
	v, err := a.DBClient.GetTranslation(ctx, text)
	if err == nil {
		slog.Info("hit deeplapi cache", "key", a.DBClient.DeepLKey(text))
		return v, nil
	}

	body, err := json.Marshal(TranslationIn{Text: text, TargetLang: "RU"})
	if err != nil {
		return "", err
	}
	bodyReader := bytes.NewReader(body)

	req, err := http.NewRequest("POST", a.BaseURL+"/v2/translate", bodyReader)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "DeepL-Auth-Key "+a.ApiKey)

	for i := 1; i <= maxRetry; i++ {
		resp, err := a.HttpClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		if resp.StatusCode/100 == 2 {
			var t TranslationOut
			if err := json.Unmarshal(body, &t); err != nil {
				return "", err
			}
			if len(t.Translations) > 0 {
				err := a.DBClient.SetTranslation(ctx, text, t.Translations[0].Text)
				if err != nil {
					slog.Info("could not write cache", "err", err)
				}
				return t.Translations[0].Text, nil
			} else {
				return "", fmt.Errorf("no translation")
			}
		}

		if i < maxRetry {
			slog.Info("got error response from api, retrying in 5 seconds...", "retry-cnt", i, "status", resp.Status, "body", string(body))
			time.Sleep(5 * time.Second)
		} else {
			return "", fmt.Errorf("got error response from api: %s - %s", resp.Status, string(body))
		}
	}

	return "", fmt.Errorf("max retries reached")
}
