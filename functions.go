package pass

import (
	"io/ioutil"
	"os/user"
	"path/filepath"
)

const (
	mainDirectory = ".password-store"
)

// getDirectoryPath returns the filepath of the overall directory used to
// store credentials by the application
func getDirectoryPath() (*string, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(usr.HomeDir, mainDirectory)
	return &dir, nil
}

// listDirContents returns a map of the names of the immediate contents of a
// directory, and whether or not that item is itself a directory
func listDirContents(dir string) (*map[string]bool, error) {
	result := make(map[string]bool)
	contents, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, item := range contents {
		path := filepath.Join(dir, item.Name())
		if item.IsDir() {
			result[path] = true
		} else {
			result[path] = false
		}
	}
	return &result, nil
}
