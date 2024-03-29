package engine

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/bumpr/pkg/config"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"io/ioutil"
	"path"
	"strings"
)

type engineGeneric struct {
	engineBase

	Scm             scm.Interface //Interface
	CurrentMetadata *metadata.GenericMetadata
	NextMetadata    *metadata.GenericMetadata
}

func (g *engineGeneric) Init(pipelineData *pipeline.Data, configData config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = configData
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(metadata.GenericMetadata)
	g.NextMetadata = new(metadata.GenericMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault(config.PACKAGR_GENERIC_VERSION_TEMPLATE, `version := "%d.%d.%d"`)
	g.Config.SetDefault(config.PACKAGR_VERSION_METADATA_PATH, "VERSION")
	return nil
}

func (g *engineGeneric) GetCurrentMetadata() interface{} {
	return g.CurrentMetadata
}
func (g *engineGeneric) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *engineGeneric) ValidateTools() error {
	return nil
}

func (g *engineGeneric) BumpVersion() error {
	//validate that the chef metadata.rb file exists

	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH))) {
		return errors.EngineBuildPackageInvalid(fmt.Sprintf("version file (%s) is required for metadata storage via generic engine", g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)))
	}

	// bump up the go package version
	if merr := g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath); merr != nil {
		return merr
	}

	if perr := g.populateNextMetadata(); perr != nil {
		return perr
	}

	if nerr := g.SetVersion(path.Join(g.PipelineData.GitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)), g.NextMetadata.Version); nerr != nil {
		return nerr
	}

	return nil
}

func (g *engineGeneric) SetVersion(versionMetadataPath string, nextVersion string) error {
	return g.writeNextMetadata(versionMetadataPath, nextVersion)
}

//Helpers
func (g *engineGeneric) retrieveCurrentMetadata(gitLocalPath string) error {
	//read VERSION file.
	versionContent, rerr := ioutil.ReadFile(path.Join(gitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)))
	if rerr != nil {
		return rerr
	}

	major := 0
	minor := 0
	patch := 0
	template := g.Config.GetString("generic_version_template")
	fmt.Sscanf(strings.TrimSpace(string(versionContent)), template, &major, &minor, &patch)

	g.CurrentMetadata.Version = fmt.Sprintf("%d.%d.%d", major, minor, patch)
	return nil
}

func (g *engineGeneric) populateNextMetadata() error {

	nextVersion, err := g.GenerateNextVersion(g.CurrentMetadata.Version)
	if err != nil {
		return err
	}

	g.NextMetadata.Version = nextVersion
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

func (g *engineGeneric) writeNextMetadata(gitLocalMetadataPath string, nextVersion string) error {

	v, nerr := semver.NewVersion(nextVersion)
	if nerr != nil {
		return nerr
	}

	template := g.Config.GetString(config.PACKAGR_GENERIC_VERSION_TEMPLATE)
	versionContent := fmt.Sprintf(template, v.Major(), v.Minor(), v.Patch())

	return ioutil.WriteFile(gitLocalMetadataPath, []byte(versionContent), 0644)
}
