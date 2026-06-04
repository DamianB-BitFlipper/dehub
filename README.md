# dehub

A terminal UI for GitHub, built from the combined work of [dlvhdr/gh-dehub](https://github.com/dlvhdr/gh-dehub) and [pierrecomputer/pierre diffshub](https://github.com/pierrecomputer/pierre).

## Install

Make sure you have [GitHub CLI (`gh`)](https://cli.github.com/), [Go](https://go.dev/doc/install), and [Bun](https://bun.sh/docs/installation) installed, then run:

```bash
git clone git@github.com:DamianB-BitFlipper/dehub.git
cd dehub
make local-install
```

This builds the bundled diff viewer, builds `dehub`, and installs the current checkout as the `dehub` GitHub CLI extension.

## Configuration

Configuration is largely the same as gh-dash. See the [gh-dash getting started guide](https://www.gh-dash.dev/getting-started/) for configuration details.

## Common Shortcuts

- `?` opens help.
- `↑` / `↓` move through items.
- `[` / `]` switch sections.
- `{` / `}` switch views.
- `Ctrl+←` focuses the main pane and `Ctrl+→` focuses the preview pane.
- `/` searches and `s` filters rows.
- `o` opens the selected item in GitHub.
- `y` copies the selected item number and `Y` copies its URL.
- `R` refreshes.
- `Q` or `Ctrl+C` quits.

## How To Run

```bash
gh dehub
```
