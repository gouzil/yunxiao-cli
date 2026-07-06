package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gouzi/yunxiao-cli/internal/app"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	application, err := app.New(app.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "X failed to initialize yunxiao: %v\n", err)
		os.Exit(1)
	}
	if err := application.Execute(context.Background(), os.Args[1:]); err != nil {
		os.Exit(app.ExitCode(err))
	}
}
