package app

import (
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

	testObject := &TestObject{"testing"}
	viperMock.On("UnmarshalKey",
		"config",
		mock.MatchedBy(func(input *TestObject) bool {
			return assert.Equal(t, input, testObject)
		})).
		Return(nil)
	tsViper = viperMock

	err := LoadConfig(testObject)
	assert.NoError(t, err)
	viperMock.AssertExpectations(t)
}

func TestLoadConfig_ReadConfigFailure(t *testing.T) {
	viperMock := new(mocks.Viper)
	viperMock.On("ReadInConfig").
		Return(assert.AnError)
	tsViper = viperMock

	testObject := &TestObject{"testing"}
	err := LoadConfig(testObject)
	assert.Error(t, err)
	viperMock.AssertExpectations(t)
}

func TestLoadConfig_UnmarshalKeyFailure(t *testing.T) {
	viperMock := new(mocks.Viper)
	viperMock.On("ReadInConfig").
		Return(nil)
	tsViper = viperMock

	testObject := &TestObject{"testing"}
	viperMock.On("UnmarshalKey",
		"config",
		mock.MatchedBy(func(input *TestObject) bool {
			return assert.Equal(t, input, testObject)
		})).
		Return(assert.AnError)

	err := LoadConfig(testObject)
	assert.Error(t, err)
	viperMock.AssertExpectations(t)
}

func TestSaveConfig_Success(t *testing.T) {
	ConfigPath, ConfigFile = "testConfigPath", "testConfigFile"
	wantPath := "testConfigPath/testConfigFile.json"
	testObject := &TestObject{"Testing"}

	viperMock := new(mocks.Viper)
	viperMock.On("Set",
		"config",
		testObject).Return()

	viperMock.On("WriteConfigAs",
		mock.MatchedBy(func(str string) bool {
			return assert.Equal(t, wantPath, str)
		})).Return(nil)
	tsViper = viperMock

	err := SaveConfig(testObject)
	assert.NoError(t, err)
	viperMock.AssertExpectations(t)
}

func TestSaveConfig_WriteConfigAsFailure(t *testing.T) {
	ConfigPath, ConfigFile = "testConfigPath", "testConfigFile"
	wantPath := "testConfigPath/testConfigFile.json"
	testObject := &TestObject{"Testing"}

	viperMock := new(mocks.Viper)
	viperMock.On("Set",
		"config",
		testObject).Return()

	viperMock.On("WriteConfigAs",
		mock.MatchedBy(func(str string) bool {
			return assert.Equal(t, wantPath, str)
		})).Return(assert.AnError)
	tsViper = viperMock

	err := SaveConfig(testObject)
	assert.Error(t, err)
	viperMock.AssertExpectations(t)
}
