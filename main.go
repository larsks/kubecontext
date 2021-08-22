/*

kubecontext (c) 2021 Lars Kellogg-Stedman <lars@oddbit.com>

Licensed under the Apache License, Version 2.0 (the "License"); you may
not use this file except in compliance with the License. You may obtain
a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
License for the specific language governing permissions and limitations
under the License.

*/

package main

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

func findConfigRecursive(config *Config) {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if cwd != "/" {
		os.Chdir("..")
		findConfigRecursive(config)
		os.Chdir(cwd)
	}

	log.Debugf("looking for .kubecontext in %s", cwd)

	if _, err := os.Stat("./.kubecontext"); err == nil {
		kubecontext := filepath.Join(cwd, ".kubecontext")
		log.Debugf("found %s", kubecontext)
		config.FromFile(kubecontext)
	}
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
	var config Config

	tmpfile, err := ioutil.TempFile("", "kubeconfig")
	if err != nil {
		panic(err)
	}
	defer func() {
		log.Debugf("removing %s", tmpfile.Name())
		os.Remove(tmpfile.Name())
	}()

	findConfigRecursive(&config)
	if config.Kubeconfig != "" {
		config.SetKubeconfig()
	}

	if err := generateKubeconfig(tmpfile); err != nil {
		panic(err)
	}

	config.Apply()

	// If you have a project that requires `oc` instead of `kubectl`,
	// you can set `K_COMMANDNAME` in your environment (or in the
	// `environment` section of your `.kubecontext` file.
	if config.Command != "" {
		commandName = config.Command
	} else {
		commandName = "kubectl"
	}

	log.Debugf("executing %s with args: %v", commandName, os.Args[1:])
	cmd := exec.Command(commandName, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
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
