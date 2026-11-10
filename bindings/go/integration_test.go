package nasduck_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	nasduck "github.com/shiplaone2q/nasduck/bindings/go"
)

func TestTerminalProcess(t *testing.T) {
	if !slices.Contains(os.Args, "--tui-go-child") {
		return
	}
	fmt.Print("\x1b[2J\x1b[Hgo-ready\r\nitem item\r\n\x1b[1mstyled\x1b[0m plain\r\n界 café\r\n\x1b]8;;https://example.test/docs\x1b\\linked docs\x1b]8;;\x1b\\\r\n")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		switch scanner.Text() {
		case "quit":
			os.Exit(0)
		case "events":
			fmt.Print("\x1b]2;Go title\x07\x1b]52;c;Z28tY2xpcGJvYXJk\x07\x07\r\nevents-done\r\n")
		case "history":
			for i := 0; i < 100; i++ {
				fmt.Printf("history-%d\r\n", i)
			}
		case "raster":
			fmt.Print("\x1b[2J\x1b[Hgo-raster-ready\r\n")
		default:
			fmt.Printf("received:%s\r\n", scanner.Text())
		}
	}
	os.Exit(0)
}

func TestNativeScrollback(t *testing.T) {
	terminal := newTerminal(t, nasduck.ClientOptions{Backend: nasduck.Alacritty})
	requireNoError(t, terminal.Submit(nasduck.Ptr("history")))
	requireNoError(t, terminal.GetByText("history-99", nasduck.TextSelectorOptions{}).Expect(nasduck.LocatorExpectOptions{}))
	full, err := terminal.Text(nasduck.TextOptions{Full: true})
	requireNoError(t, err)
	bottom, err := terminal.GetViewport()
	requireNoError(t, err)
	if bottom.Offset != 0 || bottom.History == 0 {
		t.Fatalf("unexpected bottom viewport: %+v", bottom)
	}
	requireNoError(t, terminal.Scroll(2))
	requireNoError(t, terminal.Scroll(-1))
	view, err := terminal.GetViewport()
	requireNoError(t, err)
	if view.Offset != 1 {
		t.Fatalf("negative scroll did not move down: %+v", view)
	}
	requireNoError(t, terminal.ScrollToTop())
	view, err = terminal.GetViewport()
	requireNoError(t, err)
	if view.Offset != view.History {
		t.Fatalf("unexpected top viewport: %+v", view)
	}
	stillFull, err := terminal.Text(nasduck.TextOptions{Full: true})
	requireNoError(t, err)
	if stillFull != full {
		t.Fatal("native scroll changed full text")
	}
	requireNoError(t, terminal.ScrollToBottom())
	view, err = terminal.GetViewport()
	requireNoError(t, err)
	if view.Offset != 0 {
		t.Fatalf("bottom did not restore live screen: %+v", view)
	}
	requireNoError(t, terminal.Scroll(2147483647))
	requireNoError(t, terminal.Scroll(-2147483648))
	view, err = terminal.GetViewport()
	requireNoError(t, err)
	if view.Offset != 0 {
		t.Fatalf("signed scroll limits did not clamp: %+v", view)
	}
}

func newTerminal(t *testing.T, options nasduck.ClientOptions) *nasduck.Client {
	t.Helper()
	terminal := newClient(t, options)
	runTerminal(t, terminal, nasduck.SpawnOptions{})
	return terminal
}

func newClient(t *testing.T, options nasduck.ClientOptions) *nasduck.Client {
	t.Helper()
	terminal, err := nasduck.Ephemeral(t.Name(), options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := terminal.Close(); err != nil {
			t.Errorf("close terminal: %v", err)
		}
	})
	return terminal
}
func runTerminal(t *testing.T, terminal *nasduck.Client, options nasduck.SpawnOptions) {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	options.WaitReady = nasduck.Ptr(false)
	if _, err = terminal.Run(executable, []string{"-test.run=^TestTerminalProcess$", "--", "--tui-go-child"}, options); err != nil {
		t.Fatal(err)
	}
	if err = terminal.GetByText("go-ready", nasduck.TextSelectorOptions{}).Expect(nasduck.LocatorExpectOptions{Timeout: nasduck.Ptr(10 * time.Second)}); err != nil {
		t.Fatal(err)
	}
}
func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func requireKind(t *testing.T, err error, kind nasduck.ErrorKind) {
	t.Helper()
	var failure *nasduck.Error
	if !errors.As(err, &failure) || failure.Kind != kind {
		t.Fatalf("expected %s error, got %v", kind, err)
	}
}

