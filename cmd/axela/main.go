package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hagemt/axela/brains"
)

func defaultListenSocket() string {
	port := envString("HTTP_PORT", "9001")
	host := envString("BIND_HOST", "127.0.0.1")
	if sock := envString("BIND_SOCKET", ""); sock != "" {
		return sock
	}
	return fmt.Sprintf("%s:%s", host, port)
}

// envLoad pulls .env into os.Environ() for use by os.LookupEnv() later
// It's a stupid hack to act like we're using dotenv (which we could do)
func envLoad(filename string) {
	f, err := os.Open(filename) // #nosec G304 -- hard-coded .env file path
	if err != nil {
		panic(err)
	}
	defer brains.CloseAllQuietly(f)

	r := bufio.NewScanner(f)
	for r.Scan() {
		if s := strings.TrimSpace(r.Text()); s != "" && !strings.HasPrefix(s, "#") {
			if l, r, ok := strings.Cut(s, "="); ok {
				lhs := strings.TrimPrefix(l, "export ")
				rhs := strings.Trim(r, "'")
				_ = os.Setenv(lhs, rhs)
			}
		}
	}
	if err := r.Err(); err != nil {
		panic(err)
	}
}

func envString(name, defaultValue string) string {
	if s, ok := os.LookupEnv(name); ok {
		return s
	}
	return defaultValue
}

// startTextsynth starts an assistant named Axel powered by TextSynth
func startTextsynth(ctx context.Context, addr, spec string) error {
	if !strings.HasPrefix(spec, "textsynth:") {
		log.Println("ignored AXELA_ENGINE: ", spec)
		spec = "textsynth"
	}

	var kind, engine, secret string // pulled from .env
	if prefix, suffix, ok := strings.Cut(spec, ":"); ok {
		name := strings.ToUpper(prefix) + "_SECRET"
		if secret = envString(name, ""); secret == "" {
			return fmt.Errorf("missing .env: %s", name)
		}
		kind, engine = prefix, suffix
	}
	defaultChatEngine := "falcon_40B-chat"
	if engine == "" && kind == "textsynth" {
		log.Println("defaulting to engine: ", defaultChatEngine)
		engine = defaultChatEngine
	}
	b := withBrains(func(systemPrompt string) textBrain {
		switch strings.ToLower(kind) {
		case "textsynth":
			return brains.NewTextsynth(secret, engine, systemPrompt)
		default:
			log.Panicln("unknown AXELA_ENGINE: ", kind)
			return nil
		}
	})
	return newAssistant(b).start(ctx, addr)
}

func main() {
	envLoad(".env")

	var addr, specs string
	flag.StringVar(&specs, "llm", envString("AXELA_ENGINE", "textsynth:falcon_40B-chat"), "LLM")
	flag.StringVar(&addr, "http-addr", defaultListenSocket(), "bind HTTP to TCP/UNIX interface")
	flag.Parse()

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(context.Canceled)

	if err := startTextsynth(ctx, addr, specs); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
