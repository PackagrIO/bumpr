package config_test

import (
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/bumpr/pkg/config"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestConfiguration_init_ShouldCorrectlyInitializeConfiguration(t *testing.T) {
	t.Parallel()

	//setup
	defer utils.UnsetEnv("PACKAGR_")()

	//test
	testConfig, _ := config.Create()

	//assert
	require.Equal(t, "generic", testConfig.GetString("package_type"), "should populate package_type with generic default")
	require.Equal(t, "default", testConfig.GetString("scm"), "should populate scm with default")
	require.Equal(t, "patch", testConfig.GetString("version_bump_type"), "should populate runner with default")
}

func TestConfiguration_init_EnvVariablesShouldLoadProperly(t *testing.T) {
	//setup
	os.Setenv("PACKAGR_VERSION_BUMP_TYPE", "major")

	//test
	testConfig, _ := config.Create()

	//assert
	require.Equal(t, "major", testConfig.GetString("version_bump_type"), "should populate Engine Version Bump Type from environmental variable")

	//teardown
	os.Unsetenv("PACKAGR_VERSION_BUMP_TYPE")
}
