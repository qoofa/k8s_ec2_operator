# ec2-operator

A Kubernetes operator that manages AWS EC2 instances as native resources.
Write an `Ec2instance` manifest, apply it, and the operator launches the instance, waits until it is running, and puts its details in the resource status. Delete the resource and the operator terminates the instance.

Built with [Kubebuilder](https://book.kubebuilder.io) and the [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2).

## How it works

```
kubectl apply ──▶ Ec2instance CR ──▶ Reconciler ──▶ EC2 RunInstances
                                         │              │
                                         │◀── wait until running
                                         ▼
                                 status: instanceID, state, IPs, DNS

kubectl delete ─▶ finalizer ─▶ EC2 TerminateInstances ─▶ wait ─▶ finalizer removed
```

1. **Create**: when a new `Ec2instance` appears, the controller adds the `ec2instance.compute.minhaj.blog` finalizer, calls `RunInstances`, and waits up to 3 minutes for the instance to reach `running`.
2. **Status**: it then writes the instance ID, state, public and private IPs, and DNS names to `.status`.
3. **Idempotency**: if `.status.instanceID` is already set, the controller skips creation.
4. **Delete**: when the resource is deleted, the controller terminates the instance, waits up to 5 minutes for termination, and removes the finalizer.

## API

`apiVersion: compute.minhaj.blog/v1`, `kind: Ec2instance`

### Spec

| Field               | Type              | Required | Description                         |
|---------------------|-------------------|----------|-------------------------------------|
| `instanceType`      | string            | yes      | EC2 instance type, e.g. `t2.micro`  |
| `amiId`             | string            | yes      | AMI to launch                       |
| `region`            | string            | yes      | AWS region, e.g. `us-east-1`        |
| `keyPair`           | string            | no       | EC2 key pair name for SSH           |
| `subnet`            | string            | no       | Subnet ID to launch into            |
| `availabilityZone`  | string            | no       | Availability zone                   |
| `securityGroups`    | []string          | no       | Security group IDs                  |
| `userData`          | string            | no       | Instance user data                  |
| `tags`              | map[string]string | no       | Instance tags                       |
| `storage`           | object            | no       | Root and additional volume config   |
| `associatePublicIp` | bool              | no       | Associate a public IP               |

> Only `instanceType`, `amiId`, `region`, `keyPair`, and `subnet` are wired to the AWS call at the moment.

### Status

| Field        | Description               |
|--------------|---------------------------|
| `instanceID` | EC2 instance ID           |
| `state`      | Instance state            |
| `publicIp`   | Public IPv4 address       |
| `privateIp`  | Private IPv4 address      |
| `publicDNS`  | Public DNS name           |
| `privateDNS` | Private DNS name          |

### Example

```yaml
apiVersion: compute.minhaj.blog/v1
kind: Ec2instance
metadata:
  name: ec2instance-sample
spec:
  instanceType: t2.micro
  amiId: ami-08d70e59c07c61a3a
  region: us-east-1
  keyPair: my-keypair
  subnet: subnet-0123456789abcdef0
  tags:
    environment: dev
```

## Getting started

### Prerequisites

- Go 1.24+
- Docker
- kubectl
- A Kubernetes cluster. Local configs are included for [kind](https://kind.sigs.k8s.io) (`.kind/`) and [k3d](https://k3d.io) (`.k3d/`)
- AWS credentials with EC2 permissions

### AWS credentials

The operator reads static credentials from the environment:

```sh
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
```

### Run locally

```sh
# start a local cluster (pick one)
kind create cluster --config .kind/kind-multi-cluster.yml
k3d cluster create --config .k3d/cluster.yml

# install the CRD and run the controller against your current kubeconfig
make install
make run

# in another terminal, edit the sample and apply it
kubectl apply -k config/samples/
kubectl get ec2instances -o yaml
```

### Deploy to a cluster

```sh
export IMG=<registry>/ec2-operator:tag
make docker-build docker-push IMG=$IMG
make deploy IMG=$IMG
```

When the operator runs in-cluster, the manager container needs `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` set (for example from a Secret).

### Uninstall

```sh
kubectl delete -k config/samples/   # terminates the EC2 instances
make undeploy
make uninstall
```

## Development

```sh
make manifests generate   # regenerate CRDs, RBAC, and deepcopy after API changes
make lint-fix             # lint and auto-fix
make test                 # unit tests (envtest)
make test-e2e             # e2e tests (needs a dedicated kind cluster)
make help                 # list all targets
```

### Project layout

```
api/v1/                      Ec2instance CRD types
cmd/main.go                  Manager entrypoint
internal/controller/         Reconciler and AWS EC2 helpers
config/                      Kustomize manifests (CRD, RBAC, manager, samples)
test/                        E2E tests
```

## License

Licensed under the Apache License, Version 2.0. See <http://www.apache.org/licenses/LICENSE-2.0>.
