package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"
)

type TestAction string

var (
	Run      TestAction = "run"
	Output   TestAction = "output"
	Pass     TestAction = "pass"
	Fail     TestAction = "fail"
	Start    TestAction = "start"
	Skip     TestAction = "skip"
	Pause    TestAction = "pause"
	Bench    TestAction = "bench"
	Continue TestAction = "cont"
)

type JsonLogItem struct {
	Time    time.Time
	Action  TestAction
	Test    string
	Output  string
	Elapsed *float64
}

var notedQuery = regexp.MustCompile(`(Expected|Executed)Query:\s+([^;]+)(?:;\s+(.+))?`)

func parser(verbose bool, r io.Reader, w io.Writer) (int, int, error) {
	input := bufio.NewScanner(r)

	expectedQueries := make(map[string]string)
	executedQueries := make(map[string]string)

	for {
		cont := input.Scan()
		if !cont {
			break
		}
		var data JsonLogItem
		err := json.Unmarshal(input.Bytes(), &data)
		if err != nil {
			log.Fatalf("Make sure you run tests with -json (%v)", err)
		}
		if data.Action == Output {
			data.Output = strings.TrimSuffix(data.Output, "\\n")
			data.Output = strings.TrimSpace(data.Output)
			groups := notedQuery.FindStringSubmatch(data.Output)
			if len(groups) == 0 {
				continue
			} else if len(groups) == 3 {
				groups = append(groups, "")
			}
			if groups[1] == "Expected" {
				if _, ok := expectedQueries[groups[2]]; !ok {
					expectedQueries[groups[2]] = groups[3]
				}
			} else if groups[1] == "Executed" {
				if _, ok := executedQueries[groups[2]]; !ok {
					executedQueries[groups[2]] = data.Test
				} else {
					executedQueries[groups[2]] = strings.Join([]string{executedQueries[groups[2]], data.Test}, ", ")
				}
			}
		}
	}

	found := 0
	tagCounts := make(map[string]int)
	output := make([]string, 0)
	for expected, tag := range expectedQueries {
		if tests, ok := executedQueries[expected]; !ok {
			tagSuffix := ""
			if tag != "" {
				tagSuffix = "."
			}
			tagCounts[tag]++
			output = append(output, fmt.Sprintf("Expected query not executed: %s%s%s", tag, tagSuffix, expected))
		} else {
			if verbose {
				output = append(output, fmt.Sprintf("Executed query: %s by %s", expected, tests))
			}
			found++
		}
	}

	slices.Sort(output)
	for _, line := range output {
		_, err := fmt.Fprintln(w, line)
		if err != nil {
			return found, len(expectedQueries), err
		}
	}
	if len(output) > 0 {
		_, err := fmt.Fprintln(w)
		if err != nil {
			return found, len(expectedQueries), err
		}
	}

	if len(tagCounts) > 1 {
		output = make([]string, 0)
		for tag, count := range tagCounts {
			output = append(output, fmt.Sprintf("%s: %d", tag, count))
		}
		slices.Sort(output)
		for _, line := range output {
			_, err := fmt.Fprintln(w, line)
			if err != nil {
				return found, len(expectedQueries), err
			}
		}
		_, err := fmt.Fprintln(w)
		if err != nil {
			return found, len(expectedQueries), err
		}
	}

	return found, len(expectedQueries), nil
}

func main() {
	log.SetFlags(0)
	verbose := flag.Bool("verbose", false, "Verbose output")
	printVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *printVersion {
		fmt.Printf("dotensure %s\n", version)
		return
	}

	found, expected, err := parser(*verbose, os.Stdin, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d / %d expected queries\n", found, expected)
	if found < expected {
		os.Exit(1)
	}
}
