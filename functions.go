package pass

import (
	"fmt"
	"io/ioutil"
	"os/exec"
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
	executableName  = "pass"
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

// List returns a list of items that match the supplied filter
func List(filter string) (*Items, error) {
	re, err := regexp.Compile(filter)
	if err != nil {
		return nil, err
	}
	all, err := listAll()
	if err != nil {
		return nil, err
	}
	var result Items
	for _, item := range all.Items {
		result.appendIfValid(item, re)
	}
	return &result, nil
}

// Get takes arguments about the identity of a set of credentials. If there is
// exactly one result it returns a Details item, otherwise, or if the
// credential has no path, it returns an error
func Get(filter string) (*Item, error) {
	a, err := List(filter)
	if err != nil {
		return nil, err
	}
	if matches := len(a.Items); matches == 0 {
		return nil, fmt.Errorf("credentials not found")
	} else if matches > 1 {
		return nil, fmt.Errorf("ambiguous query")
	}
	password, err := a.Items[0].getPassword()
	if err != nil {
		return nil, err
	}
	if len(a.Items[0].Path) == 0 {
		return nil, fmt.Errorf("credentials lack path")
	}
	return &Item{
		Path: a.Items[0].Path,
		Credentials: &Credentials{
			Username: a.Items[0].Credentials.Username,
			Password: password.String(),
		},
	}, nil
}

// String .
func (a *Item) String() string {
	var result string
	for _, item := range a.Path {
		if item != nil {
			result = path.Join(result, *item)
		}
	}
	if len(a.Credentials.Username) > 0 {
		result = path.Join(result, a.Credentials.Username)
	}
	return result
}

// appendIfValid adds the supplied Item if it matches the regex
func (a *Items) appendIfValid(item *Item, re *regexp.Regexp) {
	if re.MatchString(item.String()) {
		a.Items = append(a.Items, item)
	}
	return
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
				Username: elements[len(elements)-1],
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
	return path.Join(result, a.Credentials.Username)
}

// getPassword retrieves a password for an item
func (a *Item) getPassword() (*password, error) {
	cmd := exec.Command(executableName, a.getCredentialPath())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return &password{string(output[:len(output)-1])}, nil
}

func (p *password) String() string {
	if p != nil && len(p.Password) > 0 {
		return p.Password
	}
	return ""
}
