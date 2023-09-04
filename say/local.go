package say

import (
	"context"
	"fmt"
	"log"
	"os/exec"
)

type Responser interface {
	RespondTo(ctx context.Context, in string) (out string, err error)
	fmt.Stringer
}

func Ask(ctx context.Context, ask string, from Responser) (ans string, err error) {
	log.Println("---", from.String(), "was asked:", ask)
	a, err := from.RespondTo(ctx, ask)
	if err != nil {
		return
	}
	log.Println("---", from.String(), "said back:", a)
	if err := say(ctx, a); err != nil {
		log.Println("say error:", err)
	}
	return a, nil
}

func say(ctx context.Context, text string) error {
	return exec.CommandContext(ctx, "/usr/bin/say", text).Run()
}
