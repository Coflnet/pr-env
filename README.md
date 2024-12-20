# pr-env
This kubernetes operator can detect pull requests in a github repository and create preview environments based on the open pull requests
Currently only Github is supported and only stateless applications

## Description
The project is still in early development
Plans for more documentation and examples are definitly planned (so far)

## Getting Started

The standard kubebuilder project template is being used
So documentation from [kubebuilder.io](https://kubebuilder.io) is mostly applicable

### Installation

To install the project, you can apply the manifests from the latest release
Or you can create a kustomize bundle to install the operator

```yaml
resources:
  - https://github.com/Coflnet/pr-env/releases/download/0.0.1-dev7/install.yaml
```

```sh
kubectl apply -k .
```

After installing the operator you can create a `PreviewEnvironment` resource
In that resource you can configure all the necessary settings
(The project is still in early development and the configuration will definitly change in the future)
```yaml
apiVersion: coflnet.coflnet.com/v1alpha1
kind: PreviewEnvironment
metadata:
  name: previewenvironment-sample
spec:

  # configure your github repository
  gitOrganization: Flou21
  gitRepository: test-page

  # configure the container registry the images should be pushed to
  containerRegistry:
    registry: index.docker.io
    repository: muehlhansfl

  # configure the preview environment application
  applicationSettings:
    ingressHostname: preview.flou.dev
    port: 80

```


## Development

### Prerequisites
- go version v1.23.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/pr-env:tag
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
make deploy IMG=<some-registry>/pr-env:tag
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

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/pr-env:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/pr-env/<tag or branch>/dist/install.yaml
```

## Contributing
The project is in a way too early stage for that

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

