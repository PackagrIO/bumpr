package engine

import (
	"github.com/golang/mock/gomock"
	"github.com/packagrio/bumpr/pkg/config"
	mock_config "github.com/packagrio/bumpr/pkg/config/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEngineBase_BumpVersion_Patch(t *testing.T) {

	//setup
	mockCtrl := gomock.NewController(t)
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString(config.PACKAGR_VERSION_BUMP_TYPE).MinTimes(1).Return("patch")
	eng := engineBase{
		Config: fakeConfig,
	}

	//test
	ver, err := eng.GenerateNextVersion("1.2.2")
	require.Nil(t, err)

	ver2, err := eng.GenerateNextVersion("1.0.0")
	require.Nil(t, err)

	//assert
	require.Equal(t, ver, "1.2.3", "should correctly do a patch bump")
	require.Equal(t, ver2, "1.0.1", "should correctly do a patch bump")
}

func TestEngineBase_BumpVersion_Minor(t *testing.T) {

	//setup
	mockCtrl := gomock.NewController(t)
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString(config.PACKAGR_VERSION_BUMP_TYPE).MinTimes(1).Return("minor")
	eng := engineBase{
		Config: fakeConfig,
	}

	//test
	ver, err := eng.GenerateNextVersion("1.2.2")
	require.Nil(t, err)

	//assert
	require.Equal(t, ver, "1.3.0", "should correctly do a patch bump")
}

func TestEngineBase_BumpVersion_Major(t *testing.T) {

	//setup
	mockCtrl := gomock.NewController(t)
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString(config.PACKAGR_VERSION_BUMP_TYPE).MinTimes(1).Return("major")
	eng := engineBase{
		Config: fakeConfig,
	}

	//test
	ver, err := eng.GenerateNextVersion("1.2.2")
	require.Nil(t, err)

	//assert
	require.Equal(t, ver, "2.0.0", "should correctly do a patch bump")
}

func TestEngineBase_BumpVersion_InvalidCurrentVersion(t *testing.T) {

	//setup
	mockCtrl := gomock.NewController(t)
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString(config.PACKAGR_VERSION_BUMP_TYPE).MinTimes(1).Return("patch")
	eng := engineBase{
		Config: fakeConfig,
	}

	//test
	nextV, err := eng.GenerateNextVersion("abcde")

	//assert
	require.Error(t, err, "should return an error if unparsable version")
	require.Empty(t, nextV, "should be empty next version")
}

func TestEngineBase_BumpVersion_WithVPrefix(t *testing.T) {

	//setup
	mockCtrl := gomock.NewController(t)
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString(config.PACKAGR_VERSION_BUMP_TYPE).MinTimes(1).Return("patch")
	eng := engineBase{
		Config: fakeConfig,
	}

	//test
	nextV, err := eng.GenerateNextVersion("v1.2.3")
	require.Nil(t, err)

	//assert
	require.Equal(t, nextV, "1.2.4", "should correctly do a patch bump")
}
