# Flipper Operator
Flipper Operator - A Kubernetes operator for performing rolling restarts of deployments based on specific labels at configured intervals.

# Flipper Operator

The Flipper Operator is a Kubernetes operator designed to facilitate rolling restarts of deployments based on specific labels at configured intervals.

## Getting Started

### Prerequisites

Before getting started, ensure you have the following installed:

- Kubernetes cluster (local or remote)
- kubectl
- Kubebuilder


### Using Kubebuilder
The Flipper Operator is implemented using Kubebuilder, a framework for building Kubernetes APIs using the concepts of Kubernetes Custom Resource Definitions (CRDs). Kubebuilder is a comprehensive development tool that enables users to build and maintain Kubernetes APIs and controllers with ease. More information about Kubebuilder can be found [here](https://book.kubebuilder.io/).

#### Generate Flipper Operator Initial Code Using Kubebuilder

**Initialize the Project::**
Initializes a new Kubebuilder project with the specified domain, owner, repository, and project name.

```sh
kubebuilder init --domain "example.com" --owner "Abhijeet Rokade" --repo "github.com/sigsegv1989/flipper-operator" --project-name "flipper-operator"
```
**Create API and Controller:**
Creates the API definition and controller scaffolding for the RollingUpdate custom resource.
```sh
kubebuilder create api --group flipper --version v1alpha1 --kind RollingUpdate
```
#### Extensibility
To add a new Group and Kind to the existing Flipper Operator, use the following command:
```sh
kubebuilder create api --group <new-group> --version <new-version> --kind <new-kind>
```
This command scaffolds the necessary files for the new custom resource and controller, allowing you to extend the operator's functionality.

## Installtion
### Clone the repository:
Clone the Flipper Operator repository from GitHub to your local machine

```sh
git clone git@github.com:sigsegv1989/flipper-operator.git
cd flipper-operator
```
### Build the Operator binary:
To build the manager binary for the Flipper Operator:

 ```sh
make build
```
### Build the Operator Docker image:
Build the Docker image for the Flipper Operator with the specified image tag (yourimage/flipper-operator:v0.1.0):
```sh
make docker-build IMG=yourimage/flipper-operator:v0.1.0
```
### Push the Operator image to Registry
Push the Docker image of the Flipper Operator to your container registry
```sh
make docker-push IMG=yourimage/flipper-operator:v0.1.0
```
### Operator Docker Hub Image
Alternatively, you can pull the pre-built Docker image from Docker Hub:
```sh
docker pull pai314/flipper-operator:latest
```
Replace pai314/flipper-operator:latest with the appropriate image tag based on your requirements.

### Deploy the Operator:
Deploy the Flipper Operator to your Kubernetes cluster using the following command:

```sh
make deploy IMG=yourimage/flipper-operator:v0.1.0
```
This command installs and runs the Flipper Operator on your Kubernetes cluster, using the Docker image specified by IMG.

## RollingUpdate Custom Resource Definition (CRD) Documentation

For detailed information about the RollingUpdate custom resource, including its structure, fields, and usage examples, refer to the [RollingUpdate CRD README](./config/crd/README.md).


## Testing
### Unit Tests
Run unit tests to verify the correctness of the code:

```sh
make test
```
### End-to-End (E2E) Tests
Run end-to-end tests to validate the operator's functionality in a Kubernetes environment:
```sh
make test-e2e
```
### Linting
Run linting to ensure code consistency and adherence to best practices:
```sh
make lint
```
### Fix Linting Issues
Run linting with fixes to automatically correct linting issues where possible:
```sh
make lint-fix
```
### Run all Tests (Unit + E2E)
For comprehensive testing, run all tests including unit tests and end-to-end tests
```sh
make test-all
```

## Manual Testing

### Deploying Test Deployments and RollingUpdate CRs
To manually test the Flipper Operator, follow these steps to deploy test deployments and RollingUpdate custom resources (CRs) to your Kubernetes cluster:

#### Step 1: Deploy the Flipper Operator
Ensure the Flipper Operator is deployed on your Kubernetes cluster. If not already deployed, follow the [Installation](#installation) instructions in the README.

#### Step 2: Deploy Test Deployments
Deploy the test deployments using the provided YAML files in the `config/test` folder. These deployments are located in namespaces `test-1`, `test-2`, and `test-3`.

```sh
kubectl apply -f config/test/deployment-1.yaml
kubectl apply -f config/test/deployment-2.yaml
kubectl apply -f config/test/deployment-3.yaml
```
#### Step 3: Create RollingUpdate CRs
Create RollingUpdate CRs to trigger rolling updates for the deployments. Use the provided YAML files rollingupdate-1.yaml, rollingupdate-2.yaml, and rollingupdate-3.yaml.
```sh
kubectl apply -f config/test/rollingupdate-1.yaml
kubectl apply -f config/test/rollingupdate-2.yaml
kubectl apply -f config/test/rollingupdate-3.yaml
```
#### Step 4: Verify the Rolling Updates
**1. Verify Flipper Operator Reconciliation Logs**

Check the logs of the Flipper Operator to confirm that the reconciliation process is running smoothly and to monitor any related events
```sh
```

**2. Verify RollingUpdate CR Status**
Retrieve the status of the RollingUpdate custom resources to ensure they reflect the ongoing updates and their completion
```sh
```

**3. Deployment and Pod Annotations**
Verify the annotations on the deployments affected by the rolling updates to ensure they reflect the latest changes and updates
```sh
```

**4. Check ReplicaSet Revision and Pod Restart Times**
Review the ReplicaSet revision history to verify that new revisions are created during the rolling update process, and monitor the pod restart times to ensure that pods are being restarted as expected
```sh
```

## License

This project is licensed under the [MIT License](./LICENSE).
