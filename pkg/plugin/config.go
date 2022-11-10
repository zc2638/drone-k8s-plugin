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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zc2638/drone-k8s-plugin/pkg/constants"

	"github.com/zc2638/drone-k8s-plugin/pkg/kube"

	"github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type ConfigFile struct {
	Namespace string
	Name      string
	FilePath  string
	FileName  string
}

type Config struct {
	configFiles []ConfigFile

	Kubernetes kube.Config `json:"kubernetes"`

	InitTemplates []string `json:"init_templates"`
	ConfigFiles   []string `json:"config_files"` // namespace:name:file, namespace:name:file
	Templates     []string `json:"templates"`
	Namespace     string   `json:"namespace"`
	Debug         bool     `json:"debug"`
}

func (c *Config) BindEnvs() {
	c.bindEnv("debug")
	c.bindEnv("namespace")
	c.bindEnv("init_templates")
	c.bindEnv("templates")
	c.bindEnv("config_files")
	c.bindEnv("kubernetes.server", "k8s.server")
	c.bindEnv("kubernetes.token", "k8s.token")
	c.bindEnv("kubernetes.ca_crt", "k8s.ca_crt")
	c.bindEnv("kubernetes.skip_tls", "k8s.skip_tls")
}

func (c *Config) bindEnv(input ...string) {
	switch len(input) {
	case 0:
		return
	case 1:
		viper.MustBindEnv(input[0])
	default:
		values := make([]string, 0, len(input))
		values = append(values, input[0])
		for _, v := range input[1:] {
			val := fmt.Sprintf("%s_%s", constants.ProjectName, v)
			val = strings.ToUpper(val)
			values = append(values, val)
		}
		viper.MustBindEnv(values...)
	}
}

func (c *Config) GetConfigFiles() []ConfigFile {
	return c.configFiles[:]
}

func (c *Config) Validate() error {
	if len(c.InitTemplates) == 0 && len(c.ConfigFiles) == 0 && len(c.Templates) == 0 {
		return errors.New("at least one of init_templates, config_files and templates is defined")
	}

	cfs := make([]ConfigFile, 0, len(c.ConfigFiles))
	for _, v := range c.ConfigFiles {
		cf := ConfigFile{}
		parts := strings.Split(v, ":")
		switch len(parts) {
		case 3:
		case 4:
			cf.FileName = parts[3]
		default:
			return fmt.Errorf("config file (%s) format error, please use `namespace:name:file` or `namespace:name:filepath:filename` to define", v)
		}
		cf.Namespace = parts[0]
		cf.Name = parts[1]
		cf.FilePath = parts[2]

		_, err := os.Stat(cf.FilePath)
		if err != nil {
			return err
		}
		cfs = append(cfs, cf)
	}
	if len(cfs) > 0 {
		c.configFiles = cfs
	}

	return nil
}

func (c *Config) Parse(configPath string, envPrefix string) error {
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			return err
		}
		viper.AddConfigPath(filepath.Join(home, ".drone-plugin"))
		viper.SetConfigName("config.yaml")
	}
	if envPrefix != "" {
		viper.SetEnvPrefix(strings.ToUpper(envPrefix))
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv()
		viper.AllowEmptyEnv(true)
	}
	if err := viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
		case *os.PathError:
		default:
			return err
		}
	}
	return viper.Unmarshal(c, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	})
}
