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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

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

// Set context specified in the Config object
func (config *Config) SetContext() {
	if config.Context != "" {
		log.Infof("setting context to %s", config.Context)
		cmd := exec.Command("kubectl", "config", "use-context", config.Context)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			panic(fmt.Errorf("failed to set context"))
		}
	} else {
		log.Debugf("config has no context")
	}
}

// Set current namespace specified in the Config object
func (config *Config) SetNamespace() {
	if config.Namespace != "" {
		log.Infof("setting namespace to %s", config.Namespace)
		cmd := exec.Command(
			"kubectl", "config", "set-context", "--current",
			"--namespace", config.Namespace,
		)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			panic(fmt.Errorf("failed to set namespace"))
		}
	} else {
		log.Debugf("config has no namespace")
	}
}

// Set environment variables specified in the Config object
func (config *Config) SetEnv() {
	if config.Environment != nil {
		for name, value := range config.Environment {
			log.Infof("setting environment variable %s to %s", name, value)
			if err := os.Setenv(name, value); err != nil {
				panic(fmt.Errorf("failed to set environment variable"))
			}
		}
	} else {
		log.Debugf("config has no environment variables")
	}
}

// Read configuration from the specified YAML file
func (config *Config) FromFile(configfile string) {
	data, err := ioutil.ReadFile(configfile)
	if err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		panic(err)
	}

	log.Debugf("%s has config %+v", configfile, config)
}

// Apply the settings described by the Config object
func (config *Config) Apply() {
	config.SetContext()
	config.SetNamespace()
	config.SetEnv()
}
