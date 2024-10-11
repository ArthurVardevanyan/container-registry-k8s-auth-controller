# container-registry-k8s-auth-controller (WIP)

This is an abstraction to convert the openshift monitoring, and openshift user workload monitoring configmaps into custom resources.

The Controller Inputs Two Custom Resources, and Converts them to ConfigMaps, for the cluster operator to pickup.

## Incepting Controller

How to Repo was setup

- <https://book.kubebuilder.io/cronjob-tutorial/cronjob-tutorial.html>
- <https://book.kubebuilder.io/cronjob-tutorial/new-api.html>

```bash
kubebuilder init --domain arthurvardevanyan.com --repo github.com/ArthurVardevanyan/container-registry-k8s-auth-controller
kubebuilder create api --group containerregistry --version v1beta1 --kind Auth --namespaced=true
```

## Getting Started

### Building Image

Build and push your image to the location specified by `IMG`:

```bash
go get -u
go mod tidy
```

```bash
make ko-build
```

### Running on the cluster

1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/container-registry-k8s-auth-controller:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/container-registry-k8s-auth-controller:tag
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller

UnDeploy the controller from the cluster:

```sh
make undeploy
```

## Contributing

// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out

1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
