# revamp

A command line tool for managing Renovate PRs.

## Prerequisites

- [Go 1.24+](https://golang.org/dl/)
- [GitHub CLI (`gh`)](https://cli.github.com/) — authenticated with an account that can search PRs in your organisation

## Installation

```bash
go install github.com/its-the-vibe/revamp@latest
```

Or build from source:

```bash
go build -o revamp .
```

## Configuration

Copy the example config and fill in your values:

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml`:

```yaml
# The GitHub organisation to search for Renovate PRs
org: "my-org"
```

`config.yaml` is gitignored so your real configuration is never committed.

You can also point to a different config file with the `--config` flag:

```bash
revamp --config /path/to/my-config.yaml list
```

## Usage

### List open Renovate PRs

```bash
revamp list
```

This runs:

```
gh search prs --owner <org> --author "app/renovate" --state open -L 100 \
  --json title,repository \
  --jq '.[] | "\(.repository.nameWithOwner) | \(.title)"'
```

and prints the results sorted alphabetically to standard output.

### Summarise open Renovate PRs by title

```bash
revamp summary --title
```

This runs:

```
gh search prs --owner <org> --author "app/renovate" --state open -L 100 \
  --json title \
  --jq '.[].title'
```

and prints each unique PR title together with how many times it appears, sorted by count descending:

```
     5 chore(deps): update dependency foo to v2
     3 chore(deps): update dependency bar to v1.2.3
```

### Summarise open Renovate PRs by repository

```bash
revamp summary --repo
```

This runs:

```
gh search prs --owner <org> --author "app/renovate" --state open -L 100 \
  --json repository \
  --jq '.[].repository.nameWithOwner'
```

and prints each repository together with its open Renovate PR count, sorted by count descending:

```
     7 its-the-vibe/repo-one
     2 its-the-vibe/repo-two
```

Both flags can be combined in a single invocation:

```bash
revamp summary --title --repo
```

The number of PRs fetched defaults to 100 and can be changed with the `--limit` flag:

```bash
revamp summary --title --limit 500
```
