package brains

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type (
	aResponse string
	roundTrip func(*http.Request) (*http.Response, error)
	textsynth struct {
		authSecret string
		engineName string

		prompt string
		system string

		asks *http.Client
	}
)

func (a aResponse) String() string {
	return string(a)
}

func (a roundTrip) RoundTrip(r *http.Request) (re *http.Response, err error) {
	//log.Println(r.Method, r.URL)
	started := time.Now()
	re, err = a(r)
	elapsed := time.Since(started)
	if err != nil {
		log.Println(r.Method, r.URL, "failed", err)
	} else {
		log.Println(r.Method, r.URL, re.Status, "within", elapsed)
	}
	return
}

func NewTextsynth(authSecret, engineName, systemPrompt string) *textsynth {
	engine := fmt.Sprintf("https://api.textsynth.com/v1/engines/%s", engineName)
	prefix, _ := url.Parse(engine)

	to := 10 * time.Second
	dt := http.DefaultTransport.(*http.Transport).Clone()
	rt := roundTrip(func(r *http.Request) (*http.Response, error) {
		r.Header.Set("Authorization", "Bearer "+authSecret)
		if p := r.URL.Path; !strings.HasPrefix(p, "/") {
			r.URL = prefix.JoinPath(p) // e.g. completions
		} else {
			r.URL = prefix.ResolveReference(r.URL)
		}
		return dt.RoundTrip(r)
	})

	return &textsynth{
		authSecret: authSecret,
		engineName: engineName,

		prompt: "Please ",
		system: systemPrompt,

		asks: &http.Client{
			Transport: rt,
			Timeout:   to,
		},
	}
}

func (a *textsynth) RespondTo(ctx context.Context, in string) (ans string, err error) {
	o, err := a.Think(ctx, in)
	if err != nil {
		return
	}
	ans = o.String()
	return
}

func (a *textsynth) String() string {
	return "ai:textsynth:" + a.engineName
}

func (a *textsynth) Think(ctx context.Context, in string) (out fmt.Stringer, err error) {
	data := map[string]any{
		"messages": []string{a.prompt + in},
		"system":   a.system,
	}
	b, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, "chat", bytes.NewBuffer(b))
	if err != nil {
		return
	}
	r.Header.Set("Content-Type", "application/json")
	s, err := a.asks.Do(r)
	if err != nil {
		return
	}
	defer CloseAllQuietly(s.Body)

	if s.StatusCode != http.StatusOK {
		err = fmt.Errorf("not 200 OK: %s", dump(b, s))
		return
	}
	var reply struct {
		Text string `json:"text"`
	}
	if err = json.NewDecoder(s.Body).Decode(&reply); err != nil {
		return
	}
	out = aResponse(reply.Text)
	return
}
