# drone-k8s-plugin

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

| name                | required | type     | description                                                                                                                                                                                                                                                            |
|:--------------------|:--------:|:---------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| kubernetes_server   |    ✔️    | string   | The address and port of the Kubernetes API server.                                                                                                                                                                                                                     |
| k8s_server          |    ️     | string   | The same as `kubernetes_server`.                                                                                                                                                                                                                                       |
| kubernetes_token    |    ✔️    | string   | Token from ServiceAccount for authentication to the API server.                                                                                                                                                                                                        |
| k8s_token           |    ️     | string   | The same as `kubernetes_token`.                                                                                                                                                                                                                                        |
| kubernetes_ca_crt   |    ️     | string   | Certificate from ServiceAccount for authentication to the API server.                                                                                                                                                                                                  |
| k8s_ca_crt          |    ️     | string   | The same as `kubernetes_ca_crt`.                                                                                                                                                                                                                                       |
| kubernetes_skip_tls |    ️     | bool     | If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure.                                                                                                                                                    |
| k8s_skip_tls        |    ️     | bool     | The same as `kubernetes_skip_tls_verify`.                                                                                                                                                                                                                              |
| init_templates      |    ️     | []string | Path to Kubernetes Resource yaml based definition file (e.g. ConfigMap, Deployment or others), used to initialize some resources.                                                                                                                                      |
| templates           |    ️     | []string | Path to Kubernetes Resource yaml based definition file (e.g. ConfigMap, Deployment or others).                                                                                                                                                                         |
| config_files        |    ️     | []string | Config file paths for automatic creation/update of ConfigMap.The syntax is expressed as `namespace:name:file_path:file_name` or `namespace:name:file_path`, when file_name is not specified, it will default to the file name of file_path.                            |
| namespace           |    ️     | string   | Default namespace to use when namespace is not set.                                                                                                                                                                                                                    |
| debug               |    ️     | bool     | Used to enable debug level logging.                                                                                                                                                                                                                                    |

