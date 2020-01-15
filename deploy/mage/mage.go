//+build mage

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"gopkg.in/yaml.v2"
)

// Go is the namespace to group all Go commands
type Go mg.Namespace

// AWS is the namespace to group all AWS commands
type AWS mg.Namespace

// Pulumi is the namespace to group all Pulumi commands
type Pulumi mg.Namespace

// ConfigRoot is the root element in the Pulumi configuration file
type ConfigRoot struct {
	Config Config `yaml:"config"`
}

// Config contains all configuration items as key/value pairs, where the key is made up
// of <namespace>:<key>
type Config struct {
	AwsProfile string `yaml:"aws:profile"`
	AwsRegion  string `yaml:"aws:region"`
	AwsBucket  string `yaml:"aws:bucket"`
	TagsStage  string `yaml:"tags:stage"`
	TagsAuthor string `yaml:"tags:author"`
	TagsTeam   string `yaml:"tags:team"`
	MageTests  string `yaml:"mage:tests"`
}

var config = getConfig()

func getConfig() Config {
	// Get the stack name
	stack := os.Getenv("stack")

	// Read the Pulumi YAML file for the stack
	fileContent, err := ioutil.ReadFile(fmt.Sprintf("Pulumi.%s.yaml", stack))
	if err != nil {
		panic(err)
	}
	source := string(fileContent)

	// Find all configuration variables that have been set as
	// environment variables
	re := regexp.MustCompile(`\$\$.*\$\$`)
	vars := re.FindAllString(source, -1)

	// Replace the configuration variables with the actual values
	// of the environment variable
	for _, v := range vars {
		source = strings.ReplaceAll(source, v, getEnvVar(strings.ReplaceAll(v, "$$", "")))
	}

	// Unmarshal the YAML content to a proper struct and return
	// the config part of it
	d := &ConfigRoot{}
	yaml.Unmarshal([]byte(source), d)
	return d.Config
}

func getEnvVar(envvar string) string {
	b, found := os.LookupEnv(envvar)
	if !found {
		return "unknown"
	}
	return b
}

func getBasepath() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("cannot get base folder: %s", err))
	}
	return dir
}

func getGitVersion() string {
	v := run("git", "describe", "--tags", "--always", "--dirty=-dev")
	if len(v) == 0 {
		v = "dev"
	}
	return v
}

func run(name string, arg ...string) string {
	buf := &bytes.Buffer{}
	c := exec.Command(name, arg...)
	c.Stdout = buf
	c.Stderr = os.Stderr
	c.Run()
	return strings.TrimSuffix(buf.String(), "\n")
}

// Deps resolves and downloads dependencies to the current development module and then builds and installs them.
// Deps will rely on the Go environment variable GOPROXY (go env GOPROXY) to determine from where to obtain the
// sources for the build.
func (Go) Deps() error {
	goProxy := run("go", "env", "GOPROXY")
	fmt.Printf("Getting Go modules from %s", goProxy)
	return sh.Run("go", "get", "./...")
}

// 'Go test' automates testing the packages named by the import paths. go:test compiles and tests each of the
// packages listed on the command line. If a package test passes, go test prints only the final 'ok' summary
// line.
func (Go) Test() error {
	return sh.RunV("go", "test", "-cover", "./...")
}
