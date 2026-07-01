package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestRun(t *testing.T) {
	cases := []struct {
		files      map[string]string
		name       string
		version    string
		wantOut    string
		wantErrSub string
		args       []string
		dirs       []string
		wantCode   int
	}{
		{
			name:    "default lists visible entries",
			args:    []string{"ls", "/dir"},
			files:   map[string]string{"/dir/.hidden": "", "/dir/alpha.txt": "", "/dir/bravo.txt": ""},
			wantOut: "alpha.txt\nbravo.txt\n",
		},
		{
			name:    "all shows hidden entries",
			args:    []string{"ls", "-a", "/dir"},
			files:   map[string]string{"/dir/.hidden": "", "/dir/alpha.txt": ""},
			wantOut: ".hidden\nalpha.txt\n",
		},
		{
			name:    "recursive walks subdirectories",
			args:    []string{"ls", "-R", "/dir"},
			files:   map[string]string{"/dir/a.txt": "", "/dir/sub/b.txt": ""},
			wantOut: "a.txt\nsub\nsub/b.txt\n",
		},
		{
			name:    "long format emits perm size name",
			args:    []string{"ls", "-l", "/dir"},
			files:   map[string]string{"/dir/a.txt": "alpha"},
			wantOut: "-rw-r--r-- 5 a.txt\n",
		},
		{
			name:    "default path is current directory",
			args:    []string{"ls"},
			dirs:    []string{"."},
			wantOut: "",
		},
		{
			name:    "version flag reports injected version",
			version: "1.2.3",
			args:    []string{"ls", "--version"},
			wantOut: "ls version 1.2.3\n",
		},
		{
			name:       "unknown flag errors",
			args:       []string{"ls", "--nope"},
			wantCode:   1,
			wantErrSub: "ls:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			for _, dir := range tc.dirs {
				if err := fs.MkdirAll(dir, 0o755); err != nil {
					t.Fatalf("mkdir fixture %s: %v", dir, err)
				}
			}
			for path, content := range tc.files {
				if err := afero.WriteFile(fs, path, []byte(content), 0o644); err != nil {
					t.Fatalf("write fixture %s: %v", path, err)
				}
			}

			var out, errOut bytes.Buffer
			code := run(tc.version, tc.args, strings.NewReader(""), &out, &errOut, fs)

			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr=%q)", code, tc.wantCode, errOut.String())
			}
			if tc.wantErrSub == "" && out.String() != tc.wantOut {
				t.Fatalf("stdout = %q, want %q", out.String(), tc.wantOut)
			}
			if tc.wantErrSub != "" && !strings.Contains(errOut.String(), tc.wantErrSub) {
				t.Fatalf("stderr = %q, want substring %q", errOut.String(), tc.wantErrSub)
			}
		})
	}
}

func Test_main(t *testing.T) {
	origExit, origRun := osExit, runCLI
	t.Cleanup(func() { osExit, runCLI = origExit, origRun })

	gotCode := -1
	osExit = func(code int) { gotCode = code }
	runCLI = func(string, []string, io.Reader, io.Writer, io.Writer, afero.Fs) int { return 7 }

	main()

	if gotCode != 7 {
		t.Fatalf("main propagated exit code %d, want 7", gotCode)
	}
}