func TestLocatorsPreserveSelection(t *testing.T) {
	terminal := newTerminal(t, nasduck.ClientOptions{})
	items := terminal.GetByText("item", nasduck.TextSelectorOptions{})
	count, err := items.Count()
	requireNoError(t, err)
	if count != 2 {
		t.Fatalf("item count=%d", count)
	}
	_, err = items.Location()
	requireKind(t, err, nasduck.AssertionError)
	first, err := items.First().Location()
	requireNoError(t, err)
	last, err := items.Last().Location()
	requireNoError(t, err)
	if first.Start.Column >= last.Start.Column {
		t.Fatalf("unordered locations: %+v %+v", first, last)
	}
	requireNoError(t, items.Highlight(nasduck.WaitOptions{}))
	requireNoError(t, items.First().Click(nasduck.LocatorClickOptions{}))
	requireNoError(t, terminal.GetByText("absent", nasduck.TextSelectorOptions{}).Wait(nasduck.LocatorWaitOptions{State: nasduck.Hidden, Timeout: nasduck.Ptr(time.Duration(0))}))
}

func TestLocatorEnumerationAndRelativeQueries(t *testing.T) {
	terminal := newTerminal(t, nasduck.ClientOptions{})
	items := terminal.GetByText("item", nasduck.TextSelectorOptions{})
	all, err := items.All()
	requireNoError(t, err)
	for index, locator := range all {
		actual, locationErr := locator.Location()
		requireNoError(t, locationErr)
		selected, selectedErr := items.Nth(uint32(index)).Location()
		requireNoError(t, selectedErr)
		if actual.Start != selected.Start {
			t.Fatalf("all[%d]=%+v; nth=%+v", index, actual, selected)
		}
	}
	count, err := items.Count()
	requireNoError(t, err)
	if count != 2 {
		t.Fatal("selection mutated original locator")
	}
	requireNoError(t, terminal.GetByText("item item", nasduck.TextSelectorOptions{}).GetByText("item", nasduck.TextSelectorOptions{}).Nth(1).Expect(nasduck.LocatorExpectOptions{}))
	requireNoError(t, items.First().GetByText("item", nasduck.TextSelectorOptions{Direction: nasduck.After}).Expect(nasduck.LocatorExpectOptions{}))
}

func TestStylesAndWideCellCoordinates(t *testing.T) {
	terminal := newTerminal(t, nasduck.ClientOptions{})
	bold := true
	styled := terminal.GetByStyle(nasduck.TextStyle{Bold: &bold}, nasduck.StyleSelectorOptions{})
	bold = false
	requireNoError(t, styled.GetByText("styled", nasduck.TextSelectorOptions{}).Expect(nasduck.LocatorExpectOptions{}))
	requireNoError(t, terminal.GetByText("plain", nasduck.TextSelectorOptions{}).GetByStyle(nasduck.TextStyle{Bold: nasduck.Ptr(false)}, nasduck.StyleSelectorOptions{}).Expect(nasduck.LocatorExpectOptions{}))
	wide, err := terminal.GetByText("界", nasduck.TextSelectorOptions{}).Location()
	requireNoError(t, err)
	if wide.Start.Column != 0 || wide.Start.Row != 3 {
		t.Fatalf("unexpected wide cell coordinates: %+v", wide.Start)
	}
	cells, err := terminal.Cells(0, 3, 2, 1)
	requireNoError(t, err)
	if len(cells) != 2 || cells[0].Char != "界" || cells[1].Char != "" {
		t.Fatalf("wide cells=%+v", cells)
	}
}

