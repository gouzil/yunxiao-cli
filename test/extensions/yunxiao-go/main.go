package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Printf("args=%s\n", strings.Join(os.Args[1:], "|"))
	for _, name := range []string{
		"YUNXIAO_EXTENSION",
		"YUNXIAO_EXTENSION_NAME",
		"YUNXIAO_EXTENSION_DIR",
		"YUNXIAO_ENDPOINT",
		"YUNXIAO_ORGANIZATION",
		"YUNXIAO_PROJECT",
		"YUNXIAO_REPO",
	} {
		fmt.Printf("%s=%s\n", name, os.Getenv(name))
	}
	fmt.Fprintln(os.Stderr, "stderr=go")
}
