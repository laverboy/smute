package app

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type PlaceholdersWithValues map[string]string

func CLI(args []string) int {
	var app app
	err := app.fromArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "usage error: %v\n", err)
		fmt.Println("usage: smute github_repo [repo_sub_directory] output_directory")
		return 2
	}
	if err = app.run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		return 1
	}
	return 0
}

type app struct {
	templateRepo string
	repoDir      string
	outputDir    string
}

func (a *app) fromArgs(args []string) error {
	if len(args) < 3 {
		return errors.New("not enough arguments")
	}

	if !validRepo(args[1]) {
		return fmt.Errorf("%s does not look like a valid github repo", args[1])
	}

	a.templateRepo = args[1]
	if len(args) == 4 {
		a.repoDir = args[2]
		a.outputDir = args[3]
	} else {
		a.outputDir = args[2]
	}

	return nil
}

func (a *app) run() error {
	tmpDir, err := ioutil.TempDir("", "smute-clone")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	_, err = git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL: "https://" + a.templateRepo,
	})
	if err != nil {
		return err
	}

	if !validTemplate(tmpDir, a.repoDir) {
		return errors.New("does not look like a valid template directory")
	}

	placeholders, err := loadPlaceholders(filepath.Join(tmpDir, a.repoDir))
	if err != nil {
		return err
	}

	promptForPlaceholderValues(placeholders)

	err = filepath.Walk(filepath.Join(tmpDir, a.repoDir), findNReplace(placeholders, filepath.Join(tmpDir, a.repoDir), a.outputDir))
	if err != nil {
		return err
	}

	return nil
}

func validRepo(repo string) bool {
	return !strings.HasPrefix(repo, "https://github.com") || !strings.HasPrefix(repo, "github.com")
}

func findNReplace(placeholders PlaceholdersWithValues, tempDir, outputDir string) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		newOutputPath := strings.ReplaceAll(path, tempDir, outputDir)

		// If directory, create new, then move on
		if info.IsDir() {
			err = os.MkdirAll(newOutputPath, os.ModePerm)
			if err != nil {
				return err
			}
			return nil
		}

		// skip files
		filesToSkip := []string{"*.DS_Store", "keys.txt"}
		for _, skipFile := range filesToSkip {
			matched, err := filepath.Match(skipFile, info.Name())
			if err != nil {
				return err
			}
			if matched {
				return nil
			}
		}

		read, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		newContents := string(read)
		for k, v := range placeholders {
			placeholder := "<<" + k + ">>"
			newContents = strings.Replace(newContents, placeholder, v, -1)
		}

		if err := ioutil.WriteFile(newOutputPath, []byte(newContents), info.Mode()); err != nil {
			return err
		}

		return nil
	}
}

func promptForPlaceholderValues(a PlaceholdersWithValues) {
	reader := bufio.NewReader(os.Stdin)
	for key := range a {
		fmt.Printf("%s: ", key)
		text, _ := reader.ReadString('\n')
		a[key] = strings.TrimSuffix(text, "\n")
	}
}

func validTemplate(dir, subdir string) bool {
	return fileExists(filepath.Join(dir, subdir, "keys.txt"))
}

func loadPlaceholders(templateDir string) (PlaceholdersWithValues, error) {
	args := make(PlaceholdersWithValues)
	file, err := os.Open(filepath.Join(templateDir, "keys.txt"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		args[scanner.Text()] = ""
	}

	err = scanner.Err()
	return args, err
}
