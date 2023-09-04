package brains

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/hagemt/axela/say"
)

type (
	Brain interface {
		RespondTo(ctx context.Context, ask string) (ans string, err error)
		Think(ctx context.Context, in string) (out fmt.Stringer, err error)
		fmt.Stringer
	}

	dAnswers struct {
		From string   `json:"source"`
		Text []string `json:"history,omitempty"`
	}
)

var answers sync.Map

func async(ctx context.Context, b Brain, in string) {
	//ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	//defer cancel()

	start := time.Now()
	s, err := say.Ask(ctx, in, b)
	if err != nil {
		log.Println(err)
		return
	}
	after := time.Now()
	q := fmt.Sprintf("%s Human: Please %s", start, in)
	a := fmt.Sprintf("%s Axela: %s", after, s)
	answers.Store(q, a)
}

func queue(a Brain, in string) (err error) {
	if in == "" {
		err = errors.New("empty question")
		return
	}
	go async(context.Background(), a, in)
	return
}

func getRecents(as *sync.Map, max int) []string {
	ss := make([]string, 0, max)
	as.Range(func(q, a interface{}) bool {
		ss = append(ss, q.(string), a.(string))
		return true
	})
	sort.Sort(sort.Reverse(sort.StringSlice(ss)))
	if len(ss) < max {
		return ss
	}
	return ss[:max]
}

func getHistory(a Brain) ([]byte, error) {
	data := &dAnswers{
		From: a.String(), // which Brain
		Text: getRecents(&answers, 100),
	}
	return json.Marshal(data)
}

func NewHandler(a Brain) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			b, err := getHistory(a)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b)
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			q := r.Form.Get("question")
			if err := queue(a, q); err != nil {
				w.WriteHeader(http.StatusNotAcceptable)
				log.Println("bad queue", q, err.Error())
				return
			}
			w.WriteHeader(http.StatusAccepted)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
