# GitHub Actions CI/CD for Digicloud Issuer

This directory contains GitHub Actions workflows for continuous integration and deployment of the Digicloud cert-manager external issuer.

## Workflows

### 1. CI Pipeline (`ci.yml`)
**Trigger:** Push to main/develop, Pull Requests to main, Releases

**Jobs:**
- **Test**: Runs unit tests, integration tests, and code generation checks
- **Security**: Runs Gosec security scanner
- **Build Image**: Builds and pushes Docker images to GitHub Container Registry
- **Deploy Manifests**: Generates and validates Kubernetes manifests
- **Integration Test**: Tests deployment in Kind cluster (PR only)
- **Release**: Creates release artifacts (release only)

**Artifacts:**
- Docker images tagged with branch/PR/release info
- Kubernetes manifests for deployment
- Test coverage reports

### 2. Security Audit (`security.yml`)
**Trigger:** Weekly schedule, Push to main, Pull Requests

**Jobs:**
- **Audit**: Runs Go vulnerability checks and dependency scanning
- **Trivy**: Scans for vulnerabilities using Trivy scanner
- **Nancy**: Scans dependencies for known vulnerabilities

**Outputs:**
- SARIF reports uploaded to GitHub Security tab
- Vulnerability alerts and reports

### 3. Container Security (`security-scan.yml`)
**Trigger:** Push to main, Pull Requests, Weekly schedule

**Jobs:**
- **Scan Image**: Scans built Docker images for vulnerabilities
- **Scan Published Image**: Scans published images (scheduled)
- **Scan Dockerfile**: Lints Dockerfile with Hadolint

**Features:**
- Multi-platform vulnerability scanning
- Dockerfile best practices validation
- Security alerts integration

### 4. Documentation (`docs.yml`)
**Trigger:** Changes to docs, README, API, or examples

**Jobs:**
- **Generate Docs**: Generates API documentation and validates examples
- **Deploy Docs**: Deploys documentation to GitHub Pages

**Features:**
- Automated API reference generation
- Example validation
- Markdown link checking
- GitHub Pages deployment

### 5. Auto-update Dependencies (`dependencies.yml`)
**Trigger:** Weekly schedule, Manual trigger

**Jobs:**
- **Update Dependencies**: Updates Go module dependencies and creates PR

**Features:**
- Automated dependency updates
- Test validation before PR creation
- Automated PR creation with changelog

### 6. Release (`release.yml`)
**Trigger:** Git tags (v*)

**Jobs:**
- **Release**: Creates GitHub release with changelog
- **Build and Push**: Builds multi-platform Docker images
- **Generate Manifests**: Creates deployment manifests for release

**Artifacts:**
- Multi-platform Docker images
- Combined Kubernetes manifest
- Release tarball with all manifests

### 7. Code Quality (`quality.yml`)
**Trigger:** Push to main/develop, Pull Requests

**Jobs:**
- **Quality Checks**: Code formatting, linting, and static analysis
- **Dependency Check**: Vulnerability scanning and dependency auditing
- **Documentation Check**: Documentation completeness and quality
- **License Check**: License header validation
- **Performance Check**: Benchmark running and binary size monitoring

## Configuration Files

### `.golangci.yml`
Golangci-lint configuration with enabled linters:
- errcheck, gofmt, goimports
- gosec, gosimple, govet
- ineffassign, misspell, staticcheck
- typecheck, unused

### `.github/markdown-link-check-config.json`
Configuration for markdown link checking:
- Ignores localhost URLs
- Configures timeout and retry settings
- Sets acceptable status codes

## Secrets Required

The workflows require the following secrets to be configured in the repository:

- `GITHUB_TOKEN` - Automatically provided by GitHub Actions
- No additional secrets required for basic functionality

## Deployment Flow

1. **Development**: 
   - Push to feature branch → CI pipeline runs
   - Create PR → Full CI + integration tests
   - Merge to main → Build and push images with `main` tag

2. **Release**:
   - Create git tag (e.g., `v1.0.0`) → Release workflow triggers
   - Builds multi-platform images
   - Creates GitHub release with manifests
   - Tags images with version numbers

3. **Security**:
   - Weekly scans of published images
   - Vulnerability alerts via GitHub Security tab
   - Automated dependency updates

## Container Registry

Images are published to GitHub Container Registry (ghcr.io):
- `ghcr.io/vamirreza/digicloud-issuer:main` - Latest main branch
- `ghcr.io/vamirreza/digicloud-issuer:v1.0.0` - Specific version
- `ghcr.io/vamirreza/digicloud-issuer:pr-123` - Pull request builds

## Monitoring and Alerts

- **Security**: SARIF reports uploaded to GitHub Security tab
- **Quality**: Code quality checks fail the build on issues
- **Dependencies**: Automated PRs for dependency updates
- **Documentation**: Link checking and completeness validation

## Manual Triggers

Some workflows can be triggered manually:
- **Dependencies**: Update dependencies via workflow dispatch
- **Documentation**: Regenerate and deploy documentation

## Local Development

To run similar checks locally:

```bash
# Code quality
make lint
make test
make coverage

# Security
govulncheck ./...
docker run --rm -v "$(pwd):/src" aquasec/trivy fs /src

# Documentation
markdownlint README.md
```

## Troubleshooting

Common issues and solutions:

1. **Test failures**: Check test logs and fix failing tests
2. **Security alerts**: Review and update vulnerable dependencies
3. **Build failures**: Check Go version compatibility and dependencies
4. **Image scanning**: Review Dockerfile for security best practices
5. **Release issues**: Ensure proper git tagging and release notes
