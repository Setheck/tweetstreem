package app

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Setheck/tweetstreem/app/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestObject struct {
	Value string
}

func TestLoadConfig_Success(t *testing.T) {
	viperMock := new(mocks.Viper)
	viperMock.On("ReadInConfig").
		Return(nil)

	ts := &TweetStreem{}
	viperMock.On("UnmarshalKey",
		"config",
		mock.MatchedBy(func(input *TweetStreem) bool {
			return assert.Equal(t, input, ts)
		})).
		Return(nil)
	tsViper = viperMock

	err := ts.LoadConfig()
	assert.NoError(t, err)
	viperMock.AssertExpectations(t)
}

func TestLoadConfig_ReadConfigFailure(t *testing.T) {
	viperMock := new(mocks.Viper)
	viperMock.On("ReadInConfig").
		Return(assert.AnError)
	tsViper = viperMock

	ts := &TweetStreem{}
	err := ts.LoadConfig()
	assert.Error(t, err)
	viperMock.AssertExpectations(t)
}

func TestLoadConfig_UnmarshalKeyFailure(t *testing.T) {
	viperMock := new(mocks.Viper)
	viperMock.On("ReadInConfig").
		Return(nil)
	tsViper = viperMock

	ts := &TweetStreem{}
	viperMock.On("UnmarshalKey",
		"config",
		mock.MatchedBy(func(input *TweetStreem) bool {
			return assert.Equal(t, input, ts)
		})).
		Return(assert.AnError)

	err := ts.LoadConfig()
	assert.Error(t, err)
	viperMock.AssertExpectations(t)
}

func TestSaveConfig_Success(t *testing.T) {
	configPath, configFile = "testConfigPath", "testConfigFile"
	wantPath := fmt.Sprint(filepath.Join(configPath, configFile), ".", configFormat)

	ts := &TweetStreem{}
	viperMock := new(mocks.Viper)
	viperMock.On("Set",
		"config",
		ts).Return()

	viperMock.On("WriteConfigAs",
		mock.MatchedBy(func(str string) bool {
			return assert.Equal(t, wantPath, str)
		})).Return(nil)
	tsViper = viperMock

	err := ts.SaveConfig()
	assert.NoError(t, err)
	viperMock.AssertExpectations(t)
}

func TestSaveConfig_WriteConfigAsFailure(t *testing.T) {
	configPath, configFile = "testConfigPath", "testConfigFile"
	wantPath := fmt.Sprint(filepath.Join(configPath, configFile), ".", configFormat)

	ts := &TweetStreem{}
	viperMock := new(mocks.Viper)
	viperMock.On("Set",
		"config",
		ts).Return()

	viperMock.On("WriteConfigAs",
		mock.MatchedBy(func(str string) bool {
			return assert.Equal(t, wantPath, str)
		})).Return(assert.AnError)
	tsViper = viperMock

	err := ts.SaveConfig()
	assert.Error(t, err)
	viperMock.AssertExpectations(t)
}
