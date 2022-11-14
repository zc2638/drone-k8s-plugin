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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	pkgruntime "k8s.io/apimachinery/pkg/runtime"

	v1 "k8s.io/api/core/v1"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"

	"github.com/zc2638/drone-k8s-plugin/pkg/kube"
	"github.com/zc2638/drone-k8s-plugin/pkg/tpl"
)

var pluginExp = regexp.MustCompile(`^PLUGIN_(.*)=(.*)`)

func run(cfg *Config, kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) error {
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

	initObjSet, err := parseObjectSet(cfg.InitTemplates, envMap)
	if err != nil {
		return fmt.Errorf("parse init_templates failed: %v", err)
	}
	objSet, err := parseObjectSet(cfg.Templates, envMap)
	if err != nil {
		return fmt.Errorf("parse templates failed: %v", err)
	}

	gr, err := restmapper.GetAPIGroupResources(kubeClient.Discovery())
	if err != nil {
		return fmt.Errorf("get Kubernetes API group resources failed: %v", err)
	}
	mapping := restmapper.NewDiscoveryRESTMapper(gr)

	logrus.Debug("Start to apply resources from init templates")
	if err := applyResources(dynamicClient, mapping, initObjSet, cfg.Namespace); err != nil {
		return err
	}
	logrus.Debug("Start to apply configmaps from config files")
	if err := applyForConfig(kubeClient, cfg.GetConfigFiles()); err != nil {
		return err
	}
	logrus.Debug("Start to apply resources from templates")
	if err := applyResources(dynamicClient, mapping, objSet, cfg.Namespace); err != nil {
		return err
	}
	return nil
}

func parseObjectSet(templates []string, envMap map[string]string) ([][]unstructured.Unstructured, error) {
	if len(templates) == 0 {
		return nil, nil
	}

	objSet := make([][]unstructured.Unstructured, 0, len(templates))
	for _, v := range templates {
		ext := filepath.Ext(v)
		isYamlFile := ext == ".yaml" || ext == ".yml"
		if !isYamlFile {
			logrus.Warnf("Ignore file (%s), not a yaml or yml file", v)
			continue
		}

		fileBytes, err := os.ReadFile(v)
		if err != nil {
			return nil, fmt.Errorf("read template file(%s) failed: %v", v, err)
		}
		current, err := tpl.Render(fileBytes, envMap)
		if err != nil {
			return nil, fmt.Errorf("render template file(%s) failed: %v", v, err)
		}

		result, err := kube.ParseObject(current)
		if err != nil {
			return nil, fmt.Errorf("parse template file(%s) failed: %v", v, err)
		}
		objSet = append(objSet, result)
	}
	return objSet, nil
}

func applyResources(
	dynamicClient dynamic.Interface,
	mapping meta.RESTMapper,
	objSet [][]unstructured.Unstructured,
	defNamespace string,
) error {
	for _, objs := range objSet {
		eg, ctx := errgroup.WithContext(context.Background())

		for _, obj := range objs {
			objCopy := obj.DeepCopy()

			eg.Go(func() error {
				gvk := objCopy.GroupVersionKind()
				logrus.WithField("apiVersion", gvk.GroupVersion().String()).
					WithField("kind", gvk.Kind).
					WithField("namespace", objCopy.GetNamespace()).
					WithField("name", objCopy.GetName()).
					Info("Apply Resource")

				restMapping, err := mapping.RESTMapping(gvk.GroupKind(), gvk.Version)
				if err != nil {
					return err
				}

				var resourceInter dynamic.ResourceInterface
				if restMapping.Scope.Name() == meta.RESTScopeNameNamespace {
					if objCopy.GetNamespace() == "" {
						if defNamespace == "" {
							return fmt.Errorf(
								"apply resource failed: namespace must be defined, apiVersion=%s, kind=%s, name=%s",
								gvk.GroupVersion().String(), gvk.Kind, objCopy.GetName(),
							)
						}
						// set default namespace
						objCopy.SetNamespace(defNamespace)
					}
					resourceInter = dynamicClient.Resource(restMapping.Resource).Namespace(objCopy.GetNamespace())
				} else {
					resourceInter = dynamicClient.Resource(restMapping.Resource)
				}

				// apply
				origin, err := resourceInter.Get(ctx, objCopy.GetName(), metav1.GetOptions{
					TypeMeta: metav1.TypeMeta{
						Kind:       objCopy.GetKind(),
						APIVersion: objCopy.GetAPIVersion(),
					},
				})
				if err == nil {
					switch objCopy.GetKind() {
					case "Service":
						objCopy, err = completeService(origin, objCopy)
						if err != nil {
							return err
						}
					default:
					}

					rv, _ := strconv.ParseInt(origin.GetResourceVersion(), 10, 64)
					objCopy.SetResourceVersion(strconv.FormatInt(rv, 10))
					if _, err = resourceInter.Update(ctx, objCopy, metav1.UpdateOptions{}); err != nil {
						err = fmt.Errorf("update %s %s failed: %v", objCopy.GetKind(), objCopy.GetName(), err)
					}
					return err
				}
				if !apierrors.IsNotFound(err) {
					return err
				}
				if _, err = resourceInter.Create(ctx, objCopy, metav1.CreateOptions{}); err != nil {
					err = fmt.Errorf("create %s %s failed: %v", objCopy.GetKind(), objCopy.GetName(), err)
				}
				return err
			})
		}
		if err := eg.Wait(); err != nil {
			return err
		}
	}
	return nil
}

