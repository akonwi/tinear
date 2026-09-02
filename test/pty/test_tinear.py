#!/usr/bin/env python3
"""Live-terminal smoke coverage for Tinear and its Cooper command routing."""

import base64
import os
import signal
import stat
import sys
import tempfile
import time

from harness import Screen, binary_path, build, read_for, resize, send, spawn, wait_exit, wait_for

ROOT = os.path.normpath(os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", ".."))


def sgr(fd, code, col, row, final="M"):
    send(fd, f"\x1b[<{code};{col + 1};{row + 1}{final}")


def click(fd, col, row):
    sgr(fd, 0, col, row)
    sgr(fd, 0, col, row, "m")


def drag(fd, start_col, start_row, end_col, end_row):
    sgr(fd, 0, start_col, start_row)
    sgr(fd, 32, end_col, end_row)
    sgr(fd, 0, end_col, end_row, "m")


def wait_for_bytes(fd, screen, needle, timeout=4.0):
    captured = b""
    deadline = time.time() + timeout
    while time.time() < deadline:
        captured += read_for(fd, screen, 0.05)
        if needle in captured:
            return captured
    raise AssertionError(f"did not see terminal bytes {needle!r}; captured={captured!r}")


def cleanup(fd, pid):
    try:
        os.close(fd)
    except OSError:
        pass
    try:
        done, _ = os.waitpid(pid, os.WNOHANG)
    except ChildProcessError:
        return
    if done == pid:
        return
    try:
        os.kill(pid, signal.SIGTERM)
    except ProcessLookupError:
        return
    deadline = time.time() + 1.0
    while time.time() < deadline:
        try:
            done, _ = os.waitpid(pid, os.WNOHANG)
        except ChildProcessError:
            return
        if done == pid:
            return
        time.sleep(0.02)
    try:
        os.kill(pid, signal.SIGKILL)
    except ProcessLookupError:
        pass
    try:
        os.waitpid(pid, 0)
    except ChildProcessError:
        pass


def production_startup_smoke():
    binary = build("tinear", source="main.ard")
    with tempfile.TemporaryDirectory() as home, tempfile.TemporaryDirectory() as temp:
        wrapper = os.path.join(temp, "run-tinear")
        with open(wrapper, "w", encoding="utf-8") as file:
            file.write(
                "#!/bin/sh\n"
                "before=$(stty -a | sed -E 's/-?pendin/pendin/g' | tr -s '[:space:]' ' ')\n"
                f'exec_path="{binary}"\n'
                '"$exec_path"\n'
                "status=$?\n"
                "after=$(stty -a | sed -E 's/-?pendin/pendin/g' | tr -s '[:space:]' ' ')\n"
                'if [ "$before" = "$after" ]; then restored=yes; else restored=no; fi\n'
                'printf "__EXIT__=%s __TTY_RESTORED__=%s\\n" "$status" "$restored"\n'
                "exit \"$status\"\n"
            )
        os.chmod(wrapper, stat.S_IRUSR | stat.S_IWUSR | stat.S_IXUSR)
        pid, fd = spawn(wrapper, rows=24, cols=80, env={"HOME": home, "LINEAR_API_KEY": ""})
        screen = Screen(24, 80)
        try:
            wait_for(fd, screen, "Enter your Linear API key")
            send(fd, "\x03")
            wait_for(fd, screen, "__TTY_RESTORED__=yes")
            status = wait_exit(pid, fd, screen, timeout=2.0)
            assert status == 0, f"production exit status {status}"
        finally:
            cleanup(fd, pid)


def fixture_smoke():
    binary = build("tinear-pty-smoke", source="test/fixtures/pty_smoke.ard")
    home = tempfile.TemporaryDirectory()
    pid, fd = spawn(binary, rows=24, cols=100, env={"HOME": home.name})
    screen = Screen(24, 100)
    try:
        # A restored dynamic tab proves cached shell state reaches the live terminal.
        wait_for(fd, screen, "RESTORED DETAIL Restored doc")
        resize(fd, rows=20, cols=60)
        screen = Screen(20, 60)
        wait_for(fd, screen, "RESTORED DETAIL Restored doc")
        resize(fd, rows=24, cols=100)
        screen = Screen(24, 100)
        wait_for(fd, screen, "RESTORED DETAIL Restored doc")

        # Number keys, Tab, and Shift+Tab exercise vaxis key normalization through
        # Tinear's real shell command router.
        send(fd, "1")
        wait_for(fd, screen, "COPY-ME")
        send(fd, "\t")
        wait_for(fd, screen, "BOARD BODY")
        send(fd, "\x1b[Z")
        wait_for(fd, screen, "COPY-ME")

        # Bracketed paste is delivered as one normalized multiline value.
        send(fd, "\x1b[200~one\r\ntwo\x1b[201~")
        wait_for(fd, screen, "value: one↵two")

        # SGR mouse clicks route through the tab bar.
        click(fd, col=8, row=0)
        wait_for(fd, screen, "BOARD BODY")
        click(fd, col=1, row=0)
        wait_for(fd, screen, "COPY-ME")

        # Wheel input scrolls the retained viewport in the Inbox fixture.
        for _ in range(3):
            sgr(fd, 65, col=2, row=8)
            read_for(fd, screen, 0.05)
        assert "scroll row 0" not in screen.line(7), "wheel did not move retained scroll content"

        # Select live Text and use Tinear's Super+C routing to emit OSC 52.
        drag(fd, start_col=0, start_row=2, end_col=6, end_row=2)
        copied = base64.b64encode("COPY-ME".encode())
        send(fd, "\x1b[99;9u")
        wait_for_bytes(fd, screen, b"\x1b]52;c;" + copied + b"\x1b\\")

        send(fd, "\x03")
        status = wait_exit(pid, fd, screen, timeout=2.0)
        if status is None:
            raise AssertionError("Tinear PTY fixture did not exit after Ctrl+C")
        assert status == 0, f"fixture exit status {status}"
    finally:
        cleanup(fd, pid)
        home.cleanup()


def main():
    os.chdir(ROOT)
    production_startup_smoke()
    fixture_smoke()
    print("✓ Tinear startup/input/mouse/paste/clipboard/quit PTY smoke passed")


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        print(f"FAIL: {error}", file=sys.stderr)
        sys.exit(1)
