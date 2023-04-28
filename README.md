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

The extension supports all environment variables provided by [steadybit/extension-kit](https://github.com/steadybit/extension-kit#environment-variables).

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
