package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

// captureOutput captures stdout during the execution of f.
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestStatusColor(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{100, cyan},
		{101, cyan},
		{199, cyan},
		{200, green},
		{201, green},
		{299, green},
		{300, yellow},
		{301, yellow},
		{399, yellow},
		{400, red},
		{404, red},
		{499, red},
		{500, magenta},
		{503, magenta},
		{599, magenta},
		{0, ""},
		{99, ""},
		{600, ""},
		{-1, ""},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("code_%d", tt.code), func(t *testing.T) {
			got := statusColor(tt.code)
			if got != tt.want {
				t.Errorf("statusColor(%d) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestPrintStatus_ValidCode(t *testing.T) {
	output := captureOutput(func() {
		printStatus("200", false)
	})

	if !strings.Contains(output, "200") {
		t.Error("expected output to contain status code 200")
	}
	if !strings.Contains(output, "OK") {
		t.Error("expected output to contain status text 'OK'")
	}
}

func TestPrintStatus_InvalidInput(t *testing.T) {
	output := captureOutput(func() {
		printStatus("abc", false)
	})

	if !strings.Contains(output, "Invalid code: abc") {
		t.Errorf("expected 'Invalid code: abc', got %q", output)
	}
}

func TestPrintStatus_UnknownCode(t *testing.T) {
	output := captureOutput(func() {
		printStatus("999", false)
	})

	if !strings.Contains(output, "Unknown code: 999") {
		t.Errorf("expected 'Unknown code: 999', got %q", output)
	}
}

func TestPrintStatus_VerboseMode(t *testing.T) {
	output := captureOutput(func() {
		printStatus("404", true)
	})

	if !strings.Contains(output, "404") {
		t.Error("expected output to contain status code 404")
	}
	if !strings.Contains(output, "Common causes:") {
		t.Error("expected verbose output to contain 'Common causes:'")
	}
	if !strings.Contains(output, "RFC:") {
		t.Error("expected verbose output to contain 'RFC:'")
	}
}

func TestPrintStatus_VerboseNoExtra(t *testing.T) {
	// A code that exists in net/http but not in our maps should not print verbose sections
	output := captureOutput(func() {
		printStatus("200", true)
	})

	// 200 has verbose info, so check it's present
	if !strings.Contains(output, "Common causes:") {
		t.Error("expected verbose output for 200 to contain 'Common causes:'")
	}
}

func TestPrintStatus_NonVerboseNoExtras(t *testing.T) {
	output := captureOutput(func() {
		printStatus("404", false)
	})

	if strings.Contains(output, "Common causes:") {
		t.Error("non-verbose output should not contain verbose sections")
	}
}

func TestPrintStatus_AllKnownCodes(t *testing.T) {
	for code := range statusDescriptions {
		t.Run(fmt.Sprintf("code_%d", code), func(t *testing.T) {
			output := captureOutput(func() {
				printStatus(fmt.Sprintf("%d", code), false)
			})

			if !strings.Contains(output, fmt.Sprintf("%d", code)) {
				t.Errorf("output for code %d should contain the code number", code)
			}

			statusText := http.StatusText(code)
			if statusText != "" && !strings.Contains(output, statusText) {
				t.Errorf("output for code %d should contain status text %q", code, statusText)
			}
		})
	}
}

func TestStatusDescriptions_HaveMatchingVerbose(t *testing.T) {
	for code := range statusDescriptions {
		if _, ok := verboseDescriptions[code]; !ok {
			t.Errorf("code %d has a status description but no verbose description", code)
		}
	}
}

func TestVerboseDescriptions_HaveMatchingStatus(t *testing.T) {
	for code := range verboseDescriptions {
		if _, ok := statusDescriptions[code]; !ok {
			t.Errorf("code %d has a verbose description but no status description", code)
		}
	}
}

func TestStatusDescriptions_NotEmpty(t *testing.T) {
	for code, desc := range statusDescriptions {
		if strings.TrimSpace(desc) == "" {
			t.Errorf("status description for code %d is empty", code)
		}
	}
}

func TestVerboseDescriptions_NotEmpty(t *testing.T) {
	for code, desc := range verboseDescriptions {
		if strings.TrimSpace(desc) == "" {
			t.Errorf("verbose description for code %d is empty", code)
		}
	}
}

func TestStatusDescriptions_ValidHTTPCodes(t *testing.T) {
	for code := range statusDescriptions {
		if code < 100 || code >= 600 {
			t.Errorf("code %d is outside valid HTTP status code range (100-599)", code)
		}
	}
}

func TestVerboseSectionLabels(t *testing.T) {
	expectedPrefixes := []string{
		"Common causes:",
		"Real-world usage:",
		"Related codes:",
		"Troubleshooting:",
		"RFC:",
	}

	if len(verboseSectionLabels) != len(expectedPrefixes) {
		t.Fatalf("expected %d section labels, got %d", len(expectedPrefixes), len(verboseSectionLabels))
	}

	for i, expected := range expectedPrefixes {
		if verboseSectionLabels[i].prefix != expected {
			t.Errorf("section label %d: got prefix %q, want %q", i, verboseSectionLabels[i].prefix, expected)
		}
		if verboseSectionLabels[i].color == "" {
			t.Errorf("section label %d (%s) has empty color", i, expected)
		}
	}
}

func TestPrintStatus_CodeClasses(t *testing.T) {
	// Test one code from each class to ensure they all format correctly
	codes := []struct {
		code  string
		class string
	}{
		{"100", "1xx"},
		{"200", "2xx"},
		{"301", "3xx"},
		{"404", "4xx"},
		{"500", "5xx"},
	}

	for _, tt := range codes {
		t.Run(tt.class, func(t *testing.T) {
			output := captureOutput(func() {
				printStatus(tt.code, false)
			})
			if output == "" {
				t.Errorf("expected output for %s class code %s, got empty", tt.class, tt.code)
			}
		})
	}
}

func TestPrintStatus_EmptyString(t *testing.T) {
	output := captureOutput(func() {
		printStatus("", false)
	})

	if !strings.Contains(output, "Invalid code:") {
		t.Errorf("expected 'Invalid code:' for empty input, got %q", output)
	}
}
