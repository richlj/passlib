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
	fileSuffixREStr = "(.+)\\.gpg"
	separator       = "/"
	executableName  = "pass"
)

// listAll returns path information for all credentials
func listAll() ([]Item, error) {
	filepaths, err := getFilepaths()
	if err != nil {
		return nil, err
	}
	var result []Item
	for _, filepath := range filepaths {
		item, err := extractItem(filepath)
		if err != nil {
			return nil, err
		}
		if item != nil {
			result = append(result, item.value)
		}
	}
	return result, nil
}

// List returns a list of items that match the supplied filter
func List(filter string) ([]Item, error) {
	re, err := regexp.Compile(filter)
	if err != nil {
		return nil, err
	}
	all, err := listAll()
	if err != nil {
		return nil, err
	}
	var result []Item
	for _, item := range all {
		if re.MatchString(item.String()) {
			result = append(result, item)
		}
	}
	return result, nil
}

// Get takes arguments about the identity of a set of credentials. If there is
// exactly one result it returns a Details item, otherwise, or if the
// credential has no path, it returns an error
func Get(filter string) (*Item, error) {
	a, err := List(filter)
	if err != nil {
		return nil, err
	}
	switch n := len(a); {
	case n < 1:
		return nil, fmt.Errorf("credentials not found")
	case n > 1:
		return nil, fmt.Errorf("ambiguous query")
	}
	password, err := a[0].getPassword()
	if err != nil {
		return nil, err
	}
	if len(a[0].Path.Elements) < 1 {
		return nil, fmt.Errorf("credentials lack path")
	}
	return &Item{
		Path: a[0].Path,
		Credentials: Credentials{
			Username: a[0].Credentials.Username,
			Password: password.String(),
		},
	}, nil
}

// String returns the path of an item in string form
func (a *Item) String() string {
	var result string
	for _, item := range a.Path.Elements {
		if len(item.Element) > 0 {
			result = path.Join(result, item.Element)
		}
	}
	if len(a.Credentials.Username) > 0 {
		result = path.Join(result, a.Credentials.Username)
	}
	return result
}

func (a *Items) String() []string {
	var result []string
	for _, i := range a.Items {
		result = append(result, fmt.Sprintf("%s\n", i.String()))
	}
	return result
}

// getDirectoryPath returns the filepath of the overall directory used to
// store credentials by the application
func getDirectoryPath() (*directory, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	return &directory{filepath.Join(usr.HomeDir, mainDirectory)}, nil
}

// listDirContents returns a map of the names of the immediate contents of a
// directory, and whether or not that item is itself a directory
func listDirContents(dir string) (map[string]bool, error) {
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
	return result, nil
}

// extractItem takes a path of a pass file and converts it into an *Item
func extractItem(filePath string) (*itemWrapper, error) {
	d, err := getDirectoryPath()
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(path.Join(d.Directory, fileSuffixREStr))
	if err != nil {
		return nil, err
	}
	if c := re.FindStringSubmatch(filePath); len(c) == 2 {
		elements := strings.Split(c[1], separator)
		var path Path
		for i := 0; i < len(elements)-1; i++ {
			path.Elements = append(path.Elements, element{elements[i]})
		}
		return &itemWrapper{Item{
			Path: path,
			Credentials: Credentials{
				Username: elements[len(elements)-1],
			},
		}}, nil
	}
	return nil, nil
}

// extractDirectories takes a map of items and their directory status,
// returning a string of any directories within that
func extractDirectories(a map[string]bool) []string {
	var result []string
	for key, value := range a {
		if value {
			result = append(result, key)
		}
	}
	return result
}

// getFilepaths returns a list of filepaths for local pass credentials files
func getFilepaths() ([]string, error) {
	d, err := getDirectoryPath()
	if err != nil {
		return nil, err
	}
	for a := map[string]bool{d.Directory: true}; ; {
		dirRemaining := false
		for key, value := range a {
			if value {
				dirRemaining = true
				contents, err := listDirContents(key)
				if err != nil {
					return nil, err
				}
				for key, value := range contents {
					a[key] = value
				}
				delete(a, key)
			}
		}
		if !dirRemaining {
			var result []string
			for key := range a {
				result = append(result, key)
			}
			return result, nil
		}
	}
}

// getCredentialPath returns the path for a set of credentials, as understood
// by the application
func (a *Item) getCredentialPath() string {
	var result string
	for _, e := range a.Path.Elements {
		result = path.Join(result, e.Element)
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
	if p != nil && len(p.value) > 0 {
		return p.value
	}
	return ""
}

func (p *Path) String() string {
	var result string
	for _, e := range p.Elements {
		result = path.Join(result, e.Element)
	}
	return result
}
