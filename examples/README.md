# Example DigicloudIssuer and Certificate

This directory contains example manifests for using the Digicloud cert-manager external issuer.

## Prerequisites

1. Install cert-manager in your cluster
2. Install the Digicloud issuer CRDs and controller
3. Create a secret with your Digicloud API credentials

## Files

- `secret.yaml` - Example secret containing Digicloud API credentials
- `digicloud-issuer.yaml` - Example DigicloudIssuer resource
- `digicloud-cluster-issuer.yaml` - Example DigicloudClusterIssuer resource
- `certificate.yaml` - Example Certificate using the Digicloud issuer

## Usage

1. Update the secret with your actual Digicloud API token and namespace:
   ```bash
   kubectl apply -f secret.yaml
   ```

2. Create the issuer:
   ```bash
   kubectl apply -f digicloud-issuer.yaml
   ```

3. Create a certificate:
   ```bash
   kubectl apply -f certificate.yaml
   ```

The cert-manager will automatically handle the DNS01 challenge using the Digicloud DNS API.
