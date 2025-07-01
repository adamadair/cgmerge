package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
)

func main() {
	clipboardFlag := flag.Bool("clipboard", false, "Copy output to clipboard")
	dirFlag := flag.String("dir", ".", "Directory to scan for Go files")
	flag.Parse()

	var files []string
	filepath.Walk(*dirFlag, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() &&
			strings.HasSuffix(info.Name(), ".go") &&
			!strings.HasSuffix(info.Name(), "_test.go") &&
			!strings.HasPrefix(info.Name(), "_merged") {
			files = append(files, path)
		}
		return nil
	})

	sort.Strings(files)

	var imports []string
	importSet := make(map[string]struct{})
	var result []string
	packageWritten := false

	importSingleRe := regexp.MustCompile(`import\s+"([^"]+)"`)

	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", file, err)
			continue
		}
		lines := strings.Split(string(data), "\n")
		outLines := []string{}

		for i := 0; i < len(lines); i++ {
			line := lines[i]
			// Handle package main (only include once)
			if strings.HasPrefix(line, "package ") {
				if !packageWritten {
					outLines = append(outLines, line)
					packageWritten = true
				}
				continue
			}
			// Handle import block
			if strings.HasPrefix(line, "import (") {
				block := ""
				i++
				for i < len(lines) && !strings.HasPrefix(lines[i], ")") {
					block += lines[i] + "\n"
					i++
				}
				for _, imp := range strings.Split(block, "\n") {
					imp = strings.TrimSpace(imp)
					if imp == "" {
						continue
					}
					// remove comments, aliases, etc. for simplicity
					imp = strings.Split(imp, " ")[0]
					imp = strings.Trim(imp, `"`)
					if _, exists := importSet[imp]; !exists && imp != "" {
						importSet[imp] = struct{}{}
						imports = append(imports, fmt.Sprintf("\t\"%s\"", imp))
					}
				}
				continue
			}
			// Handle single import line
			if importSingleRe.MatchString(line) {
				imp := importSingleRe.FindStringSubmatch(line)[1]
				if _, exists := importSet[imp]; !exists {
					importSet[imp] = struct{}{}
					imports = append(imports, fmt.Sprintf("\t\"%s\"", imp))
				}
				continue
			}
			outLines = append(outLines, line)
		}

		result = append(result, fmt.Sprintf("// --- %s ---", file))
		result = append(result, outLines...)
	}

	// Assemble everything
	final := []string{}
	final = append(final, "package main")
	if len(imports) > 0 {
		final = append(final, "import (")
		final = append(final, imports...)
		final = append(final, ")")
	}
	final = append(final, result...)
	merged := strings.Join(final, "\n")

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
