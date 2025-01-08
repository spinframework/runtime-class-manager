# runtime-class-manager

runtime-class-manager is a Kubernetes operator that manages installation of Wasm shims onto nodes and related Runtimeclasses via [Shim custom resources](../../config/crd/bases/runtime.spinkube.dev_shims.yaml).

## Prerequisites

- [Kubernetes v1.20+](https://kubernetes.io/docs/setup/)
- [Helm v3](https://helm.sh/docs/intro/install/)

## Installing the chart

The following installs the runtime-class-manager chart with the release name `rcm`:

```shell
helm upgrade --install \
  --namespace rcm \
  --create-namespace \
  --wait \
  rcm .
```

## Post-installation

With runtime-class-manager running, you're ready to create one or more Wasm Shims. See the samples in the [config/samples directory](../../config/samples/).

> Note: Ensure that the `location` for the specified shim binary points to the correct architecture for your Node(s)

For example, here we install the Spin shim:

```shell
kubectl apply -f https://raw.githubusercontent.com/spinkube/runtime-class-manager/refs/heads/main/config/samples/test_shim_spin.yaml
```

Now when you annotate one or more nodes with a label corresponding to the `nodeSelector` declared in the Shim, runtime-class-manager will install the shim as well as create the corresponding Runtimeclass:

```shell
kubectl label node --all spin=true
```

You are now ready to deploy your Wasm workloads.
