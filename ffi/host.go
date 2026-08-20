// Package ffi hosts Go-side support for tinear's `extern fn` declarations
// (e.g. tui/../os.ard). Each exported symbol here pairs with an extern
// using the `tinear.` prefix.
package ffi

import (
	"fmt"
	"os"
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

// Spawn starts a program with args without waiting for it to exit.
// Exists because direct Go variadic calls from Ard accept at most one
// trailing argument. The child's output goes to a log file — tinear
// owns the terminal, so the child can't have it, but silent failures
// are undebuggable.
func Spawn(name string, args []string) error {
	cmd := exec.Command(name, args...)
	if f, err := os.OpenFile("/tmp/tinear-spawn.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err == nil {
		fmt.Fprintf(f, "--- spawn %s %v\n", name, args)
		cmd.Stdout = f
		cmd.Stderr = f
	}
	return cmd.Start()
}

// LookPath reports whether an executable is available on PATH.
func LookPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// CmdOutput runs a program and returns its combined output. Exists for
// the same variadic reason as Spawn; used for cheap capability probes.
func CmdOutput(name string, args []string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return string(out), err
}
