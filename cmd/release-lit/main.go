package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/joelvoss/release-lit/internal/changelog"
	"github.com/joelvoss/release-lit/internal/git"
	"github.com/joelvoss/release-lit/internal/golang"
	"github.com/joelvoss/release-lit/internal/node"
	"github.com/joelvoss/release-lit/internal/python"
	"github.com/joelvoss/release-lit/internal/semver"

	"github.com/urfave/cli/v2"
)

// NOTE(joel): This variable will be set by the linker during the build process.
var version string = "undefined"

func main() {
	app := &cli.App{
		Name:    "release-me",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "cpath",
				Aliases: []string{"cp"},
				Value:   "./CHANGELOG.md",
				Usage:   "path of the changelog file",
			},
			&cli.StringFlag{
				Name:    "type",
				Aliases: []string{"t"},
				Value:   "node",
				Usage:   "project type (node, python, go)",
			},
		},
		Action: func(cCtx *cli.Context) error {
			fmt.Println("INFO: Starting release process...")

			// NOTE(joel): Get git root. This conveniently also checks if the current
			// directory is a git repository.
			root, err := git.GetRoot(nil)
			if err != nil {
				return cli.Exit(err, 1)
			}

			// NOTE(joel): Get all tags sorted by version in descending order
			tags, err := git.GetTags(nil)
			if err != nil {
				return cli.Exit(err, 1)
			}

			// NOTE(joel): Get sha of latest tag to use as a reference point from
			// where new commits are analyzed.
			sha, err := git.GetTagHead(tags[0].Original, nil)
			if err != nil {
				return cli.Exit(err, 1)
			}

			// NOTE(joel): Get commits since latest tag.
			commits, err := git.GetCommits(sha, nil)
			if err != nil {
				return cli.Exit(err, 1)
			}

			// NOTE(joel): Get next release type.
			nextRelease := git.GetNextReleaseType(commits)

			// NOTE(joel): Bump version based on release type.
			// Create a copy of the first tag to use as a reference point.
			newVersion, err := semver.Bump(*tags[0], nextRelease)
			if err != nil {
				return cli.Exit(err, 1)
			}

			// NOTE(joel): Generate changelog and write to file
			var changelogFilePath string
			if path.IsAbs(cCtx.String("cpath")) {
				changelogFilePath = cCtx.String("cpath")
			} else {
				changelogFilePath = path.Join(root, cCtx.String("cpath"))
			}
			err = changelog.Generate(commits, newVersion, changelogFilePath)
			if err != nil {
				return cli.Exit(err, 1)
			}

			// NOTE(joel): Update version file based on project type.
			switch cCtx.String("type") {
			case "node":
				if err := node.UpdateVersion(newVersion, root); err != nil {
					return cli.Exit(err, 1)
				}
			case "python":
				if err := python.UpdateVersion(newVersion, root); err != nil {
					return cli.Exit(err, 1)
				}
			case "go":
				if err := golang.UpdateVersion(newVersion, root); err != nil {
					return cli.Exit(err, 1)
				}
			default:
				return cli.Exit("unsupported project type", 1)
			}

			// NOTE(joel): Create release commit + tag.
			err = git.CreateRelease(newVersion, nil)
			if err != nil {
				return cli.Exit(err, 1)
			}

			fmt.Println("INFO: Release created successfully. If applicable, don't forget to push the release commit + tag.")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
