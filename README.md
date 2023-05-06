# drone-k8s-plugin

![LICENSE](https://img.shields.io/github/license/zc2638/drone-k8s-plugin.svg?style=flat-square&color=blue)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/zc2638/drone-k8s-plugin/main.yml?branch=main&style=flat-square)](https://github.com/zc2638/drone-k8s-plugin/actions)


Drone CI plugin for creating & updating K8s Resources.  
This plugin supports all Kubernetes resources and also supports creating/updating Configmaps from config files.

## Usage

```shell
docker run --rm \
  -e PLUGIN_K8S_SERVER=https://localhost:6443 \
  -e PLUGIN_K8S_TOKEN=<your-token> \
  -e PLUGIN_K8S_SKIP_TLS_VERIFY=true \
  -e PLUGIN_TEMPLATES=testdata/deployment.yaml,testdata/service.yaml \
  -v <your-host-path>:/work/testdata
  zc2638/drone-k8s-plugin
```

### Environments

| name                | required | type     | description                                                                                                                                                                                                                                                                  |
|:--------------------|:--------:|:---------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| kubernetes_server   |    ✔️    | string   | The address and port of the Kubernetes API server.                                                                                                                                                                                                                           |
| k8s_server          |    ️     | string   | The same as `kubernetes_server`.                                                                                                                                                                                                                                             |
| kubernetes_token    |    ✔️    | string   | Token from ServiceAccount for authentication to the API server. The value must be base64 encoded.                                                                                                                                                                            |
| k8s_token           |    ️     | string   | The same as `kubernetes_token`.                                                                                                                                                                                                                                              |
| kubernetes_ca_crt   |    ️     | string   | Certificate from ServiceAccount for authentication to the API server. The value must be base64 encoded.                                                                                                                                                                      |
| k8s_ca_crt          |    ️     | string   | The same as `kubernetes_ca_crt`.                                                                                                                                                                                                                                             |
| kubernetes_skip_tls |    ️     | bool     | If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure.                                                                                                                                                          |
| k8s_skip_tls        |    ️     | bool     | The same as `kubernetes_skip_tls_verify`.                                                                                                                                                                                                                                    |
| init_templates      |    ️     | []string | Path to Kubernetes Resource yaml based definition file (e.g. ConfigMap, Deployment or others), used to initialize some resources.                                                                                                                                            |
| templates           |    ️     | []string | Path to Kubernetes Resource yaml based definition file (e.g. ConfigMap, Deployment or others).                                                                                                                                                                               |
| config_files        |    ️     | []string | Config file paths for automatic creation/update of ConfigMap.The syntax is expressed as `namespace:name:file_path:file_name` or `namespace:name:file_path`, when file_name is not specified, it will default to the file name of file_path.                                  |
| namespace           |    ️     | string   | Default namespace to use when namespace is not set.                                                                                                                                                                                                                          |
| debug               |    ️     | bool     | Used to enable debug level logging.                                                                                                                                                                                                                                          |

## Drone Example

```yaml
---
kind: pipeline
type: docker
name: drone-k8s-plugin-test

steps:
  - name: deploy
    image: zc2638/drone-k8s-plugin
    pull: if-not-exists
    settings:
      k8s_server: https://localhost:6443
      k8s_token:
        from_secret: k8s_token
      k8s_ca_crt:
        from_secret: k8s_ca_crt
      k8s_skip_tls: false
      namespace: default
      init_templates:
        - testdata/namespace.yaml
      config_files:
        - default:test-config:testdata/config.yaml
        - default:test-config:testdata/config.yaml:a.yaml
      templates:
        - testdata/deployment.yaml
        - testdata/service.yaml
        - testdata/*.yaml
      app_name: ${DRONE_REPO_NAME}
```

OR

```yaml
kind: pipeline
type: docker
name: drone-k8s-plugin-test

steps:
  - name: deploy
    image: zc2638/drone-k8s-plugin
    pull: if-not-exists
    environment:
      K8S_SERVER: https://localhost:6443
      K8S_TOKEN:
        from_secret: k8s_token
      K8S_SKIP_TLS: true
      NAMESPACE: default
      TEMPLATES: testdata/deployment.yaml,testdata/service.yaml
      APP_NAME: ${DRONE_REPO_NAME}
```