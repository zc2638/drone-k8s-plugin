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

package tpl

import (
	"bytes"
	"text/template"

	"github.com/zc2638/drone-k8s-plugin/pkg/constants"
)

var tpl = template.New(constants.ProjectName)

func Render(in []byte, envMap map[string]string) ([]byte, error) {
	t, err := tpl.Parse(string(in))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]interface{}{"env": envMap}); err != nil {
		return nil, err
	}
	current := buf.Bytes()
	out := bytes.ReplaceAll(current, []byte("<no value>"), []byte(""))
	return out, nil
}
