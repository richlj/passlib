package pass

import (
	"fmt"
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

// listAll returns path information for all credentials
func listAll() (*Items, error) {
	filepaths, err := getFilepaths()
	if err != nil {
		return nil, err
	}
	var result Items
	for _, filepath := range filepaths {
		item, err := extractItem(filepath)
		if err != nil {
			return nil, err
		}
		if item != nil {
			result.Items = append(result.Items, item)
		}
	}
	return &result, nil
}

// match returns a bool as to whether the value in first pointer slice
// argument is contained within the value in the second pointer slice argument
func match(a, b *string) bool {
	if a == nil || b == nil {
		return false
	}
	result, err := regexp.MatchString(fmt.Sprintf(".*%s.*", *b), *a)
	if err != nil {
		return false
	}
	return result
}

// testMatch returns a bool as to whether the pointer receiver matches the
// supplied filter pointer slice
func (a *Item) testMatch(filter []*string) bool {
	if len(a.Path) == len(filter)-1 &&
		match(a.Credentials.Username, filter[len(filter)-1]) {
		for i := 0; i < len(filter)-1; i++ {
			if !match(a.Path[i], filter[i]) {
				return false
			}
		}
		return true
	}
	return false
}

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

// getFilepaths returns a list of filepaths for local pass credentials files
func getFilepaths() ([]*string, error) {
	dir, err := getDirectoryPath()
	if err != nil {
		return nil, err
	}
	a := map[string]bool{*dir: true}
	for {
		dirRemaining := false
		for key, value := range a {
			if value {
				dirRemaining = true
				contents, err := listDirContents(key)
				if err != nil {
					return nil, err
				}
				for key, value := range *contents {
					a[key] = value
				}
				delete(a, key)
			}
		}
		if !dirRemaining {
			var result []*string
			for key := range a {
				elem := key
				result = append(result, &elem)
			}
			return result, nil
		}
	}
}

// getCredentialPath returns the path for a set of credentials, as understood
// by the application
func (a *Item) getCredentialPath() string {
	var result string
	for _, dir := range a.Path {
		result = path.Join(result, *dir)
	}
	return path.Join(result, *a.Credentials.Username)
}
