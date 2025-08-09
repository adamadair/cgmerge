package main

import (
	"flag"
	"fmt"
	"github.com/atotto/clipboard"
	"os"
)

func main() {
	clipboardFlag := flag.Bool("clipboard", false, "Copy output to clipboard")
	dirFlag := flag.String("dir", ".", "Directory to scan for Go files")
	flag.Parse()
	merged, err := MergeGoProjectFiles(*dirFlag)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error merging Go files: %v\n", err)
		os.Exit(1)
	}
	// Output
	fmt.Print(merged)
	if *clipboardFlag {
		err := clipboard.WriteAll(merged)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Clipboard error: %v\n", err)
		} else {
			_, _ = fmt.Fprintln(os.Stderr, "Merged output copied to clipboard.")
		}
	}
}
