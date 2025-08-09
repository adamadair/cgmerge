package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/atotto/clipboard"
)

func main() {
	clipboardFlag := flag.Bool("clipboard", false, "Copy output to clipboard")
	dirFlag := flag.String("dir", ".", "Directory to scan for Go files")
	langFlag := flag.String("lang", "go", "Language of the files to merge (default: go). Supported: go, csharp")
	flag.Parse()
	var merged string
	var err error
	if *langFlag == "go" {
		merged, err = MergeGoProjectFiles(*dirFlag)
	} else if *langFlag == "csharp" {
		merged, err = MergeCSharpProjectFiles(*dirFlag)
	} else {
		err = fmt.Errorf("unsupported language: %s", *langFlag)
	}

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error merging files: %v\n", err)
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