func TestLinkAndComposableLocators(t *testing.T) {
	terminal := newTerminal(t, nasduck.ClientOptions{})
	row := terminal.GetByText("linked docs", nasduck.TextSelectorOptions{})
	link := terminal.GetByLink("https://example.test/docs", nasduck.LinkSelectorOptions{})
	intersection, err := row.And(link)
	requireNoError(t, err)
	match, err := intersection.Location()
	requireNoError(t, err)
	if match.Text != "linked docs" {
		t.Fatalf("intersection=%+v", match)
	}
	union, err := terminal.GetByText("absent", nasduck.TextSelectorOptions{}).Or(link)
	requireNoError(t, err)
	requireNoError(t, union.Expect(nasduck.LocatorExpectOptions{}))
	filtered, err := row.Filter(nasduck.LocatorFilterOptions{Has: link})
	requireNoError(t, err)
	requireNoError(t, filtered.Expect(nasduck.LocatorExpectOptions{}))
	linked, err := link.Location()
	requireNoError(t, err)
	cells, err := terminal.Cells(uint16(linked.Start.Column), uint16(linked.Start.Row), 1, 1)
	requireNoError(t, err)
	if len(cells) != 1 || cells[0].Link != "https://example.test/docs" {
		t.Fatalf("linked cell=%+v", cells)
	}
}

func TestInputAndState(t *testing.T) {
	terminal := newTerminal(t, nasduck.ClientOptions{})
	state, err := terminal.State()
	requireNoError(t, err)
	if state.Cols != 80 || state.Rows != 30 {
		t.Fatalf("default size=%dx%d", state.Cols, state.Rows)
	}
	if state.Cursor.Shape == "" || state.Cursor.Color == "" || state.Colors.Foreground == "" || state.Colors.Background == "" || state.Colors.Cursor == "" || state.Modes == nil || state.MouseMode == "" {
		t.Fatalf("incomplete rich state=%+v", state)
	}
	requireNoError(t, terminal.Resize(90, 32))
	size, err := terminal.GetSize()
	requireNoError(t, err)
	if size.Cols != 90 || size.Rows != 32 {
		t.Fatalf("resized size=%+v", size)
	}
	_, err = terminal.GetCursor()
	requireNoError(t, err)
	_, err = terminal.GetCommand()
	requireNoError(t, err)
	_, err = terminal.GetOutput()
	requireNoError(t, err)
	_, err = terminal.GetExitCode()
	requireNoError(t, err)
	_, err = terminal.GetCwd()
	requireNoError(t, err)
	requireNoError(t, terminal.Type("hello"))
	requireNoError(t, terminal.Press("Enter"))
	requireNoError(t, terminal.GetByText("received:hello", nasduck.TextSelectorOptions{}).Expect(nasduck.LocatorExpectOptions{}))
	requireNoError(t, terminal.Submit(nasduck.Ptr("quit")))
	requireNoError(t, terminal.WaitExit(nasduck.WaitOptions{Timeout: nasduck.Ptr(10 * time.Second)}))
}

func TestTerminalEvents(t *testing.T) {
	terminal := newTerminal(t, nasduck.ClientOptions{})
	requireNoError(t, terminal.Submit(nasduck.Ptr("events")))
	requireNoError(t, terminal.WaitTitle("Go title", nasduck.TitleOptions{}))
	requireNoError(t, terminal.ExpectTitle("^Go", nasduck.TitleOptions{Regex: true}))
	requireNoError(t, terminal.WaitClipboard(nasduck.ClipboardWaitOptions{Text: nasduck.Ptr("go-clipboard")}))
	title, err := terminal.GetTitle()
	requireNoError(t, err)
	if title == nil || *title != "Go title" {
		t.Fatalf("title=%v", title)
	}
	clipboard, err := terminal.GetClipboard()
	requireNoError(t, err)
	if clipboard != "go-clipboard" {
		t.Fatalf("clipboard=%q", clipboard)
	}
	requireNoError(t, terminal.ExpectBellCount(1, nasduck.WaitOptions{}))
	events, err := terminal.GetBellEvents()
	requireNoError(t, err)
	if len(events) != 1 || events[0].Sequence != 1 {
		t.Fatalf("bell events=%+v", events)
	}
	count, err := terminal.GetBellCount()
	requireNoError(t, err)
	if count != 1 {
		t.Fatalf("bell count=%d", count)
	}
}

