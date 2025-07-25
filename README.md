# Digicloud Issuer for cert-manager

A cert-manager external issuer for Digicloud DNS that supports DNS01 challenges for automatic certificate provisioning.

## Overview

This project implements a cert-manager external issuer that integrates with the Digicloud Edge DNS API to automatically handle DNS01 ACME challenges. It allows you to obtain SSL/TLS certificates for domains managed through Digicloud's DNS service.

## Features

- **DNS01 Challenge Support**: Automatically creates and removes TXT records for ACME validation
- **Wildcard Certificates**: Supports wildcard domain certificates (*.example.com)
- **Namespace Support**: Works with Digicloud's namespace-based multi-tenancy
- **High Availability**: Supports both Issuer (namespace-scoped) and ClusterIssuer (cluster-scoped) resources
- **Kubernetes Native**: Fully integrated with cert-manager and Kubernetes

## Prerequisites

- Kubernetes cluster with cert-manager installed
- Digicloud account with Edge DNS service
- API credentials for Digicloud
- Domain managed through Digicloud DNS

## Installation

### 1. Install cert-manager

If you haven't already installed cert-manager:

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.2/cert-manager.yaml
```

### 2. Install Digicloud Issuer

```bash
kubectl apply -f https://github.com/digicloud/digicloud-issuer/releases/latest/download/install.yaml
```

### 3. Create API Credentials Secret

Create a Kubernetes secret with your Digicloud API credentials:

```bash
kubectl create secret generic digicloud-credentials \
  --from-literal=token=your-api-token \
  --from-literal=namespace=your-digicloud-namespace
```

## Usage

### Creating an Issuer

Create a namespace-scoped issuer:

```yaml
apiVersion: digicloud.io/v1alpha1
kind: DigicloudIssuer
metadata:
  name: digicloud-issuer
  namespace: default
spec:
  # Digicloud API configuration
  apiUrl: "https://api.digicloud.ir"  # Optional, defaults to this
  namespace: "your-digicloud-namespace"
  
  # Reference to the secret containing API credentials
  authSecretName: digicloud-credentials
```

### Creating a ClusterIssuer

Create a cluster-scoped issuer:

```yaml
apiVersion: digicloud.io/v1alpha1
kind: DigicloudClusterIssuer
metadata:
  name: digicloud-cluster-issuer
spec:
  apiUrl: "https://api.digicloud.ir"
  namespace: "your-digicloud-namespace"
  authSecretName: digicloud-credentials
```

### Issuing Certificates

Request a certificate using the issuer:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-tls
  namespace: default
spec:
  secretName: example-tls-secret
  issuerRef:
    name: digicloud-issuer
    kind: DigicloudIssuer
    group: digicloud.io
  dnsNames:
  - example.com
  - www.example.com
  - "*.example.com"  # Wildcard domain
```

### Automatic Certificate with Ingress

Use annotations to automatically provision certificates:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: example-ingress
  annotations:
    cert-manager.io/issuer: "digicloud-issuer"
    cert-manager.io/issuer-kind: "DigicloudIssuer"
    cert-manager.io/issuer-group: "digicloud.io"
spec:
  tls:
  - hosts:
    - example.com
    - www.example.com
    secretName: example-tls-secret
  rules:
  - host: example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: example-service
            port:
              number: 80
```

## Configuration

### IssuerSpec Fields

| Field | Description | Required | Default |
|-------|-------------|----------|---------|
| `apiUrl` | Digicloud API base URL | No | `https://api.digicloud.ir` |
| `namespace` | Digicloud namespace | Yes | - |
| `authSecretName` | Name of secret containing credentials | Yes | - |

### Secret Format

The authentication secret must contain:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: digicloud-credentials
type: Opaque
data:
  token: <base64-encoded-api-token>
  namespace: <base64-encoded-digicloud-namespace>
```

## Development

### Prerequisites

- Go 1.21+
- Docker
- Kind (for local testing)
- kubectl
- Kubebuilder

### Building

```bash
# Build the manager binary
make build

# Build Docker image
make docker-build

# Run tests
make test

# Generate manifests
make manifests
```

### Local Development

1. Start a Kind cluster:
```bash
make kind-cluster
```

2. Install cert-manager:
```bash
make deploy-cert-manager
```

3. Build and load the image:
```bash
make docker-build kind-load
```

4. Deploy the issuer:
```bash
make deploy
```

5. Run end-to-end tests:
```bash
make e2e
```

## API Reference

### DigicloudIssuer

DigicloudIssuer is a namespace-scoped resource for issuing certificates.

```yaml
apiVersion: digicloud.io/v1alpha1
kind: DigicloudIssuer
metadata:
  name: example-issuer
  namespace: default
spec:
  apiUrl: "https://api.digicloud.ir"
  namespace: "digicloud-namespace"
  authSecretName: "digicloud-credentials"
status:
  conditions:
  - type: Ready
    status: "True"
    reason: "Verified"
    message: "DigicloudIssuer verified and ready to issue certificates"
```

### DigicloudClusterIssuer

DigicloudClusterIssuer is a cluster-scoped resource for issuing certificates.

```yaml
apiVersion: digicloud.io/v1alpha1
kind: DigicloudClusterIssuer
metadata:
  name: example-cluster-issuer
spec:
  apiUrl: "https://api.digicloud.ir"
  namespace: "digicloud-namespace"
  authSecretName: "digicloud-credentials"
```

## Troubleshooting

### Common Issues

1. **Certificate stuck in pending state**
   - Check issuer status: `kubectl describe digicloudissuer <name>`
   - Verify API credentials and permissions
   - Check cert-manager logs

2. **DNS01 challenge failures**
   - Ensure domain is properly configured in Digicloud
   - Verify API credentials have DNS management permissions
   - Check network connectivity to Digicloud API

3. **Rate limiting**
   - Digicloud API may have rate limits
   - Consider using staging ACME server for testing

### Logs

Check issuer logs:
```bash
kubectl logs -n digicloud-issuer-system deployment/digicloud-issuer-controller-manager
```

Check cert-manager logs:
```bash
kubectl logs -n cert-manager deployment/cert-manager
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Support

- GitHub Issues: [Report bugs and request features](https://github.com/digicloud/digicloud-issuer/issues)
- Documentation: [Full documentation](https://docs.digicloud.ir/certificates/)
- Community: [Digicloud Community Forum](https://community.digicloud.ir/)
