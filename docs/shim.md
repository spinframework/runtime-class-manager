## Shim

Runtime Class Manager operates on `Shim` custom resources based on the [Shim CRD](../config/crd/bases/runtime.spinkube.dev_shims.yaml).

Whenever a Shim is created, updated or deleted, Runtime-Class-Manager will perform the necessary actions, e.g. creating, updating or removing the associated [RuntimeClass](./runtimeclass.md), installing or removing shim binaries on any [Nodes](https://kubernetes.io/docs/concepts/architecture/nodes/) labeled with a corresponding to its `nodeSelector`, etc.

### Configuration

For full, detailed configuration options, see the [Shim CRD](../config/crd/bases/runtime.spinkube.dev_shims.yaml). Here we point out a few pertinent items.

* `spec.nodeSelector`: The label key and value applied to Nodes where this particular shim should be installed
* `spec.fetchStrategy`: The strategy for fetching the shim binary
  * `spec.fetchStrategy.type`: `anonymousHttp` is the only option currently supported.
  * `spec.fetchStrategy.anonHttp.location`: The URL where the shim binary can be downloaded
* `spec.containerdRuntimeOptions`: Options specific to the shim that should be added to the containerd configuration

### Operation

You may observe the "install" and "uninstall" [Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/) responsible for downloading and installing (or uninstalling) the shim binary. These will run on every Node that matches the Shim's `nodeSelector`.
