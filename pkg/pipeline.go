package pkg

import (
	"fmt"
	"github.com/packagrio/bumpr/pkg/config"
	"github.com/packagrio/bumpr/pkg/engine"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"os"
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

	sourceScm, err := scm.Create(p.Config.GetString(config.PACKAGR_SCM), p.Data)
	if err != nil {
		fmt.Printf("FATAL: %+v\n", err)
		os.Exit(1)
	}

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
	if err := p.Scm.Notify(); err != nil {
		fmt.Printf("FATAL: %+v\n", err)
		os.Exit(1)
	}

	return nil
}
