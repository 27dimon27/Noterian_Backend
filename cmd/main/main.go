package main

import (
	"log"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/run"

	_ "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/docs"
)

func main() {
	if err := run.Run(); err != nil {
		log.Fatalf("Error to run application: %v", err)
	}
}
