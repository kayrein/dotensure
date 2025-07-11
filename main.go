package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
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

var notedQuery = regexp.MustCompile(`(Expected|Executed)Query:\s+(.+)`)

func main() {
	printVersion := flag.Bool("v", false, "print the version and exit")
	flag.Parse()

	if *printVersion {
		fmt.Printf("dotensure %s\n", version)
		return
	}

	input := bufio.NewScanner(os.Stdin)

	expectedQueries := make(map[string]struct{})
	executedQueries := make(map[string]struct{})

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
			groups := notedQuery.FindStringSubmatch(strings.TrimSuffix(data.Output, "\\n"))
			if len(groups) == 0 {
				continue
			}
			if groups[1] == "Expected" {
				expectedQueries[groups[2]] = struct{}{}
			} else if groups[1] == "Executed" {
				executedQueries[groups[2]] = struct{}{}
			}
		}
	}

	found := 0
	for expected, _ := range expectedQueries {
		if _, ok := executedQueries[expected]; !ok {
			fmt.Printf("Expected query not executed: %s\n", expected)
		} else {
			found++
		}
	}
	fmt.Printf("Found %d / %d expected queries\n", found, len(expectedQueries))
	if found < len(expectedQueries) {
		os.Exit(1)
	}
}