func TestScreenshotsAndSnapshots(t *testing.T) {
	terminal := newTerminal(t, nasduck.ClientOptions{})
	text, err := terminal.Screenshot("", nasduck.ScreenshotOptions{})
	requireNoError(t, err)
	if !strings.Contains(text, "go-ready") {
		t.Fatalf("screenshot omitted terminal text: %.100s", text)
	}
	directory := t.TempDir()
	path, err := terminal.Screenshot(filepath.Join(directory, "terminal.svg"), nasduck.ScreenshotOptions{})
	requireNoError(t, err)
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatal(statErr)
	}
	// Raster tests use bundled font coverage rather than host CJK fonts.
	requireNoError(t, terminal.Submit(nasduck.Ptr("raster")))
	requireNoError(t, terminal.GetByText("go-raster-ready", nasduck.TextSelectorOptions{}).Expect(nasduck.LocatorExpectOptions{}))
	transparent, err := terminal.Screenshot(filepath.Join(directory, "transparent.png"), nasduck.ScreenshotOptions{Transparent: true})
	requireNoError(t, err)
	if _, statErr := os.Stat(transparent); statErr != nil {
		t.Fatal(statErr)
	}
	data, err := os.ReadFile(transparent)
	requireNoError(t, err)
	baseline, err := png.DecodeConfig(bytes.NewReader(data))
	requireNoError(t, err)
	for index, options := range []nasduck.ScreenshotOptions{
		{Zoom: nasduck.Ptr(2.0), Transparent: true},
		{Zoom: nasduck.Ptr(2.0), Background: "#123456"},
	} {
		path, err := terminal.Screenshot(filepath.Join(directory, fmt.Sprintf("scaled-%d.png", index)), options)
		requireNoError(t, err)
		data, err := os.ReadFile(path)
		requireNoError(t, err)
		scaled, err := png.DecodeConfig(bytes.NewReader(data))
		requireNoError(t, err)
		if scaled.Width != 2*baseline.Width || scaled.Height != 2*baseline.Height {
			t.Fatalf("scaled screenshot=%dx%d, baseline=%dx%d", scaled.Width, scaled.Height, baseline.Width, baseline.Height)
		}
	}
	_, err = terminal.Screenshot(filepath.Join(directory, "conflict.png"), nasduck.ScreenshotOptions{Background: "#000000", Transparent: true})
	requireKind(t, err, nasduck.UsageError)
	snapshot, err := terminal.ExpectSnapshot("terminal", nasduck.SnapshotOptions{Cwd: directory, Update: true})
	requireNoError(t, err)
	if snapshot != nasduck.SnapshotWritten && snapshot != nasduck.SnapshotUpdated {
		t.Fatalf("snapshot=%s", snapshot)
	}
	_, err = terminal.ExpectSnapshot("terminal", nasduck.SnapshotOptions{Cwd: directory, IncludeStyle: true, Update: true})
	requireNoError(t, err)
	requireNoError(t, terminal.Submit(nasduck.Ptr("quit")))
	requireNoError(t, terminal.WaitExit(nasduck.WaitOptions{Timeout: nasduck.Ptr(10 * time.Second)}))
}

func TestRestartAndFailureAwareTraceClose(t *testing.T) {
	traceDirectory := t.TempDir()
	terminal := newClient(t, nasduck.ClientOptions{Trace: &nasduck.TraceOptions{Mode: nasduck.TraceOn, Directory: traceDirectory}})
	runTerminal(t, terminal, nasduck.SpawnOptions{})
	result, err := terminal.Restart(nasduck.RestartOptions{GracefulTimeout: nasduck.Ptr(2 * time.Second)})
	requireNoError(t, err)
	if result.Session != terminal.Session() {
		t.Fatalf("restart result=%+v", result)
	}
	requireNoError(t, terminal.GetByText("go-ready", nasduck.TextSelectorOptions{}).Expect(nasduck.LocatorExpectOptions{Timeout: nasduck.Ptr(10 * time.Second)}))
	requireNoError(t, terminal.Close(nasduck.CloseOptions{Failed: nasduck.Ptr(true)}))
	entries, err := os.ReadDir(traceDirectory)
	requireNoError(t, err)
	if len(entries) == 0 {
		t.Fatal("failure-aware close did not retain a trace")
	}
}

