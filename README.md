# container-registry-k8s-auth-controller

This controller handles the creation and rotation of short lived pull secrets to

- Quay
- Artifact Registry

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
