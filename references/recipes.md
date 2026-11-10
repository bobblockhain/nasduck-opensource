# Recipes

[Back to the skill](../SKILL.md)

## Run a command

```sh
nasduck open
nasduck submit "npm test"
nasduck wait command
nasduck expect exit-code 0
```

## Drive an app

```sh
nasduck run my-app
nasduck expect text "Ready"
nasduck click text "Continue"
nasduck expect text "Done"
```

## Match relative text

```sh
nasduck click text "Save" --after-text "Settings" --match unique
```

```python
await (
    terminal
    .get_by_text("Settings")
    .get_by_text("Save", direction="after")
    .unique()
    .click()
)
```

```js
await terminal
  .getByText("Settings")
  .getByText("Save", { direction: "after" })
  .unique()
  .click();
```

## Wait for loading

```python
loading = terminal.get_by_text("Loading")
await loading.wait()
await loading.wait(state="hidden")
await terminal.get_by_text("Ready").expect()
```

## Click with modifiers

```sh
nasduck click text "Open" --button right --ctrl
```

```python
await terminal.get_by_text("Open").click(button="right", ctrl=True)
```

```js
await terminal.getByText("Open").click({ button: "right", ctrl: true });
```

## Wait for clipboard

```python
await terminal.wait_clipboard("copied")
```

```js
await terminal.waitClipboard(/copied/i);
```

## Keep failure artifacts

```sh
nasduck --failure-artifacts artifacts/failures \
  --failure-artifact-recording \
  expect text "Ready" --timeout 5000
```

```python
terminal = Nasduck(
    artifacts={
        "dir": "artifacts/failures",
        "on_failure": "all",
        "include_recording": True,
    },
    trace={"mode": "on-failure", "directory": "artifacts/traces"},
)
```

```js
const terminal = new Nasduck("test", {
  artifacts: {
    dir: "artifacts/failures",
    onFailure: "all",
    includeRecording: true,
  },
  trace: { mode: "on-failure", directory: "artifacts/traces" },
});
```

Users can open `failure.html` to view the failure. Agents should read `failure.md`, `failure.json`, and `timeline.json` instead.
