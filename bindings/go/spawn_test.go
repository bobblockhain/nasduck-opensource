package nasduck_test

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	nasduck "github.com/shiplaone2q/nasduck/bindings/go"
)

func TestRetryTerminalProcess(t *testing.T) {
	if !slices.Contains(os.Args, "--tui-go-retry-child") {
		return
	}
	const marker = "first-launch"
	_, err := os.Stat(marker)
	if errors.Is(err, os.ErrNotExist) {
		requireNoError(t, os.WriteFile(marker, []byte("first launch"), 0600))
	} else {
		requireNoError(t, err)
		fmt.Print("\x1b]133;B\x07retry-ready\r\n")
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Printf("received:%s\r\n", scanner.Text())
	}
	requireNoError(t, scanner.Err())
	os.Exit(0)
}

func clientLaunchTimeouts() nasduck.Timeouts {
	return nasduck.Timeouts{
		Text: nasduck.Ptr(11 * time.Millisecond), Idle: nasduck.Ptr(12 * time.Millisecond),
		Command: nasduck.Ptr(13 * time.Millisecond), Exit: nasduck.Ptr(14 * time.Millisecond),
		Ready: nasduck.Ptr(15 * time.Millisecond),
	}
}

type launchTimeoutCase struct {
	name     string
	options  nasduck.Timeouts
	expected nasduck.EffectiveTimeouts
}

func launchTimeoutCases() []launchTimeoutCase {
	zero := time.Duration(0)
	return []launchTimeoutCase{
		{name: "inherit", expected: nasduck.EffectiveTimeouts{Text: 11 * time.Millisecond, Idle: 12 * time.Millisecond, Command: 13 * time.Millisecond, Exit: 14 * time.Millisecond, Ready: 15 * time.Millisecond}},
		{name: "partial override", options: nasduck.Timeouts{Text: &zero, Ready: nasduck.Ptr(42 * time.Millisecond)}, expected: nasduck.EffectiveTimeouts{Idle: 12 * time.Millisecond, Command: 13 * time.Millisecond, Exit: 14 * time.Millisecond, Ready: 42 * time.Millisecond}},
		{name: "explicit zero", options: nasduck.Timeouts{Text: &zero, Idle: &zero, Command: &zero, Exit: &zero, Ready: &zero}},
	}
}

func TestClientTimeoutDefaultsReachLaunch(t *testing.T) {
	for _, launch := range []string{"open", "run"} {
		for _, testCase := range launchTimeoutCases() {
			t.Run(launch+"/"+testCase.name, func(t *testing.T) { verifyLaunchTimeouts(t, launch, testCase) })
		}
	}
}

func verifyLaunchTimeouts(t *testing.T, launch string, testCase launchTimeoutCase) {
	t.Helper()
	terminal := newClient(t, nasduck.ClientOptions{Timeouts: clientLaunchTimeouts()})
	options := nasduck.SpawnOptions{WaitReady: nasduck.Ptr(false), Timeouts: testCase.options}
	if launch == "run" {
		runTerminal(t, terminal, options)
	} else {
		_, err := terminal.Open(nasduck.OpenOptions{SpawnOptions: options})
		requireNoError(t, err)
	}
	state, err := terminal.State()
	requireNoError(t, err)
	if state.Timeouts != testCase.expected {
		t.Fatalf("launch timeouts=%+v, expected %+v", state.Timeouts, testCase.expected)
	}
}

func TestInvalidLaunchRetryPreservesSharedSession(t *testing.T) {
	for _, launch := range []string{"open", "run"} {
		t.Run(launch, func(t *testing.T) {
			terminal := newTerminal(t, nasduck.ClientOptions{})
			peer, err := nasduck.New(terminal.Session(), nasduck.ClientOptions{})
			requireNoError(t, err)
			options := nasduck.SpawnOptions{Backend: nasduck.Backend("unknown"), Retries: 1, Restart: nasduck.Ptr(false)}
			if launch == "run" {
				_, err = peer.Run("ignored", nil, options)
			} else {
				_, err = peer.Open(nasduck.OpenOptions{SpawnOptions: options})
			}
			requireKind(t, err, nasduck.UsageError)
			requireNoError(t, terminal.GetByText("go-ready", nasduck.TextSelectorOptions{}).Expect(nasduck.LocatorExpectOptions{Timeout: nasduck.Ptr(time.Duration(0))}))
			requireNoError(t, terminal.Submit(nasduck.Ptr("session-survived")))
			requireNoError(t, peer.GetByText("received:session-survived", nasduck.TextSelectorOptions{}).Expect(nasduck.LocatorExpectOptions{Timeout: nasduck.Ptr(time.Second)}))
		})
	}
}

func TestLaunchRetriesReadinessFailure(t *testing.T) {
	directory := t.TempDir()
	terminal := newClient(t, nasduck.ClientOptions{})
	executable, err := os.Executable()
	requireNoError(t, err)
	marker := filepath.Join(directory, "first-launch")
	result, err := terminal.Run(executable, []string{"-test.run=^TestRetryTerminalProcess$", "--", "--tui-go-retry-child"}, nasduck.SpawnOptions{
		Cwd: directory, WaitReady: nasduck.Ptr(true), Retries: 1,
		Timeouts: nasduck.Timeouts{Ready: nasduck.Ptr(3 * time.Second)},
	})
	requireNoError(t, err)
	if !result.Ready {
		t.Fatal("retry returned before the second launch was ready")
	}
	_, err = os.Stat(marker)
	requireNoError(t, err)
	requireNoError(t, terminal.Submit(nasduck.Ptr("retried")))
	requireNoError(t, terminal.GetByText("received:retried", nasduck.TextSelectorOptions{}).Expect(nasduck.LocatorExpectOptions{Timeout: nasduck.Ptr(3 * time.Second)}))
}
