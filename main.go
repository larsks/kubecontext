package main

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		Context     string
		Namespace   string
		Environment map[string]string
	}
)

// Discover .kubecontext files starting in the current directory
// and iterating over parents directories until we reach "/".
func findKubecontext() ([]string, error) {
	var contexts []string

	for {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		if cwd == "/" {
			break
		}

		log.Debugf("looking for .kubecontext in %s", cwd)

		if _, err := os.Stat("./.kubecontext"); err == nil {
			kubecontext := path.Join(cwd, ".kubecontext")
			log.Debugf("found %s", kubecontext)
			contexts = append(contexts, kubecontext)
		}

		os.Chdir("..")
	}

	return contexts, nil
}

// Apply settings from the specified .kubecontext file.
func processKubecontext(context string) error {
	var config Config

	log.Infof("processing configuration from %s", context)

	data, err := ioutil.ReadFile(context)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	if config.Context != "" {
		log.Infof("setting context to %s", config.Context)
		cmd := exec.Command("kubectl", "config", "use-context", config.Context)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	if config.Namespace != "" {
		log.Infof("setting namespace to %s", config.Namespace)
		cmd := exec.Command(
			"kubectl", "config", "set-context", "--current",
			"--namespace", config.Namespace,
		)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	if config.Environment != nil {
		for name, value := range config.Environment {
			log.Infof("setting environment variable %s to %s", name, value)
			os.Setenv(name, value)
		}
	}

	return nil
}

// Configure log level based on the K_LOGLEVEL environment variable.
// Valid values are "debug" and "info"; anything else results in
// logging messages at WARN and above.
func configureLogging() {
	var loglevel log.Level

	switch strings.ToLower(os.Getenv("K_LOGLEVEL")) {
	case "debug":
		loglevel = log.DebugLevel
	case "info":
		loglevel = log.InfoLevel
	default:
		loglevel = log.WarnLevel
	}

	log.SetLevel(loglevel)

}

// Write kubectl configuration to a temporary file. We manipulate the
// temporary file rather than modifying the real kubeconfig.
func generateKubeconfig(kubeconfig *os.File) error {
	log.Debugf("writing temporary kubeconfig to %s", kubeconfig.Name())

	cmd := exec.Command("kubectl", "config", "view", "--flatten", "--merge")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go io.Copy(kubeconfig, stdout)
	cmd.Wait()

	os.Setenv("KUBECONFIG", kubeconfig.Name())
	return nil
}

func Kubecontext() {
	var commandName string

	tmpfile, err := ioutil.TempFile("", "kubeconfig")
	if err != nil {
		panic(err)
	}
	defer func() {
		log.Debugf("removing %s", tmpfile.Name())
		os.Remove(tmpfile.Name())
	}()

	if err := generateKubeconfig(tmpfile); err != nil {
		panic(err)
	}

	kubecontexts, err := findKubecontext()
	if err != nil {
		panic(err)
	}

	// If we discovered one or more .kubecontext files, iterate over them
	// in reverse order, applying the configuration from each one.
	if kubecontexts != nil {
		for i := range kubecontexts {
			current := kubecontexts[len(kubecontexts)-i-1]
			if err := processKubecontext(current); err != nil {
				panic(err)
			}
		}
	}

	// If you have a project that requires `oc` instead of `kubectl`,
	// you can set `K_COMMANDNAME` in your environment (or in the
	// `environment` section of your `.kubecontext` file.
	if value := os.Getenv("K_COMMANDNAME"); value != "" {
		commandName = value
	} else {
		commandName = "kubectl"
	}

	log.Debugf("executing %s with args: %v", commandName, os.Args[1:])
	cmd := exec.Command(commandName, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func main() {
	configureLogging()

	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("ERROR: %s", err)
		}
	}()

	Kubecontext()
}
