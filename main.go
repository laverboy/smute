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

type Args map[string]string

func main() {
	templateRepo := "github.com/laverboy/smute-templates"
	subDir := "basic"

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

	args, err := keysFileToArgs(filepath.Join(dir, subDir))
	CheckIfError(err)

	promptForArgValues(args)

	fmt.Println(args)
}

func promptForArgValues(a Args) {
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

func keysFileToArgs(templateDir string) (Args, error) {
	args := make(Args)
	file, err := os.Open(templateDir + "/keys.txt")
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