func completeService(origin, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	var (
		originSvc v1.Service
		objSvc    v1.Service
	)
	if err := pkgruntime.DefaultUnstructuredConverter.FromUnstructured(origin.UnstructuredContent(), &originSvc); err != nil {
		return nil, fmt.Errorf("convert origin unstructured object %s to Service failed: %v", obj.GetName(), err)
	}
	if err := pkgruntime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &objSvc); err != nil {
		return nil, fmt.Errorf("convert unstructured object %s to Service failed: %v", obj.GetName(), err)
	}

	objSvc.Spec.ClusterIP = originSvc.Spec.ClusterIP
	objSvc.Spec.ClusterIPs = originSvc.Spec.ClusterIPs

	unstructuredContent, err := pkgruntime.DefaultUnstructuredConverter.ToUnstructured(&objSvc)
	if err != nil {
		return nil, fmt.Errorf("convert Service %s to unstructured object failed: %v", objSvc.GetName(), err)
	}

	current := &unstructured.Unstructured{}
	current.SetUnstructuredContent(unstructuredContent)
	return current, nil
}

func applyForConfig(kubeClient kubernetes.Interface, cfs []ConfigFile) error {
	if len(cfs) == 0 {
		return nil
	}

	cmSet := make(map[string]*v1.ConfigMap)
	for _, v := range cfs {
		key := fmt.Sprintf("%s/%s", v.Namespace, v.Name)
		cm, ok := cmSet[key]
		if !ok {
			cm = &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      v.Name,
					Namespace: v.Namespace,
				},
				Data: make(map[string]string),
			}
			cmSet[key] = cm
		}

		fileBytes, err := os.ReadFile(v.FilePath)
		if err != nil {
			return err
		}
		filename := v.FileName
		if len(filename) == 0 {
			_, filename = filepath.Split(v.FilePath)
		}
		cm.Data[filename] = string(fileBytes)
	}

	for _, cm := range cmSet {
		cmInter := kubeClient.CoreV1().ConfigMaps(cm.Namespace)
		origin, err := cmInter.Get(context.Background(), cm.Name, metav1.GetOptions{})
		if err == nil {
			rv, _ := strconv.ParseInt(origin.GetResourceVersion(), 10, 64)
			cm.SetResourceVersion(strconv.FormatInt(rv, 10))
			if _, err := cmInter.Update(context.Background(), cm, metav1.UpdateOptions{}); err != nil {
				return fmt.Errorf("update ConfigMap %s failed: %v", cm.Name, err)
			}
			logrus.WithField("namespace", cm.Namespace).
				WithField("name", cm.Name).
				Infof("Update ConfigMap")
			continue
		}
		if !apierrors.IsNotFound(err) {
			return err
		}
		if _, err := cmInter.Create(context.Background(), cm, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("create ConfigMap %s failed: %v", cm.Name, err)
		}
		logrus.WithField("namespace", cm.Namespace).
			WithField("name", cm.Name).
			Infof("Create ConfigMap")
	}
	return nil
}
