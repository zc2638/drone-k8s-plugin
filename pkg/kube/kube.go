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

package kube

import (
	"encoding/base64"
	"errors"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
)

type Config struct {
	Server  string `json:"server"`
	SkipTLS bool   `json:"skip_tls"`
	CaCrt   string `json:"ca_crt"`
	Token   string `json:"token"`
}

func NewRestConfig(config *Config) (*rest.Config, error) {
	var restConfig *rest.Config
	if config.Token == "" {
		return nil, errors.New("kubernetes token must be defined")
	}

	token := strings.ReplaceAll(config.Token, " ", "")
	tokenBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	restConfig = &rest.Config{
		BearerToken: string(tokenBytes),
		Host:        config.Server,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	if !config.SkipTLS {
		restConfig.Insecure = false
		restConfig.CAData = []byte(config.CaCrt)
	}

	restConfig.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(1000, 1000)
	return restConfig, nil
}
