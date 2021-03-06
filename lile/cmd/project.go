package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
)

type project struct {
	Name         string
	RelativeName string
	ProjectDir   string
	RelDir       string
	Folder       folder
}

func newProject(path, relativeName string) project {
	name := lastFromSplit(path, string(os.PathSeparator))
	relDir := projectBase(path)

	f := folder{Name: name, AbsPath: path}

	s := f.addFolder("server")
	s.addFile("server.go", "server.tmpl")
	s.addFile("server_test.go", "server_test.tmpl")

	subs := f.addFolder("subscribers")
	subs.addFile("subscribers.go", "subscribers.tmpl")

	cmd := f.addFolder(name)
	cmd.addFile("main.go", "cmd_main.tmpl")

	cmds := cmd.addFolder("cmd")
	cmds.addFile("root.go", "cmd_root.tmpl")
	cmds.addFile("up.go", "cmd_up.tmpl")

	f.addFile(name+".proto", "proto.tmpl")
	f.addFile("client.go", "client.tmpl")
	f.addFile("Makefile", "Makefile.tmpl")
	f.addFile("Dockerfile", "Dockerfile.tmpl")
	f.addFile(".gitignore", "gitignore.tmpl")

	return project{
		Name:         name,
		RelativeName: relativeName,
		RelDir:       relDir,
		ProjectDir:   path,
		Folder:       f,
	}
}

func (p project) write(templatePath string) error {
	err := os.MkdirAll(p.ProjectDir, os.ModePerm)
	if err != nil {
		return err
	}

	return p.Folder.render(templatePath, p)
}

// CamelCaseName returns a CamelCased name of the service
func (p project) CamelCaseName() string {
	return strcase.ToCamel(p.Name)
}

// SnakeCaseName returns a snake_cased_type name of the service
func (p project) SnakeCaseName() string {
	return strings.Replace(strcase.ToSnake(p.Name), "-", "_", -1)
}

// DNSName returns a snake-cased-type name that be used as a URL or packageName
func (p project) DNSName() string {
	return strings.Replace(strcase.ToSnake(p.Name), "_", "-", -1)
}

// Copied and re-worked from
// https://github.com/spf13/cobra/blob/master/cobra/cmd/helpers.go
func projectPath(inputPath string) string {
	// if no path is provided... assume CWD.
	if inputPath == "" {
		x, err := os.Getwd()
		if err != nil {
			er(err)
		}

		return x
	}

	var projectPath string
	var projectBase string
	srcPath := srcPath()

	// if provided, inspect for logical locations
	if strings.ContainsRune(inputPath, os.PathSeparator) {
		if filepath.IsAbs(inputPath) || filepath.HasPrefix(inputPath, string(os.PathSeparator)) {
			// if Absolute, use it
			projectPath = filepath.Clean(inputPath)
			return projectPath
		}
		// If not absolute but contains slashes,
		// assuming it means create it from $GOPATH
		count := strings.Count(inputPath, string(os.PathSeparator))

		switch count {
		// If only one directory deep, assume "github.com"
		case 1:
			projectPath = filepath.Join(srcPath, "github.com", inputPath)
			return projectPath
		case 2:
			projectPath = filepath.Join(srcPath, inputPath)
			return projectPath
		default:
			er(errors.New("Unknown directory"))
		}
	}

	// hardest case.. just a word.
	if projectBase == "" {
		x, err := os.Getwd()
		if err == nil {
			projectPath = filepath.Join(x, inputPath)
			return projectPath
		}
		er(err)
	}

	projectPath = filepath.Join(srcPath, projectBase, inputPath)
	return projectPath
}

func projectBase(absPath string) string {
	rel, err := filepath.Rel(srcPath(), absPath)
	if err != nil {
		return filepath.ToSlash(absPath)
	}
	return filepath.ToSlash(rel)
}

func lastFromSplit(input, split string) string {
	rel := strings.Split(input, split)
	return rel[len(rel)-1]
}

func srcPath() string {
	return filepath.Join(gopath, "src") + string(os.PathSeparator)
}
