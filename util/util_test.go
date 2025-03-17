package util

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{
			name:           "Basic file name",
			input:          "test.yaml",
			expectedOutput: "test.yaml",
		},
		{
			name:           "Basic full path",
			input:          "/x/test.yaml",
			expectedOutput: "/x/test.yaml",
		},
		{
			name:           "Relative path",
			input:          "x/test.yaml",
			expectedOutput: "x/test.yaml",
		},
		{
			name:           "Relative path with traversal segment",
			input:          "../x/test.yaml",
			expectedOutput: "x/test.yaml",
		},
		{
			name:           "Relative path with traversal segment",
			input:          "/../x/test.yaml",
			expectedOutput: "/x/test.yaml",
		},
	}

	for _, tt := range tests {
		if result := SanitizePath(tt.input); result != tt.expectedOutput {
			t.Errorf("Test %q failed, expectedOutput %q, got %q", tt.name, tt.expectedOutput, result)
		}

	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		name        string
		enforcePath bool
		path        string
		expected    bool
	}{
		{

			name: "Empty file name", enforcePath: false,
			path:     "   ",
			expected: false,
		},
		{

			name:        "Directory does not exist",
			enforcePath: false,
			path:        "/tmp/not/exists/",
			expected:    false,
		},
		{

			name:        "File does not exist",
			enforcePath: false,
			path:        "/tmp/not/exists/",
			expected:    false,
		},
		{
			name:        "Is temp dir exists",
			enforcePath: false,
			path:        os.TempDir(),
			expected:    true,
		},
		{

			name:        "File just created in temp dir",
			enforcePath: true,
			path:        filepath.Join(os.TempDir(), "test"),
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.enforcePath {
				if err := enforceFile(tt.path); err != nil {
					t.Errorf("Unexpected error in process of enforceFile, err %v", err)
				}
			}

			got := Exists(tt.path)
			if got != tt.expected {
				t.Errorf("unexpected result, expected %t, got %t", tt.expected, got)
			}

		})
	}
}

