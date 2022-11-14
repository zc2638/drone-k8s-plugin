// Copyright © 2022 zc2638 <zc2638@qq.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/zc2638/drone-k8s-plugin/pkg/constants"
	"github.com/zc2638/drone-k8s-plugin/pkg/kube"
)

type Option struct {
	ConfigPath string
}

func (o *Option) Config() *Config {
	cfg := &Config{}
	return cfg
}

func NewCommand() *cobra.Command {
	opt := new(Option)
	opt.ConfigPath = os.Getenv(constants.ProjectName + "_CONFIG_PATH")

	cmd := &cobra.Command{
		Use:          constants.ProjectName,
		Short:        "Drone Kubernetes Plugin",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			cfg := opt.Config()
			cfg.BindEnvs()
			logrus.Infof("Config Path: %s", opt.ConfigPath)
			if err := cfg.Parse(opt.ConfigPath, constants.ProjectName); err != nil {
				logrus.Fatal(err)
			}

			envMap := getEnvMap()
			envs := envToSlice(envMap)
			if err := cfg.Validate(envs); err != nil {
				logrus.Fatal(err)
			}
			if cfg.Debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			logrus.Debugf("%#v\n", cfg)

			restConfig, err := kube.NewRestConfig(&cfg.Kubernetes)
			if err != nil {
				logrus.Fatal(err)
			}
			kubeClient, err := kubernetes.NewForConfig(restConfig)
			if err != nil {
				logrus.Fatal(err)
			}
			dynamicClient, err := dynamic.NewForConfig(restConfig)
			if err != nil {
				logrus.Fatal(err)
			}
			if err := run(cfg, kubeClient, dynamicClient, envMap); err != nil {
				logrus.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVarP(&opt.ConfigPath, "config", "c", opt.ConfigPath,
		"config file (default is $HOME/.drone-plugin/config.yaml)")
	return cmd
}

func getEnvMap() map[string]string {
	envMap := make(map[string]string)
	envs := os.Environ()
	for _, v := range envs {
		if pluginExp.MatchString(v) {
			matches := pluginExp.FindStringSubmatch(v)
			key := strings.ToLower(matches[1])
			envMap[key] = matches[2]
			logrus.Debugf("env: %s=%s", key, matches[2])
		}

		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			continue
		}
		envMap[parts[0]] = parts[1]
		logrus.Debugf("env: %s=%s", parts[0], parts[1])
	}
	return envMap
}

func envToSlice(set map[string]string) []string {
	env := make([]string, 0, len(set))
	for k, v := range set {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}
