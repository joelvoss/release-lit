#!/bin/bash

set -e

NAME="release-lit"
VERSION="0.1.0"

# //////////////////////////////////////////////////////////////////////////////
# START tasks

build_dev() {
  go build \
    -ldflags="-X main.version=${VERSION}" \
    -o ./bin/${NAME} ./cmd/${NAME}

  cp ./bin/${NAME} ./e2e/golang
  cp ./bin/${NAME} ./e2e/node
  cp ./bin/${NAME} ./e2e/python

  rm -rf ./bin/${NAME}
}

build() {
  rm -rf ./bin
  # NOTE(joel): To get a list of all platforms: `go tool dist list`
  platforms=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/386"
    "linux/arm64"
    "linux/amd64"
    "windows/amd64"
    "windows/arm64"
  )

  for platform in "${platforms[@]}"
  do
      platform_split=(${platform//\// })
      GOOS=${platform_split[0]}
      GOARCH=${platform_split[1]}
      OUT=${NAME}'_'${VERSION}'_'${GOOS}'_'${GOARCH}
      if [ $GOOS = "windows" ]; then
          OUT+='.exe'
      fi

      echo "Building for $GOOS/$GOARCH..."
      env GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags="-X main.version=${VERSION}" \
        -o ./bin/${OUT} ./cmd/${NAME}
      if [ $? -ne 0 ]; then
          echo 'An error has occurred! Aborting the script execution...'
          exit 1
      fi
  done
}

format() {
  echo "Running 'go fmt'..."
  go fmt ./...
}

lint() {
  echo "Running 'golangci-lint run'..."
  golangci-lint run ./...
}

test() {
  echo "Running 'go test'..."
  go test ./internal/... -cover
}

validate() {
  lint
  test
}

help() {
  echo "Usage: $0 <command>"
  echo
  echo "Commands:"
  echo "  build      Build for production"
  echo "  format     Format code"
  echo "  lint       Lint code"
  echo "  test       Run tests"
  echo "  validate   Validate code"
  echo "  help       Show help"
  echo
}

# END tasks
# //////////////////////////////////////////////////////////////////////////////

${@:-help}
