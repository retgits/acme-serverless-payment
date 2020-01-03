//+build mage

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

type Go mg.Namespace
type Lambda mg.Namespace

type Config struct {
	Stage       string
	Author      string
	Team        string
	Project     string
	Version     string
	AWSS3Bucket string
	BasePath    string
	Tests       string
}

var c Config

func init() {
	source, err := ioutil.ReadFile("mageconfig.yaml")
	if err != nil {
		panic(err)
	}

	m := make(map[string]map[string]string)

	err = yaml.Unmarshal(source, &m)
	if err != nil {
		panic(err)
	}

	for key, val := range m["config"] {
		if strings.HasPrefix(val, "$$") && strings.HasSuffix(val, "$$") {
			m["config"][key] = getEnvVar(strings.ReplaceAll(val, "$$", ""))
		}
	}

	mapstructure.Decode(m["config"], &c)

	c.BasePath = getBasepath()
	c.Project = path.Base(c.BasePath)
	c.Version = getGitVersion()
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

// Build compiles the individual commands in the cmd folder, along with their dependencies. All built executables
// are stored in the 'bin' folder. Specifically for deployment to AWS Lambda, GOOS is set to linux and GOARCH is
// set to amd64.
func (Lambda) Build() error {
	env := make(map[string]string)
	env["GOOS"] = "linux"
	env["GOARCH"] = "amd64"
	files, err := ioutil.ReadDir(path.Join(c.BasePath, "cmd"))
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			if err := sh.RunWith(env, "go", "build", "-o", path.Join(c.BasePath, "bin", f.Name()), path.Join(c.BasePath, "cmd", f.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

// Clean remove all generated files.
func (Lambda) Clean() error {
	err := sh.Run("rm", "-rf", "bin")
	err = sh.Run("rm", "-rf", "packaged.yaml")
	return err
}

// Local executes functions in a Lambda-like environment locally. The input and events can be passed in by setting
// the 'tests' variable in the magefile with space separated entries in the form of <Function>/<input file>.
func (Lambda) Local() error {
	tests := strings.Split(c.Tests, " ")
	for _, test := range tests {
		t := strings.Split(test, "/")
		fmt.Printf("Running %s for %s", t[1], t[0])
		if err := sh.Run("sam", "local", "invoke", t[0], "-e", path.Join(c.BasePath, "test", t[1])); err != nil {
			return err
		}
	}
	return nil
}

// Destroy deletes the created stack, described in the template.yaml file. Once the call completes successfully, stack
// deletion starts.
func (Lambda) Destroy() error {
	return sh.RunV("aws", "cloudformation", "delete-stack", "--stack-name", fmt.Sprintf("%s-%s", c.Project, c.Stage))
}

// Deploy carries out the AWS CloudFormation commands 'package, 'deploy', and 'describe-stacks'. It first packages the
// local artifacts that template.yaml references. It uploads the files to the S3 bucket specified. The second step is
// to deploy the specified AWS CloudFormation template by creating and then executing a change set. The third, and final,
// command returns the description for the specified stack.
func (Lambda) Deploy() error {
	mg.Deps(Lambda.Clean, Lambda.Build)

	if err := sh.RunV("aws", "cloudformation", "package", "--template-file", "template.yaml", "--output-template-file", "packaged.yaml", "--s3-bucket", c.AWSS3Bucket); err != nil {
		return err
	}

	if err := sh.RunV("aws", "cloudformation", "deploy", "--template-file", "packaged.yaml", "--stack-name", fmt.Sprintf("%s-%s", c.Project, c.Stage), "--capabilities", "CAPABILITY_IAM", "--parameter-overrides", fmt.Sprintf("Version=%s", c.Version), fmt.Sprintf("Author=%s", c.Author), fmt.Sprintf("Team=%s", c.Team)); err != nil {
		return err
	}

	if err := sh.RunV("aws", "cloudformation", "describe-stacks", "--stack-name", fmt.Sprintf("%s-%s", c.Project, c.Stage), "--query", "'Stacks[].Outputs'"); err != nil {
		return err
	}

	return nil
}
