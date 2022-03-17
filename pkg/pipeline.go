package pkg

import (
	"errors"
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/bumpr/pkg/config"
	"github.com/packagrio/bumpr/pkg/engine"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"log"
	"net/http"
	"os"
	"path"
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

	//Parse Repo config if present.
	if err := p.ParseRepoConfig(); err != nil {
		return err
	}

	sourceScm, err := scm.Create(p.Config.GetString(config.PACKAGR_SCM), p.Data, p.Config, &http.Client{})
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

func (p *Pipeline) ParseRepoConfig() error {
	log.Println("parse_repo_config")
	// update the config with repo config file options
	repoConfig := path.Join(p.Data.GitLocalPath, p.Config.GetString(config.PACKAGR_ENGINE_REPO_CONFIG_PATH))
	if utils.FileExists(repoConfig) {
		log.Println("Found config file in working dir!")
		if err := p.Config.ReadConfig(repoConfig); err != nil {
			return errors.New("An error occured while parsing repository packagr.yml file")
		}
	} else {
		log.Println("No repo packagr.yml file found, using existing config.")
	}

	return nil
}