func TestScreenHistoryLimitValidation(t *testing.T) {
	_, err := nasduck.New("history", nasduck.ClientOptions{ScreenHistoryLimit: nasduck.Ptr(uint16(51))})
	requireKind(t, err, nasduck.UsageError)
}

func TestNamedHandlesReopen(t *testing.T) {
	terminal := newTerminal(t, nasduck.ClientOptions{})
	peer, err := nasduck.New(terminal.Session(), nasduck.ClientOptions{})
	requireNoError(t, err)
	names, err := nasduck.Sessions()
	requireNoError(t, err)
	if !slices.Contains(names, terminal.Session()) {
		t.Fatalf("missing session in %v", names)
	}
	requireNoError(t, terminal.Close())
	_, err = peer.Text(nasduck.TextOptions{})
	requireKind(t, err, nasduck.NoSessionError)
	runTerminal(t, terminal, nasduck.SpawnOptions{})
	requireNoError(t, peer.GetByText("go-ready", nasduck.TextSelectorOptions{}).Expect(nasduck.LocatorExpectOptions{}))
}

func TestCloseInterruptsPendingWait(t *testing.T) {
	t.Skip("Known upstream limitation accepted for this binding: named Close blocks behind pending waits; https://github.com/shiplaone2q/nasduck/issues/207")
	terminal := newTerminal(t, nasduck.ClientOptions{})
	verifyWaitInterruption(t, terminal, func() error { return terminal.Close() })
}

func TestCloseAllInterruptsPendingWait(t *testing.T) {
	terminal := newTerminal(t, nasduck.ClientOptions{})
	verifyWaitInterruption(t, terminal, nasduck.CloseAll)
}

