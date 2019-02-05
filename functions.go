package pass

import (
	"io/ioutil"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	mainDirectory   = ".password-store"
	fileSuffixREStr = "([\\d\\D]{1,}.*)\\.gpg"
	separator       = "/"
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

// extractItem takes a path of a pass file and converts it into an *Item
func extractItem(filePath *string) (*Item, error) {
	dir, err := getDirectoryPath()
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(path.Join(*dir, fileSuffixREStr))
	if err != nil {
		return nil, err
	}
	if c := re.FindStringSubmatch(*filePath); len(c) == 2 {
		elements := strings.Split(c[1], separator)
		var path []*string
		for i := 0; i < len(elements)-1; i++ {
			path = append(path, &elements[i])
		}
		return &Item{
			Path: path,
			Credentials: &Credentials{
				Username: &elements[len(elements)-1],
			},
		}, nil
	}
	return nil, nil
}

// extractDirectories takes a map of items and their directory status,
// returning a string of any directories within that
func extractDirectories(a map[string]bool) []*string {
	var result []*string
	for key, value := range a {
		if value {
			result = append(result, &key)
		}
	}
	return result
}
