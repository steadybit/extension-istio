<img src="./logo.png" height="130" align="right" alt="Istio logo depicting sails with the text 'Datadog'">

# Steadybit extension-istio

This extension enables the injection of HTTP and gRPC faults into [Istio's virtual services](https://istio.io/latest/docs/reference/config/networking/virtual-service). Currently supported capabilities:
 - Discover virtual services and
   - Inject HTTP delay faults
   - Inject HTTP abort faults
   - Inject gRPC abort faults

## Configuration

| Environment Variable                  | Meaning                                                                                                                                                                | Default |
|---------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|
| `STEADYBIT_EXTENSION_CLUSTER_NAME`    | Kubernetes cluster name.                                                                                                                                               |         |
| `STEADYBIT_EXTENSION_PORT`            | Port number that the HTTP server should bind to.                                                                                                                       | 8080    |
| `STEADYBIT_EXTENSION_TLS_SERVER_CERT` | Optional absolute path to a TLS certificate that will be used to open an **HTTPS** server.                                                                             |         |
| `STEADYBIT_EXTENSION_TLS_SERVER_KEY`  | Optional absolute path to a file containing the key to the server certificate.                                                                                         |         |
| `STEADYBIT_EXTENSION_TLS_CLIENT_CAS`  | Optional comma-separated list of absolute paths to files containing TLS certificates. When specified, the server will expect clients to authenticate using mutual TLS. |         |
| `STEADYBIT_LOG_FORMAT`                | Defines the log format that the extension will use. Possible values are `text` and `json`.                                                                             | text    |
| `STEADYBIT_LOG_LEVEL`                 | Defines the active log level. Possible values are `debug`, `info`, `warn` and `error`.                                                                                 | info    |

## Running the Extension

### Using Helm in Kubernetes

```sh
$ helm repo add steadybit-extension-istio https://steadybit.github.io/extension-istio
$ helm repo update
$ helm upgrade steadybit-extension-istio \
    --install \
    --wait \
    --timeout 5m0s \
    --create-namespace \
    --namespace steadybit-extension \
    --set kubernetes.clusterName="my-cluster" \
    steadybit-extension-istio/steadybit-extension-istio
```
