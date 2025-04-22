package engine

import (
	"bufio"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/bumpr/pkg/config"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"os"
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
	filePath := path.Join(gitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH))
	template := g.Config.GetString("generic_version_template")

	// Handle if the user wants to merge the version file and not overwrite it
	if g.Config.GetBool(config.PACKAGR_GENERIC_MERGE_VERSION_FILE) {
		versionContent, err := g.matchAsSingleLine(filePath, template)
		if err != nil {
			return err
		}
		g.CurrentMetadata.Version = versionContent
		return nil
	}

	match, err := g.matchAsMultiLine(filePath, template)
	if err != nil {
		return err
	}
	g.CurrentMetadata.Version = match
	return nil
}

// Matches the template with the entire file, useful for simple version files
func (g *engineGeneric) matchAsMultiLine(filePath string, template string) (string, error) {
	versionContent, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return g.getVersionFromString(string(versionContent), template)
}

// Only matches the version for a single line, used when you have a version on a single line within a complete multiline file
func (g *engineGeneric) matchAsSingleLine(filePath string, template string) (string, error) {
	fileReader, rerr := os.Open(filePath)
	scanner := bufio.NewScanner(fileReader)
	if rerr != nil {
		return "", rerr
	}

	for scanner.Scan() {
		readLine := scanner.Text()
		version, err := g.getVersionFromString(readLine, template)
		if err != nil {
			continue
		}
		return version, nil
	}
	return "", errors.EngineUnspecifiedError(fmt.Sprintf(
		"Was unable to find a version with the format `%s` in file %s", template, filePath,
	))
}

func (g *engineGeneric) getVersionFromString(versionContent string, template string) (string, error) {
	major := 0
	minor := 0
	patch := 0
	_, err := fmt.Sscanf(strings.TrimSpace(string(versionContent)), template, &major, &minor, &patch)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
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
	filePath := path.Join(gitLocalMetadataPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH))

	if g.Config.GetBool(config.PACKAGR_GENERIC_MERGE_VERSION_FILE) {
		completeVersionContent, err := os.ReadFile(filePath)
		if err == nil {
			oldVersion, err := semver.NewVersion(g.CurrentMetadata.Version)
			if err != nil {
				return err
			}
			oldVersionContent := fmt.Sprintf(template, oldVersion.Major(), oldVersion.Minor(), oldVersion.Patch())
			versionContent = strings.Replace(string(completeVersionContent), oldVersionContent, versionContent, 1)
		} else {
			println(fmt.Sprintf("Error reading file for merge `%s` with error: `%s`, creating new one ", filePath, err.Error()))
		}
	}

	return os.WriteFile(filePath, []byte(versionContent), 0644)
}