func TestNormPath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error while os.Getwd(), err - %v", err)
	}

	tests := []struct {
		name          string
		input         string
		expected      string
		expectedError error
	}{
		{
			name:          "Test Absolute path Unix",
			input:         "/test/abc/x",
			expected:      "/test/abc/x",
			expectedError: nil,
		},
		{
			name:          "Test relative path Unix",
			input:         "test/abc/x",
			expected:      fmt.Sprintf("%s/test/abc/x", cwd),
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormPath(tt.input)
			if err != tt.expectedError {
				t.Errorf("unpexected error, expected %q, got %q", SafeErrorString(tt.expectedError), SafeErrorString(err))
			}

			if tt.expected != got {
				t.Errorf("unexpected result, expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func Test_normPath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error while os.Getwd(), err - %v", err)
	}

	tests := []struct {
		name          string
		input         string
		os            string
		expected      string
		expectedError error
	}{
		{
			name:          "Test Absolute path Unix",
			input:         "/test/abc/x",
			os:            "linux",
			expected:      "/test/abc/x",
			expectedError: nil,
		},
		{
			name:          "Test Absolute path Windows",
			input:         "\\\\?\\\\file.txt",
			os:            "windows",
			expected:      "\\\\?\\\\file.txt",
			expectedError: nil,
		},
		{
			name:          "Test relative path Unix",
			input:         "test/abc/x",
			os:            "linux",
			expected:      fmt.Sprintf("%s/test/abc/x", cwd),
			expectedError: nil,
		},
		{
			name:          "Test relative path Windows",
			input:         "file.txt",
			os:            "windows",
			expected:      fmt.Sprintf("\\\\?\\%s\\file.txt", strings.ReplaceAll(cwd, "/", `\`)),
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normPath(tt.input, tt.os)
			if err != tt.expectedError {
				t.Errorf("unpexected error, expected %q, got %q", SafeErrorString(tt.expectedError), SafeErrorString(err))
			}

			if tt.expected != got {
				t.Errorf("unexpected result, expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestSafeErrorString(t *testing.T) {
	tests := []struct {
		name     string
		input    error
		expected string
	}{
		{
			name:     "Nil error return expected value",
			input:    nil,
			expected: "<nil>",
		},
		{
			name:     "Not nil error return expected value",
			input:    fmt.Errorf("unexpected error"),
			expected: "unexpected error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeErrorString(tt.input)
			if tt.expected != got {
				t.Errorf("unexpected result, expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestTempFile(t *testing.T) {
	tests := []struct {
		name               string
		dir                string
		pattern            string
		fileNamePattern    *regexp.Regexp
		expectedErrPattern *regexp.Regexp
	}{
		{
			name:            "Directory does not exists, file creation failed",
			dir:             filepath.Join(os.TempDir(), "does_not_exists"),
			pattern:         "test_file",
			fileNamePattern: nil,
			expectedErrPattern: regexp.MustCompilePOSIX(
				fmt.Sprintf("open %s: no such file or directory", filepath.Join(os.TempDir(), "does_not_exists/test_file([0-9]+).tmp"))),
		},
		{
			name:               "Create file in existing directory, file created successfully",
			dir:                os.TempDir(),
			pattern:            "test_file",
			expectedErrPattern: regexp.MustCompilePOSIX("<nil>"),
			fileNamePattern:    regexp.MustCompilePOSIX(filepath.Join(os.TempDir(), "test_file([0-9]+).tmp")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fd, err := TempFile(tt.dir, tt.pattern, os.ModePerm)

			if !tt.expectedErrPattern.MatchString(SafeErrorString(err)) {
				t.Errorf("unexpected error, expected to match %q  got %q",
					tt.expectedErrPattern.String(), SafeErrorString(err))
			}

			if err == nil && !tt.fileNamePattern.MatchString(fd.Name()) {
				t.Errorf("unexpected filename, expected to match %q  got %q",
					tt.fileNamePattern.String(), fd.Name())
			}
		})
	}
}

func TestDefaultRunnerRun(t *testing.T) {
	tests := []struct {
		name           string
		cmd            *exec.Cmd
		expectedStdout []byte
		expectedStderr []byte
		expectedErr    error
	}{
		{
			name:           "Run echo \"Hello\", successfully finish",
			cmd:            exec.Command("echo", "Hello"),
			expectedStdout: []byte("Hello\n"),
			expectedStderr: []byte(""),
			expectedErr:    nil,
		},
		{
			name:           "Run script that fail, stderr propogated correctly",
			cmd:            exec.Command("bash", "not_exists.sh"),
			expectedStdout: []byte(""),
			expectedStderr: []byte("bash: not_exists.sh: No such file or directory\n"),
			expectedErr:    fmt.Errorf("exit status 127"),
		},
		{
			name:           "Run non existing binary, fail with expected error",
			cmd:            exec.Command("does_not_exists", "Hello"),
			expectedStdout: []byte(""),
			expectedStderr: []byte(""),
			expectedErr:    fmt.Errorf("exec: \"does_not_exists\": executable file not found in $PATH"),
		},
	}

	ctx, defaultRunner := context.Background(), &DefaultRunner{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := defaultRunner.Run(ctx, tt.cmd)
			if SafeErrorString(err) != SafeErrorString(tt.expectedErr) {
				t.Errorf("unexpected error, expected  %q  got %q", SafeErrorString(tt.expectedErr), SafeErrorString(err))
			}

			if !bytes.Equal(tt.expectedStdout, stdout) {
				t.Errorf("unexpected stdout, expected %+v, got %+v", tt.expectedStdout, stdout)
			}

			if !bytes.Equal(tt.expectedStderr, stderr) {
				t.Errorf("unexpected stderr, expected %+v, got %+v", tt.expectedStderr, stderr)
			}
		})
	}
}

func TestAtomicWrite(t *testing.T) {
	tests := []struct {
		name               string
		path               string
		content            []byte
		expectedErrPattern *regexp.Regexp
	}{
		{
			name:               "write file to the tmp dir",
			path:               filepath.Join(os.TempDir(), strconv.FormatInt(time.Now().Unix(), 16)),
			content:            []byte("test content"),
			expectedErrPattern: regexp.MustCompilePOSIX("<nil>"),
		},
		{
			name:               "Attempt to write file to the dir with no permissions, fail with expected error",
			path:               "/bin/does_not_exist",
			content:            []byte("test content"),
			expectedErrPattern: regexp.MustCompilePOSIX("unable to create temp file: open /bin/does_not_exist([0-9]+).tmp: permission denied"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AtomicWrite(tt.path, tt.content, os.ModePerm)
			if !tt.expectedErrPattern.MatchString(SafeErrorString(err)) {
				t.Fatalf("unexpected error, expected err to match %q, got %q", tt.expectedErrPattern, SafeErrorString(err))
			}

			if err != nil { //If err is not nil we don't want to compare content
				return
			}

			got, err := os.ReadFile(tt.path)
			if err != nil {
				t.Errorf("unexpected error, while reading the written file, %s", SafeErrorString(err))
			}

			if !bytes.Equal(tt.content, got) {
				t.Errorf("unexpected content, expected: %+v, got: %+v", tt.content, got)
			}
		})
	}
}

func enforceFile(path string) error {
	_, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	return nil
}
