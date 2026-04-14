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

You can override the organisation from the config file for a single invocation with the `--org` (or `-o`) flag:

```bash
revamp --org another-org list
```

This flag is available for all commands and takes precedence over the value in the config file.

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

### Summarise open Renovate PRs by branch name

```bash
revamp summary --branch
```

This runs:

```
gh search prs --owner <org> --author "app/renovate" --state open -L 100 \
  --json headRefName \
  --jq '.[].headRefName'
```

and prints each unique Renovate branch name together with how many times it appears, sorted by count descending:

```
     5 renovate/some-branch
     3 renovate/another-branch
```

### List repositories with open PRs from a specific branch

```bash
revamp summary --head renovate/foo
```

This runs:

```
gh search prs --owner <org> --head renovate/foo --state open \
  --json repository \
  --jq '.[].repository.nameWithOwner'
```

and prints each repository that has an open PR from that branch to standard output:

```
its-the-vibe/repo-one
its-the-vibe/repo-two
```

### Merge open Renovate PRs

```bash
revamp merge
```

Squash-merges all open Renovate PRs for the configured organisation and prints a summary:

```
Summary: 5 merged, 1 failed
```

#### Merge PRs from a specific branch

```bash
revamp merge --branch renovate/foo
```

Only merges open PRs whose head branch matches the given name.

#### Dry run

```bash
revamp merge --dry-run
```

Prints the PRs that would be merged without performing any merge operations:

```
Dry run: would merge 3 PR(s):
  https://github.com/its-the-vibe/repo-one/pull/42  chore(deps): update dependency foo to v2
  https://github.com/its-the-vibe/repo-two/pull/7   chore(deps): update dependency bar to v1.2.3
  https://github.com/its-the-vibe/repo-two/pull/8   chore(deps): update dependency baz to v3
```

#### Verbose output

```bash
revamp merge --verbose
```

Prints each PR URL and title as it is processed:

```
Merging: https://github.com/its-the-vibe/repo-one/pull/42  chore(deps): update dependency foo to v2
Merged:  https://github.com/its-the-vibe/repo-one/pull/42  chore(deps): update dependency foo to v2
...
Summary: 3 merged, 0 failed
```

All flags can be combined:

```bash
revamp merge --branch renovate/foo --dry-run --verbose
```

The number of PRs fetched defaults to 100 and can be changed with the `--limit` flag:

```bash
revamp merge --limit 500
```

To cap how many PRs are merged in a single run, use the `--max` flag:

```bash
revamp merge --max 10
```

This fetches up to `--limit` PRs but merges at most 10 of them, which is useful for rolling out updates incrementally.
