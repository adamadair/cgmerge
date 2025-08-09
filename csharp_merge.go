package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func MergeCSharpProjectFiles(dir string) (string, error) {
	var files []string
	var globalUsingsFile string

	// Walk directory to find .cs files
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip obj directory
		if info.IsDir() && info.Name() == "obj" {
			return filepath.SkipDir
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".cs") {
			// Exclude merged files, test files, and auto-generated files
			if strings.HasSuffix(info.Name(), "merged.cs") ||
				strings.Contains(strings.ToLower(info.Name()), "test") ||
				strings.HasSuffix(info.Name(), ".Designer.cs") ||
				strings.HasSuffix(info.Name(), ".g.cs") {
				return nil
			}

			// Check for GlobalUsings.cs
			if strings.ToLower(info.Name()) == "globalusings.cs" {
				globalUsingsFile = path
			} else {
				files = append(files, path)
			}
		}
		return nil
	})

	sort.Strings(files)

	var usings []string
	usingSet := make(map[string]struct{})
	var result []string
	var mainNamespace string

	// Regular expressions for parsing C# code
	usingRe := regexp.MustCompile(`^\s*using\s+([^;]+);`)
	globalUsingRe := regexp.MustCompile(`^\s*global\s+using\s+([^;]+);`)
	namespaceRe := regexp.MustCompile(`^\s*namespace\s+([^\s{;]+)`)
	fileScopedNamespaceRe := regexp.MustCompile(`^\s*namespace\s+([^\s{;]+)\s*;\s*$`)
	mainMethodRe := regexp.MustCompile(`\bstatic\s+void\s+Main\s*\(`)

	var isFileScopedNamespace bool

	// Process GlobalUsings.cs first if it exists
	if globalUsingsFile != "" {
		data, err := os.ReadFile(globalUsingsFile)
		if err != nil {
			return "", fmt.Errorf("error reading %s: %v", globalUsingsFile, err)
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if globalUsingRe.MatchString(line) {
				usingStmt := globalUsingRe.FindStringSubmatch(line)[1]
				usingStmt = strings.TrimSpace(usingStmt)
				if _, exists := usingSet[usingStmt]; !exists && usingStmt != "" {
					usingSet[usingStmt] = struct{}{}
					usings = append(usings, fmt.Sprintf("using %s;", usingStmt))
				}
			}
		}
	}

	// First pass: find the main namespace and detect if using file-scoped namespaces
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("error reading %s: %v", file, err)
		}

		content := string(data)
		lines := strings.Split(content, "\n")

		// Check for file-scoped namespace first
		for _, line := range lines {
			if fileScopedNamespaceRe.MatchString(line) {
				isFileScopedNamespace = true
				if mainMethodRe.MatchString(content) {
					mainNamespace = fileScopedNamespaceRe.FindStringSubmatch(line)[1]
					break
				}
			}
		}

		// If found Main method and namespace, we're done
		if mainNamespace != "" && mainMethodRe.MatchString(content) {
			break
		}

		// Otherwise, check for traditional namespace
		if mainMethodRe.MatchString(content) {
			for _, line := range lines {
				if namespaceRe.MatchString(line) && !fileScopedNamespaceRe.MatchString(line) {
					mainNamespace = namespaceRe.FindStringSubmatch(line)[1]
					break
				}
			}
			break
		}
	}

	// If no namespace found with Main, use the first namespace encountered
	if mainNamespace == "" {
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if fileScopedNamespaceRe.MatchString(line) {
					isFileScopedNamespace = true
					mainNamespace = fileScopedNamespaceRe.FindStringSubmatch(line)[1]
					break
				} else if namespaceRe.MatchString(line) {
					mainNamespace = namespaceRe.FindStringSubmatch(line)[1]
					break
				}
			}
			if mainNamespace != "" {
				break
			}
		}
	}

	// Second pass: process all files
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("error reading %s: %v", file, err)
		}

		lines := strings.Split(string(data), "\n")
		outLines := []string{}
		insideNamespace := false
		namespaceIndent := ""

		for _, line := range lines {
			// Handle using statements
			if usingRe.MatchString(line) {
				usingStmt := usingRe.FindStringSubmatch(line)[1]
				usingStmt = strings.TrimSpace(usingStmt)
				if _, exists := usingSet[usingStmt]; !exists && usingStmt != "" {
					usingSet[usingStmt] = struct{}{}
					usings = append(usings, fmt.Sprintf("using %s;", usingStmt))
				}
				continue
			}

			// Handle file-scoped namespace declarations (skip them entirely)
			if fileScopedNamespaceRe.MatchString(line) {
				continue
			}

			// Handle traditional namespace declarations
			if namespaceRe.MatchString(line) && !fileScopedNamespaceRe.MatchString(line) {
				insideNamespace = true
				// Capture the indentation for proper formatting
				namespaceIndent = line[:len(line)-len(strings.TrimLeft(line, " \t"))]
				continue
			}

			// Handle namespace opening brace
			if insideNamespace && strings.TrimSpace(line) == "{" {
				continue
			}

			// Handle namespace closing brace (last closing brace in file)
			if insideNamespace && strings.TrimSpace(line) == "}" {
				// Check if this is the last meaningful line
				remainingLines := false
				for i := len(lines) - 1; i >= 0; i-- {
					if strings.TrimSpace(lines[i]) != "" && strings.TrimSpace(lines[i]) != "}" {
						remainingLines = true
						break
					}
				}
				if !remainingLines {
					continue
				}
			}

			// Remove one level of indentation if we're inside a namespace
			if insideNamespace && len(line) > 0 {
				// Remove namespace-level indentation
				if strings.HasPrefix(line, namespaceIndent+"    ") {
					line = line[len(namespaceIndent+"    "):]
				} else if strings.HasPrefix(line, namespaceIndent+"\t") {
					line = line[len(namespaceIndent+"\t"):]
				}
			}

			outLines = append(outLines, line)
		}

		result = append(result, fmt.Sprintf("// --- %s ---", file))
		result = append(result, outLines...)
	}

	// Assemble everything
	final := []string{}

	// Add using statements
	if len(usings) > 0 {
		sort.Strings(usings)
		final = append(final, usings...)
		final = append(final, "")
	}

	// Add namespace declaration if we found one
	if mainNamespace != "" {
		if isFileScopedNamespace {
			// Use file-scoped namespace (C# 10+)
			final = append(final, fmt.Sprintf("namespace %s;", mainNamespace))
			final = append(final, "")
			final = append(final, result...)
		} else {
			// Use traditional namespace with braces
			final = append(final, fmt.Sprintf("namespace %s", mainNamespace))
			final = append(final, "{")

			// Indent all content
			for _, line := range result {
				if line == "" {
					final = append(final, "")
				} else {
					final = append(final, "    "+line)
				}
			}

			final = append(final, "}")
		}
	} else {
		final = append(final, result...)
	}

	merged := strings.Join(final, "\n")
	return merged, nil
}
