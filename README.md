# Kubermatic TLS Certificate Manager Controller

### Introduction :

The TLS Certificate Management Controller provides an automated and self-service 
mechanism for requesting and managing TLS certificates.

Developers can define a custom `Certificate` resource in their application manifest to automatically generate TLS certificate, which are stored in a Kubernetes `Secret` and do not require interaction with an external system or Certified Authority (CA)
### Key Features :

- **Automated TLS Certificate Generation** : Automatically generate self-signed certificates to secure application endpoints.
- **Self-Service** : Developers can define a `Certificate` custom resource to request TLS certificates without external dependencies.
- **Secret Management** : Stores generated certificates in Kubernetes `Secret` objects for easy consumption by applications

 ## Demo
[![asciinema](https://asciinema.org/a/wxRKhzbqbnBNmcplRJJ4R7fnK.svg)(https://asciinema.org/a/wxRKhzbqbnBNmcplRJJ4R7fnK?speed=2)