func verifyWaitInterruption(t *testing.T, terminal *nasduck.Client, closeSession func() error) {
	t.Helper()
	started := make(chan struct{})
	waitResult := make(chan error, 1)
	go func() {
		close(started)
		waitResult <- terminal.GetByText("never-visible", nasduck.TextSelectorOptions{}).Wait(nasduck.LocatorWaitOptions{Timeout: nasduck.Ptr(5 * time.Second)})
	}()
	<-started
	select {
	case err := <-waitResult:
		t.Fatalf("wait returned before close: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	closeResult := make(chan error, 1)
	go func() { closeResult <- closeSession() }()
	select {
	case err := <-closeResult:
		requireNoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("close blocked behind the pending wait")
	}
	select {
	case err := <-waitResult:
		if err == nil {
			t.Fatal("interrupted wait unexpectedly succeeded")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("wait did not return after close")
	}
}

func TestTimeoutOptionsAndTypedFailures(t *testing.T) {
	_, err := nasduck.New("", nasduck.ClientOptions{Timeouts: nasduck.Timeouts{Text: nasduck.Ptr(-time.Nanosecond)}})
	requireKind(t, err, nasduck.UsageError)
	terminal := newTerminal(t, nasduck.ClientOptions{Timeouts: nasduck.Timeouts{Text: nasduck.Ptr(time.Duration(0))}, Artifacts: &nasduck.ArtifactOptions{Dir: t.TempDir(), OnFailure: nasduck.ArtifactAll}})
	requireNoError(t, terminal.Type("typed input"))
	err = terminal.GetByText("missing", nasduck.TextSelectorOptions{}).Expect(nasduck.LocatorExpectOptions{})
	requireKind(t, err, nasduck.AssertionError)
	var failure *nasduck.Error
	if !errors.As(err, &failure) || failure.Details == nil || failure.Artifact == nil || failure.Artifact.ScreenSVG == nil {
		t.Fatalf("missing diagnostic/artifact: %#v", failure)
	}
	if failure.Details.Operation != "locator.expect" {
		t.Fatalf("failure operation=%q", failure.Details.Operation)
	}
	_, err = terminal.Screenshot("", nasduck.ScreenshotOptions{Zoom: nasduck.Ptr(2.0)})
	requireKind(t, err, nasduck.UsageError)
	requireNoError(t, terminal.Close())
	runTerminal(t, terminal, nasduck.SpawnOptions{Timeouts: nasduck.Timeouts{Text: nasduck.Ptr(1500 * time.Microsecond), Idle: nasduck.Ptr(time.Duration(0))}})
	state, err := terminal.State()
	requireNoError(t, err)
	if state.Timeouts.Text != 2*time.Millisecond || state.Timeouts.Idle != 0 {
		t.Fatalf("effective timeouts=%+v", state.Timeouts)
	}
}

func TestPublicOperationNamesReachDiagnostics(t *testing.T) {
	traceDirectory := t.TempDir()
	terminal := newTerminal(t, nasduck.ClientOptions{
		Timeouts: nasduck.Timeouts{Text: nasduck.Ptr(time.Duration(0))},
		Trace:    &nasduck.TraceOptions{Mode: nasduck.TraceOn, Directory: traceDirectory},
	})
	requireNoError(t, terminal.Type("typed input"))
	err := terminal.GetByText("missing", nasduck.TextSelectorOptions{}).Expect(nasduck.LocatorExpectOptions{})
	requireKind(t, err, nasduck.AssertionError)
	var failure *nasduck.Error
	if !errors.As(err, &failure) || failure.Details == nil || failure.Details.Operation != "locator.expect" {
		t.Fatalf("failure details=%#v", failure)
	}
	requireNoError(t, terminal.Close(nasduck.CloseOptions{Failed: nasduck.Ptr(true)}))

	manifests, err := filepath.Glob(filepath.Join(traceDirectory, "*", "trace.json"))
	requireNoError(t, err)
	if len(manifests) != 1 {
		t.Fatalf("trace manifests=%v", manifests)
	}
	manifestData, err := os.ReadFile(manifests[0])
	requireNoError(t, err)
	var manifest struct {
		RecentOperations []struct {
			Name string `json:"name"`
		} `json:"recent_operations"`
	}
	requireNoError(t, json.Unmarshal(manifestData, &manifest))
	if !slices.ContainsFunc(manifest.RecentOperations, func(operation struct {
		Name string `json:"name"`
	}) bool {
		return operation.Name == "type"
	}) {
		t.Fatalf("recent operations=%+v", manifest.RecentOperations)
	}
}

func TestDefaultRecordingIsDisabled(t *testing.T) {
	terminal, err := nasduck.Ephemeral(t.Name(), nasduck.ClientOptions{})
	requireNoError(t, err)
	t.Cleanup(func() { _ = terminal.Close() })
	executable, err := os.Executable()
	requireNoError(t, err)
	result, err := terminal.Run(executable, []string{"-test.run=^TestTerminalProcess$", "--", "--tui-go-child"}, nasduck.SpawnOptions{WaitReady: nasduck.Ptr(false)})
	requireNoError(t, err)
	if result.Recording != "" {
		t.Fatalf("default recording path=%q", result.Recording)
	}
}

func TestRecordingAndCloseAll(t *testing.T) {
	directory := t.TempDir()
	terminal := newTerminal(t, nasduck.ClientOptions{Recording: &nasduck.AutomaticRecording{Mode: nasduck.RecordingAlways, Directory: directory}})
	recording, err := nasduck.Recording(terminal.Session())
	requireNoError(t, err)
	if !strings.Contains(recording, `"version":2`) {
		t.Fatal("missing automatic recording content")
	}
	requireNoError(t, terminal.StartRecording(filepath.Join(directory, "explicit.cast"), nasduck.RecordingOptions{Format: nasduck.Cast}))
	requireNoError(t, terminal.Submit(nasduck.Ptr("recorded")))
	requireNoError(t, terminal.GetByText("received:recorded", nasduck.TextSelectorOptions{}).Expect(nasduck.LocatorExpectOptions{}))
	explicit, err := terminal.StopRecording()
	requireNoError(t, err)
	if _, statErr := os.Stat(explicit); statErr != nil {
		t.Fatal(statErr)
	}
	requireNoError(t, nasduck.CloseAll())
	retained, err := nasduck.Recording(terminal.Session())
	requireNoError(t, err)
	if !strings.Contains(retained, "recorded") {
		t.Fatalf("retained recording omitted terminal output: %q", retained)
	}
}
