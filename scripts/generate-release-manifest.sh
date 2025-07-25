#!/bin/bash

# Script to generate release manifests
set -e

REGISTRY=${REGISTRY:-"ghcr.io"}
IMAGE_NAME=${IMAGE_NAME:-"vamirreza/digicloudissuer"}
TAG=${TAG:-"latest"}

echo "Generating release manifests..."
echo "Registry: $REGISTRY"
echo "Image: $IMAGE_NAME"
echo "Tag: $TAG"

# Ensure we have the latest manifests
make manifests

# Create release manifests directory
mkdir -p release-manifests

# Create combined manifest
cat > release-manifests/digicloud-issuer.yaml << EOF
# Digicloud Issuer - $TAG
# Generated on $(date)
---
EOF

# Add CRDs first
echo "Adding CRDs..."
for crd in config/crd/bases/*.yaml; do
  if [ -f "$crd" ]; then
    cat "$crd" >> release-manifests/digicloud-issuer.yaml
    echo "---" >> release-manifests/digicloud-issuer.yaml
  fi
done

# Add namespace
echo "Adding namespace..."
cat >> release-manifests/digicloud-issuer.yaml << EOF
apiVersion: v1
kind: Namespace
metadata:
  name: digicloud-issuer-system
---
EOF

# Add RBAC resources (excluding kustomization files)
echo "Adding RBAC..."
for rbac in config/rbac/*.yaml; do
  if [ -f "$rbac" ] && [[ ! "$rbac" =~ kustomization\.yaml$ ]]; then
    cat "$rbac" >> release-manifests/digicloud-issuer.yaml
    echo "---" >> release-manifests/digicloud-issuer.yaml
  fi
done

# Add manager deployment (excluding kustomization files)
echo "Adding manager deployment..."
for manager in config/manager/*.yaml; do
  if [ -f "$manager" ] && [[ ! "$manager" =~ kustomization\.yaml$ ]]; then
    # Update the image tag and fix namespace
    sed -e "s|image: .*|image: $REGISTRY/$IMAGE_NAME:$TAG|" \
        -e "s|namespace: system|namespace: digicloud-issuer-system|" \
        "$manager" >> release-manifests/digicloud-issuer.yaml
    echo "---" >> release-manifests/digicloud-issuer.yaml
  fi
done

# Remove the last separator if it exists
sed -i '' '$ { /^---$/d; }' release-manifests/digicloud-issuer.yaml

echo "✅ Release manifest generated: release-manifests/digicloud-issuer.yaml"
echo "📦 Validating manifest..."

# Validate the manifest
kubectl apply --dry-run=client -f release-manifests/digicloud-issuer.yaml

echo "✅ Manifest validation successful!"
