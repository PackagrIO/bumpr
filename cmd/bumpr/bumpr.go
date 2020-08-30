package main

import (
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/bumpr/pkg/config"
	"github.com/packagrio/bumpr/pkg/engine"
	"github.com/packagrio/bumpr/pkg/version"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/urfave/cli"
	"log"
	"os"
	"time"
)

var goos string
var goarch string

func main() {
	app := &cli.App{
		Name:     "bumpr",
		Usage:    "Language agnostic tool to bump version files using SemVer",
		Version:  version.VERSION,
		Compiled: time.Now(),
		Authors: []cli.Author{
			cli.Author{
				Name:  "Jason Kulatunga",
				Email: "jason@thesparktree.com",
			},
		},
		Before: func(c *cli.Context) error {

			packagrUrl := "github.com/packagrio/bumpr"

			versionInfo := fmt.Sprintf("%s.%s-%s", goos, goarch, version.VERSION)

			subtitle := packagrUrl + utils.LeftPad2Len(versionInfo, " ", 53-len(packagrUrl))

			fmt.Fprintf(c.App.Writer, fmt.Sprintf(utils.StripIndent(
				`
			  ___   __   ____  ____  _  _  __    ____  ___  ____
			 / __) / _\ (  _ \/ ___)/ )( \(  )  (  __)/ __)(    \
			( (__ /    \ ) __/\___ \) \/ (/ (_/\ ) _)( (__  ) D (
			 \___)\_/\_/(__)  (____/\____/\____/(____)\___)(____/
			%s

			`), subtitle))
			return nil
		},

		Commands: []cli.Command{
			{
				Name:  "start",
				Usage: "Start a new CapsuleCD package pipeline",
				Action: func(c *cli.Context) error {

					configuration, _ := config.Create()
					configuration.Set(config.PACKAGR_SCM, c.String("scm"))
					configuration.Set(config.PACKAGR_PACKAGE_TYPE, c.String("package_type"))
					//config.Set("dry_run", c.String("dry_run"))

					fmt.Println("package type:", configuration.GetString(config.PACKAGR_PACKAGE_TYPE))
					fmt.Println("scm:", configuration.GetString(config.PACKAGR_SCM))

					pipelineData := new(pipeline.Data)
					sourceScm, err := scm.Create(configuration.GetString(config.PACKAGR_SCM), pipelineData)
					if err != nil {
						fmt.Printf("FATAL: %+v\n", err)
						os.Exit(1)
					}

					bumpEngine, err := engine.Create(
						configuration.GetString(config.PACKAGR_PACKAGE_TYPE),
						pipelineData, configuration, sourceScm)
					if err != nil {
						fmt.Printf("FATAL: %+v\n", err)
						os.Exit(1)
					}

					if err := bumpEngine.ValidateTools(); err != nil {
						fmt.Printf("FATAL: %+v\n", err)
						os.Exit(1)
					}

					if err := bumpEngine.BumpVersion(); err != nil {
						fmt.Printf("FATAL: %+v\n", err)
						os.Exit(1)
					}

					return nil
				},

				Flags: []cli.Flag{
					//TODO: currently not applicable
					//&cli.StringFlag{
					//	Name:  "runner",
					//	Value: "default", // can be :none, :circleci or :shippable (check the readme for why other hosted providers arn't supported.)
					//	Usage: "The cloud CI runner that is running this PR. (Used to determine the Environmental Variables to parse)",
					//},

					&cli.StringFlag{
						Name:  "scm",
						Value: "default",
						Usage: "The scm for the code, for setting additional SCM specific metadata",
					},

					&cli.StringFlag{
						Name:  "package_type",
						Value: "generic",
						Usage: "The type of package being built.",
					},

					&cli.BoolFlag{
						Name:  "dry_run",
						Usage: "When dry run is enabled, no data is written to file system",
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}
