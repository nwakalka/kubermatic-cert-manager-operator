# Kubermatic TLS Certificate Manager Controller
### Introduction :

The TLS Certificate Management Controller provides an automated and self-service 
mechanism for requesting and managing TLS certificates.

Developers can define a custom `Certificate` resource in their application manifest to automatically generate TLS certificate, which are stored in a Kubernetes `Secret` and do not require interaction with an external system or Certified Authority (CA)
### Key Features :

- **Automated TLS Certificate Generation** : Automatically generate self-signed certificates to secure application endpoints.
- **Self-Service** : Developers can define a `Certificate` custom resource to request TLS certificates without external dependencies.
- **Secret Management** : Stores generated certificates in Kubernetes `Secret` objects for easy consumption by applications
- **Renewal Certificate** : Renew the certificate based on expiry

### Demo

#### Demo shows working of Kubermatic TLS Certificate Manager Controller

[![asciinema](https://asciinema.org/a/wxRKhzbqbnBNmcplRJJ4R7fnK.svg)](https://asciinema.org/a/wxRKhzbqbnBNmcplRJJ4R7fnK?speed=2)

### Getting started

#### 1. Prerequisites
* Go
* Kubernetes,Docker (i.e. Kind)
* Operator SDK 


#### 2. Installation (basic manifests)
* Docker image `nwakalka/k8c:latest`
* Deploy Operator including all manifests (CRDs, RBAC and Deployment) using Make command
```bash
make tls
```
##### Manifests location

* Install file which consists of all manifests [Install](./dist/install.yaml)
* Docker file for building docker images [Dockerfile](./Dockerfile)
* Deployment file for Operator [Deployment](./config/manager/manager.yaml)
* RBAC are cluster scoped [RBAC](./config/rbac)
* Custom Resource Definition [CRD](./config/crd/bases/certs.k8c.io_certificates.yaml)

#### 3. Testing the Controller
Once the installation is completed, Operator is running.

- Create Custom Resource `Certificate` as sample available [here](./config/samples/certs_v1_certificate.yaml)

```bash
apiVersion: certs.k8c.io/v1
kind: Certificate
metadata:
  labels:
    app.kubernetes.io/name: kubermatic-cert-manager-operator
    app.kubernetes.io/managed-by: kustomize
  name: my-certificate
  namespace: default
spec:
  dnsName: example.k8c.com
  validity: 365d
  secretRef:
    name: my-certificate-secret
```
- Operator will watch the Certificate and create a Self-signed TLS certificate for `example.k8c.com` for 365 days duration.
- Verify the Certificate details below

`kubectl get certificate -n default -o yaml`

- Secret named `my-certificate-secret` is created, verify the details.

`kubectl get secret <my-certificate-secret> -n <defaultNamespace>`

#### 4. Troubleshooting

Operator logs can be checked at 

`kubectl logs <podName> -n <operatorNamespace> -f`

#### 5. Uninstalling 

As we can install all manifests using make command, similarly we can uninstall using

`make remove`
