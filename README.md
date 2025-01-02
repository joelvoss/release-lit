# release-lit

This repository contains a small CLI tool to automate the release process of
typical Node/Python/Go projects.
It is opinionated and assumes a certain project structure that I personally
use most often.

## Installation

To install the tool, download a precompiled binary of the latest release from
the [releases page](https://github.com/joelvoss/release-lit/releases).<br>
Available binaries:
- `darwin/amd64`
- `darwin/arm64`
- `linux/arm64`
- `linux/amd64`
- `windows/amd64`
- `windows/arm64`

After downloading the binary, make it executable and move it to a directory in
your `PATH` or place it inside your project directory.

```bash
$ chmod +x release-lit
$ mv release-lit /usr/local/bin
```

## Usage

To create a new release, run the following command:

```bash
$ ./release-lit
```

This will
- read all commits from the last release tag to the current `HEAD`
- determine the new version based on the commit messages
- write a detailed changelog to `./CHANGELOG.md`
- update the application version in the `package.json` / `pyproject.toml` /
  `Taskfile.sh` file (depending on the project type)
- create a new git commit and tag for the release

`release-lit` will not push the changes to the remote repository. You can do
this manually by running:

```bash
$ git push && git push --tags
```

To get a list of all available options, run:

```bash
$ ./release-lit --help
```

## Options

## `--cpath`

Alias: `-cp`

Path to the changelog file (default: `./CHANGELOG.md`)

## `--type`

Alias: `-t`

Project type (default: `node`). This can be one of `node`, `python`, or `go`.
It defines, which file to update with the new version.
- `node`: The `version` field in the `package.json` file will be updated using
  the following regular expression: `"version":\s*".*"`
- `python`: The `version` field in the `pyproject.toml` file will be updated using
  the following regular expression: `version\s*=\s*".*"`
- `go`: The `version` field in the `Taskfile.sh` file will be updated using
  the following regular expression: `VERSION=".*"`

## Development

To build the tool from source, you need to have Go installed on your machine.
After cloning the repository, run the following commands:

```bash
$ ./Taskfile.sh build
```

This will create a new binary in the `./bin` directory.

To format the code, execute:

```bash
$ ./Taskfile.sh format
```

To run the tests, execute:

```bash
$ ./Taskfile.sh test
```

To run the linter, execute:

```bash
$ ./Taskfile.sh lint
```

> Note: The linter requires `golangci-lint` to be installed on your machine.
> See [golangci-lint](https://golangci-lint.run/usage/install/) for installation
> instructions.
