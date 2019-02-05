package pass

import (
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
