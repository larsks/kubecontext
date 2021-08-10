package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

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
func findKubecontext() []string {
	var contexts []string

	for {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("failed to determine current directory: %v", err)
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

	return contexts
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

func main() {
	var commandName string

	configureLogging()

	kubecontexts := findKubecontext()

	// If we discovered one or more .kubecontext files, iterate over them
	// in reverse order, applying the configuration from each one.
	if kubecontexts != nil {
		for i := range kubecontexts {
			current := kubecontexts[len(kubecontexts)-i-1]
			if err := processKubecontext(current); err != nil {
				log.Fatalf("failed to process %s: %v", current, err)
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

	if path, err := exec.LookPath(commandName); err == nil {
		os.Args[0] = path
		log.Debugf("executing %s with args: %v", path, os.Args)
		if err = syscall.Exec(path, os.Args, os.Environ()); err != nil {
			log.Fatalf("failed to execute kubectl: %v", err)
		}
	} else {
		log.Errorf("did not find %s in PATH", commandName)
	}
}
