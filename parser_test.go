package main

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func jsonToStream(entries []JsonLogItem) (io.Reader, error) {
	sb := strings.Builder{}
	for _, entry := range entries {
		line, err := json.Marshal(entry)
		if err != nil {
			return nil, err
		}
		sb.Write(line)
		sb.WriteString("\n")
	}
	return strings.NewReader(sb.String()), nil
}

func TestParser_BasicCase(t *testing.T) {
	var input = []JsonLogItem{
		{Action: Output, Output: "ExpectedQuery: a; foo"},
		{Action: Output, Output: "something else"},
		{Action: Output, Output: "ExpectedQuery: b; bar"},
		{Action: Output, Output: "ExpectedQuery: a"},
		{Action: Output, Output: "ExpectedQuery: c\\n"},
		{Action: Output, Output: "ExpectedQuery: d; bar"},
		{Action: Fail, Output: "test failed"},
		{Action: Output, Output: "ExecutedQuery: a"},
		{Action: Output, Output: "ExecutedQuery: a; bar\\n"},
	}
	text, err := jsonToStream(input)
	require.NoError(t, err)
	var buffer bytes.Buffer
	found, expected, err := parser(false, text, &buffer)
	require.NoError(t, err)
	assert.Equal(t, 1, found)
	assert.Equal(t, 4, expected)

	output := `Expected query not executed: bar.b
Expected query not executed: bar.d
Expected query not executed: c

: 1
bar: 2

`
	assert.Equal(t, output, buffer.String())
}
