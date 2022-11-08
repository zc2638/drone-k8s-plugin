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
	"os"

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
			if err := cfg.Parse(opt.ConfigPath, constants.ProjectName); err != nil {
				logrus.Fatal(err)
			}
			if err := cfg.Validate(); err != nil {
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
			if err := run(cfg, kubeClient, dynamicClient); err != nil {
				logrus.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVarP(&opt.ConfigPath, "config", "c", opt.ConfigPath,
		"config file (default is $HOME/.drone-plugin/config.yaml)")
	return cmd
}
