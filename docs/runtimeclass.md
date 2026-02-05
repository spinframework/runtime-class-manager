## RuntimeClass

Runtime Class Manager is in charge of creating a [RuntimeClass](https://kubernetes.io/docs/concepts/containers/runtime-class) for each [Shim](./shim.md) resource created on the cluster.

The `spec.runtimeClass` section of the Shim resource configures the RuntimeClass that will be created.

* `spec.runtimeClass.name`: Name of the Kubernetes RuntimeClass
  - This name should match what is expected by shim-specific operator(s) on the cluster
  - For example, the [Spin Operator](https://github.com/spinframework/spin-operator) utilizes a [SpinAppExecutor](https://www.spinkube.dev/docs/reference/spin-app-executor/) resource
    to run Spin Apps; the default RuntimeClass name it expects can be seen [here](https://github.com/spinframework/spin-operator/blob/main/config/samples/spin-shim-executor.yaml)
* `spec.runtimeClass.handler`: Name of the shim as it is referenced in the containerd config
