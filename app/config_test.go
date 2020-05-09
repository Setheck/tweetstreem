package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestApp_Config(t *testing.T) {
	t.SkipNow()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	ConfigPath = filepath.Join(dir, "testData")
	ConfigFile = fmt.Sprint(t.Name(), "_config")

	ts := NewTweetStreem(context.TODO())
	loadConfig(ts)
	// TODO:(smt) blank these out for safety. (maybe move twitter to anther package)
	ts.TwitterConfiguration.UserToken = ""
	ts.TwitterConfiguration.UserSecret = ""
	saveConfig(ts)
}
