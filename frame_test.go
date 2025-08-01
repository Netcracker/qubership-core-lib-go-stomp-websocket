package go_stomp_websocket

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFrame(t *testing.T) {
	tests := []struct {
		name    string
		command string
		headers []string
		want    *Frame
	}{
		{
			name:    "normal frame",
			command: "CONNECT",
			headers: []string{"version:1.2", "heart-beat:1000,1000"},
			want:    createTestFrame("CONNECT", []string{"version:1.2", "heart-beat:1000,1000"}, ""),
		},
		{
			name:    "empty headers",
			command: "CONNECT",
			headers: []string{},
			want:    createTestFrame("CONNECT", []string{}, ""),
		},
		{
			name:    "nil headers",
			command: "CONNECT",
			headers: nil,
			want:    createTestFrame("CONNECT", nil, ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateFrame(tt.command, tt.headers)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReadFrame(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    *Frame
		wantErr bool
	}{
		{
			name:  "normal frame",
			input: createTestInput("CONNECTED", []string{"version:1.2", "heart-beat:1000,1000"}, `{"test": "json"}`),
			want:  createTestFrame("CONNECTED", []string{"version:1.2", "heart-beat:1000,1000"}, `{"test": "json"}`),
		},
		{
			name:  "frame with escaped characters",
			input: createTestInput("CONNECTED", []string{"version:1.2", "header:value"}, "body"),
			want:  createTestFrame("CONNECTED", []string{"version:1.2", "header:value"}, "body"),
		},
		{
			name:  "frame with empty body",
			input: createTestInput("CONNECTED", []string{"version:1.2"}, ""),
			want:  createTestFrame("CONNECTED", []string{"version:1.2"}, ""),
		},
		{
			name:  "frame with empty headers",
			input: createTestInput("CONNECTED", nil, "body"),
			want:  createTestFrame("CONNECTED", nil, "body"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReadFrame(tt.input)
			if tt.wantErr {
				assert.Nil(t, got)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestFrame_Bytes(t *testing.T) {
	tests := []struct {
		name     string
		frame    *Frame
		expected string
	}{
		{
			name:     "normal frame",
			frame:    createTestFrame("CONNECTED", []string{"version:1.2", "heart-beat:1000,1000"}, `{"test": "json"}`),
			expected: `["CONNECTED\nversion:1.2\nheart-beat:1000,1000\n\n{"test": "json"}\u0000"]`,
		},
		{
			name:     "frame with empty body",
			frame:    createTestFrame("CONNECTED", []string{"version:1.2"}, ""),
			expected: `["CONNECTED\nversion:1.2\n\n\u0000"]`,
		},
		{
			name:     "frame with empty headers",
			frame:    createTestFrame("CONNECTED", []string{}, "body"),
			expected: `["CONNECTED\n\nbody\u0000"]`,
		},
		{
			name:     "frame with special characters",
			frame:    createTestFrame("CONNECTED", []string{"header:value"}, "body"),
			expected: `["CONNECTED\nheader:value\n\nbody\u0000"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tt.frame.Bytes())
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestFrame_Contains(t *testing.T) {
	tests := []struct {
		name          string
		frame         *Frame
		header        string
		expectedValue string
		expectedFound bool
	}{
		{
			name:          "existing header",
			frame:         createTestFrame("CONNECTED", []string{"version:1.2", "heart-beat:1000,1000"}, ""),
			header:        "version",
			expectedValue: "1.2",
			expectedFound: true,
		},
		{
			name:          "non-existent header",
			frame:         createTestFrame("CONNECTED", []string{"version:1.2"}, ""),
			header:        "nonexistent",
			expectedValue: "",
			expectedFound: false,
		},
		{
			name:          "empty headers",
			frame:         createTestFrame("CONNECTED", []string{}, ""),
			header:        "version",
			expectedValue: "",
			expectedFound: false,
		},
		{
			name:          "header with no value",
			frame:         createTestFrame("CONNECTED", []string{"version:"}, ""),
			header:        "version",
			expectedValue: "",
			expectedFound: true,
		},
		{
			name:          "header with special characters",
			frame:         createTestFrame("CONNECTED", []string{"header:value\nwith\nnewlines"}, ""),
			header:        "header",
			expectedValue: "value\nwith\nnewlines",
			expectedFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, found := tt.frame.Contains(tt.header)
			assert.Equal(t, tt.expectedValue, value)
			assert.Equal(t, tt.expectedFound, found)
		})
	}
}

//-----------------------------------------------------------------------------------

func createTestFrame(command string, headers []string, body string) *Frame {
	return &Frame{
		Command: command,
		Headers: headers,
		Body:    body,
	}
}

func createTestInput(command string, headers []string, body string) []byte {
	var result []string
	result = append(result, "a[\"")
	result = append(result, command+"\\n")
	for _, header := range headers {
		result = append(result, header+"\\n")
	}
	result = append(result, "\\n")
	result = append(result, body)
	result = append(result, "\\u0000\"]")
	return []byte(strings.Join(result, ""))
}
