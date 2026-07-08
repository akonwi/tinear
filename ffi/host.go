// Package ffi hosts Go-side support for tinear's `extern fn` declarations
// (e.g. tui/../os.ard). Each exported symbol here pairs with an extern
// using the `tinear.` prefix.
package ffi

import (
	"os/exec"
	goruntime "runtime"
)

// OpenURL launches the system's default handler for url (browser for
// http(s), etc.). Backs the `tinear.OpenURL` extern declared in os.ard.
//
// Returns once the process is spawned; we do not wait for the launched
// program to exit.
func OpenURL(url string) error {
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

// StrFromBytes converts raw bytes to a string. Stopgap until the language
// provides Str::from_bytes (filed upstream); JSON marshalling returns [Byte].
func StrFromBytes(b []byte) string {
	return string(b)
}
