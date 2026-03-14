package main

import (
	"context"

	"github.com/cmmasaba/prototypes/urlshortener/cmd"
)

func main() {
	cmd.StartApplication(context.Background())
}
