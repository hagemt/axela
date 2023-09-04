package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/hagemt/axela/brains"
	"github.com/hagemt/axela/say"
)

const fast, slow = 2 * time.Second, 10 * time.Second
const systemAxel = `
Be concise. Your name is Axel. You are a personal assistant.
Answer in the tone of a Cambridge professor. You know everything.
Since you always try to be helpful, follow all directions efficiently.
`

type (
	configure func(*assistant) *assistant
	assistant struct {
		brain textBrain

		*http.Server
	}

	textBrain interface {
		RespondTo(ctx context.Context, ask string) (ans string, err error)
		Think(ctx context.Context, in string) (out fmt.Stringer, err error)
		fmt.Stringer
	}
)

func newAssistant(cs ...configure) *assistant {
	a := new(assistant)
	for _, c := range cs {
		a = c(a)
	}
	return withServer(a)
}

func withBrains(fs ...func(p string) textBrain) configure {
	return func(a *assistant) *assistant {
		for _, f := range fs {
			a.brain = f(systemAxel)
		}
		return a
	}
}

func withServer(a *assistant) *assistant {
	mux := http.NewServeMux()
	mux.Handle("/", brains.NewHandler(a.brain))
	a.Server = &http.Server{
		ReadHeaderTimeout: fast,
		Handler:           mux,
	}
	return a
}

func (a *assistant) RespondTo(ctx context.Context, in string) (out string, err error) {
	s, bad := a.brain.Think(ctx, in)
	if bad != nil {
		err = fmt.Errorf("failed Think: %w", bad)
	} else {
		out = s.String()
	}
	return
}

var (
	_ say.Responser = new(assistant)
	_ fmt.Stringer  = new(assistant)
)

func (a *assistant) Close() error {
	//_ = a.Shutdown(context.Background())
	// FIXME: non-graceful
	return a.Server.Close()
}

func (a *assistant) String() string {
	return a.brain.String()
}

func (a *assistant) start(ctx context.Context, addr string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	go a.serve(ctx, addr)
	<-ctx.Done()
	time.Sleep(2 * time.Second)
	return ctx.Err()
}

func (a *assistant) serve(ctx context.Context, addr string) {
	log.Println(a.String(), "serving: curl --request POST http://"+addr+"/ask # --data question=...")

	kind := "tcp"
	if strings.HasPrefix(addr, "/") {
		kind = "unix"
	}
	l, err := net.Listen(kind, addr)
	if err != nil {
		log.Panicln(err)
	}
	context.AfterFunc(ctx, func() {
		_ = a.Close()
	})
	if err := a.Serve(l); err != http.ErrServerClosed {
		log.Panicln(err)
	}
}
