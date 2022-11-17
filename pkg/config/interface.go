package config

import (
	"github.com/spf13/viper"
)

// Create mock using:
// mockgen -source=pkg/config/interface.go -destination=pkg/config/mock/mock_config.go
type Interface interface {
	Init() error
	Set(key string, value interface{})
	SetDefault(key string, value interface{})
	AllSettings() map[string]interface{}
	IsSet(key string) bool
	Get(key string) interface{}
	GetBool(key string) bool
	GetInt(key string) int
	GetString(key string) string
	GetStringSlice(key string) []string
	GetStringMapString(key string) map[string]string
	GetStringMap(key string) map[string]interface{}
	UnmarshalKey(key string, rawVal interface{}, decoder ...viper.DecoderConfigOption) error
	ReadConfig(configFilePath string) error
}

const PACKAGR_PACKAGE_TYPE = "package_type"
const PACKAGR_SCM = "scm"
const PACKAGR_VERSION_BUMP_TYPE = "version_bump_type"
const PACKAGR_VERSION_METADATA_PATH = "version_metadata_path"
const PACKAGR_ADDL_VERSION_METADATA_PATHS = "addl_version_metadata_paths"
const PACKAGR_ENGINE_REPO_CONFIG_PATH = "engine_repo_config_path"
const PACKAGR_GENERIC_VERSION_TEMPLATE = "generic_version_template"
