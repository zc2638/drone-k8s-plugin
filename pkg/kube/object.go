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
	"bytes"
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
	serializeryaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

var serializer = serializeryaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

func ParseObject(data []byte) ([]unstructured.Unstructured, error) {
	var result []unstructured.Unstructured
	decoder := utilyaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 256)
	for {
		var rawObj pkgruntime.RawExtension
		if err := decoder.Decode(&rawObj); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decode data to raw object failed: %v", err)
		}

		var obj unstructured.Unstructured
		if err := pkgruntime.DecodeInto(serializer, rawObj.Raw, &obj); err != nil {
			return nil, fmt.Errorf("decode raw object failed: %v", err)
		}
		result = append(result, obj)
	}
	return result, nil
}
