# Runtime Class Manager

[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/spinframework/runtime-class-manager/badge)](https://securityscorecards.dev/viewer/?uri=github.com/spinframework/runtime-class-manager)

## Overview

Runtime Class Manager is Kubernetes Operator that assists with [Wasm](https://webassembly.org/) runtime installation and configuration on a Kubernetes cluster. It does so by handling creation and installation of [RuntimeClasses](./docs/runtimeclass.md) and [containerd](https://containerd.io/) shim binaries for each [Shim](./docs/shim.md) custom resource created on a cluster.

## Background

The Runtime Class Manager is the spiritual successor to the kwasm-operator. kwasm has been developed as an experimental, simple way to install Wasm runtimes. This experiment has been relatively successful, as more and more users utilized it to fiddle around with Wasm on Kubernetes. However, the kwasm-operator has some limitations that make it difficult to use in production. The Runtime Class Manager is an attempt to address these limitations to make it a reliable and secure way to deploy arbitrary containerd shims.

The implementation of Runtime Class Manager follows [this](https://hackmd.io/TwC8Fc8wTCKdoWlgNOqTgA) community proposal.

## Roadmap

For the 1.0 release of Runtime Class Manager, we consider three milestones:

- **M1: [RCM MVP for Spinkube](https://github.com/spinframework/runtime-class-manager/milestone/1)**  
This milestone is about getting RCM to a state where Spinkube can rely on RCM and use it as a dependency instead of Kwasm. This means, that the focus is on managing lifecycle of [containerd-shim-spin](https://github.com/spinframework/containerd-shim-spin) on nodes. _This is now complete._
- **M2: [Kwasm Feature Parity](https://github.com/spinframework/runtime-class-manager/milestone/2)**  
All shims that kwasm can install, should be installable via rcm. Automated tests are in place to ensure installation of RCM and shims that are supported by Kwasm.
- **M3: [Full implementation of the initial spec](https://github.com/spinframework/runtime-class-manager/milestone/3)**  
Stable spec of the Shim CRD based on the [initial proposal](https://hackmd.io/TwC8Fc8wTCKdoWlgNOqTgA). After 1.0 we assume no breaking changes of the Shim CRD. Arbitrary shims can be installed via RCM and prominent shims are tested automatically, on various Kubernetes distributions.
- Future (ideas):
  - support for additional container runtimes, like CRI-O to enable RCM on OpenShift
  - alternative shim installation via Daemonset instead of Jobs
  - treating node-installer as a daemon process, to enable better conflict resolution

## Development

### To Deploy on the cluster

**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/runtime-class-manager:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/runtime-class-manager:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall

**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## License

See [LICENSE](./LICENSE).
