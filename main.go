package main

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

func main() {
	templateRepo := "github.com/laverboy/smute-templates"
	subDir := "basic"
	outputDir := "/tmp/basic"

	dir, err := ioutil.TempDir("", "smute-clone")
	CheckIfError(err)

	defer os.RemoveAll(dir)

	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL: "https://" + templateRepo,
	})
	CheckIfError(err)

	if !validTemplate(dir, subDir) {
		CheckIfError(errors.New("does not look like a valid template"))
	}

	placeholders, err := loadPlaceholders(filepath.Join(dir, subDir))
	CheckIfError(err)

	promptForPlaceholderValues(placeholders)

	err = filepath.Walk(filepath.Join(dir, subDir), findNReplace(placeholders, filepath.Join(dir, subDir), outputDir))
	CheckIfError(err)
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
