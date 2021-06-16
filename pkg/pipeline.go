package pkg

import (
	"fmt"
	"github.com/packagrio/bumpr/pkg/config"
	"github.com/packagrio/bumpr/pkg/engine"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"os"
	"path/filepath"
)

type Pipeline struct {
	Data   *pipeline.Data
	Config config.Interface
	Scm    scm.Interface
	Engine engine.Interface
}

func (p *Pipeline) Start(configData config.Interface) error {
	// Initialize Pipeline.
	p.Config = configData
	p.Data = new(pipeline.Data)

	//by default the current working directory is the local directory to execute in
	cwdPath, _ := os.Getwd()
	p.Data.GitLocalPath = cwdPath
	p.Data.GitParentPath = filepath.Dir(cwdPath)

	sourceScm, err := scm.Create(p.Config.GetString(config.PACKAGR_SCM), p.Data, p.Config, nil)
	if err != nil {
		fmt.Printf("FATAL: %+v\n", err)
		os.Exit(1)
	}
	p.Scm = sourceScm

	payload, err := p.Scm.RetrievePayload()
	if err != nil {
		return err
	}
	p.Data.GitHeadInfo = payload.Head
	p.Data.GitBaseInfo = payload.Base

	bumpEngine, err := engine.Create(
		p.Config.GetString(config.PACKAGR_PACKAGE_TYPE),
		p.Data, p.Config, sourceScm)
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

	//notify the SCM after the run is complete.
	if err := p.Scm.SetOutput("release_version", p.Data.ReleaseVersion); err != nil {
		fmt.Printf("FATAL: %+v\n", err)
		os.Exit(1)
	}
	fmt.Printf("version bumped to %s", p.Data.ReleaseVersion)
	return nil
}
