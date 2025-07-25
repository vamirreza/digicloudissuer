/*
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
*/

package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/cert-manager/issuer-lib/controllers/signer"

	digicloudv1alpha1 "github.com/vamirreza/digicloud-issuer/api/v1alpha1"
	"github.com/vamirreza/digicloud-issuer/internal/dnsprovider"
)

// DigicloudIssuerReconciler reconciles a DigicloudIssuer object
type DigicloudIssuerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=digicloud.issuer.vamirreza.github.io,resources=digicloudissuers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=digicloud.issuer.vamirreza.github.io,resources=digicloudissuers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=digicloud.issuer.vamirreza.github.io,resources=digicloudissuers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *DigicloudIssuerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the DigicloudIssuer instance
	var issuer digicloudv1alpha1.DigicloudIssuer
	if err := r.Get(ctx, req.NamespacedName, &issuer); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("DigicloudIssuer resource not found, ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get DigicloudIssuer")
		return ctrl.Result{}, err
	}

	// Initialize status conditions if not set
	if issuer.Status.Conditions == nil {
		issuer.Status.Conditions = []cmapi.IssuerCondition{}
	}

	// Validate the issuer configuration
	if err := r.validateIssuer(ctx, &issuer); err != nil {
		logger.Error(err, "Invalid issuer configuration")
		r.setReadyCondition(&issuer, "Failed", err.Error())
		if statusErr := r.Status().Update(ctx, &issuer); statusErr != nil {
			logger.Error(statusErr, "Failed to update status")
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{}, err
	}

	// Set ready condition
	r.setReadyCondition(&issuer, "Checked", "Issuer configuration is valid")
	if err := r.Status().Update(ctx, &issuer); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	logger.Info("DigicloudIssuer reconciled successfully")
	return ctrl.Result{}, nil
}

// validateIssuer validates the issuer configuration
func (r *DigicloudIssuerReconciler) validateIssuer(ctx context.Context, issuer *digicloudv1alpha1.DigicloudIssuer) error {
	// Validate API token secret reference
	secretName := issuer.Spec.Provisioner.APITokenSecretRef.Name
	secretKey := issuer.Spec.Provisioner.APITokenSecretRef.Key

	if secretName == "" || secretKey == "" {
		return fmt.Errorf("API token secret reference must specify both name and key")
	}

	// Check if the secret exists
	var secret corev1.Secret
	secretNamespacedName := types.NamespacedName{
		Name:      secretName,
		Namespace: issuer.Namespace,
	}

	if err := r.Get(ctx, secretNamespacedName, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("API token secret %s not found in namespace %s", secretName, issuer.Namespace)
		}
		return fmt.Errorf("failed to get API token secret: %w", err)
	}

	// Check if the secret contains the specified key
	if _, exists := secret.Data[secretKey]; !exists {
		return fmt.Errorf("API token secret %s does not contain key %s", secretName, secretKey)
	}

	return nil
}

// setReadyCondition sets the Ready condition on the issuer
func (r *DigicloudIssuerReconciler) setReadyCondition(issuer *digicloudv1alpha1.DigicloudIssuer, reason, message string) {
	status := cmmeta.ConditionTrue
	if reason == "Failed" {
		status = cmmeta.ConditionFalse
	}

	// Find existing condition
	for i, condition := range issuer.Status.Conditions {
		if condition.Type == cmapi.IssuerConditionReady {
			issuer.Status.Conditions[i].Status = status
			issuer.Status.Conditions[i].Reason = reason
			issuer.Status.Conditions[i].Message = message
			now := metav1.Now()
			issuer.Status.Conditions[i].LastTransitionTime = &now
			return
		}
	}

	// Add new condition if not found
	now := metav1.Now()
	issuer.Status.Conditions = append(issuer.Status.Conditions, cmapi.IssuerCondition{
		Type:               cmapi.IssuerConditionReady,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: &now,
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *DigicloudIssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&digicloudv1alpha1.DigicloudIssuer{}).
		Complete(r)
}

// DigicloudClusterIssuerReconciler reconciles a DigicloudClusterIssuer object
type DigicloudClusterIssuerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=digicloud.issuer.vamirreza.github.io,resources=digicloudclusterissuers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=digicloud.issuer.vamirreza.github.io,resources=digicloudclusterissuers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=digicloud.issuer.vamirreza.github.io,resources=digicloudclusterissuers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop for cluster issuers
func (r *DigicloudClusterIssuerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the DigicloudClusterIssuer instance
	var issuer digicloudv1alpha1.DigicloudClusterIssuer
	if err := r.Get(ctx, req.NamespacedName, &issuer); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("DigicloudClusterIssuer resource not found, ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get DigicloudClusterIssuer")
		return ctrl.Result{}, err
	}

	// Initialize status conditions if not set
	if issuer.Status.Conditions == nil {
		issuer.Status.Conditions = []cmapi.IssuerCondition{}
	}

	// Validate the cluster issuer configuration
	if err := r.validateClusterIssuer(ctx, &issuer); err != nil {
		logger.Error(err, "Invalid cluster issuer configuration")
		r.setClusterReadyCondition(&issuer, "Failed", err.Error())
		if statusErr := r.Status().Update(ctx, &issuer); statusErr != nil {
			logger.Error(statusErr, "Failed to update status")
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{}, err
	}

	// Set ready condition
	r.setClusterReadyCondition(&issuer, "Checked", "Cluster issuer configuration is valid")
	if err := r.Status().Update(ctx, &issuer); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	logger.Info("DigicloudClusterIssuer reconciled successfully")
	return ctrl.Result{}, nil
}

// validateClusterIssuer validates the cluster issuer configuration
func (r *DigicloudClusterIssuerReconciler) validateClusterIssuer(ctx context.Context, issuer *digicloudv1alpha1.DigicloudClusterIssuer) error {
	// For cluster issuers, we need to look for secrets in a specific namespace
	// This is typically controlled by configuration, but for now we'll use a default
	secretNamespace := "digicloud-issuer-system" // TODO: Make this configurable

	secretName := issuer.Spec.Provisioner.APITokenSecretRef.Name
	secretKey := issuer.Spec.Provisioner.APITokenSecretRef.Key

	if secretName == "" || secretKey == "" {
		return fmt.Errorf("API token secret reference must specify both name and key")
	}

	// Check if the secret exists
	var secret corev1.Secret
	secretNamespacedName := types.NamespacedName{
		Name:      secretName,
		Namespace: secretNamespace,
	}

	if err := r.Get(ctx, secretNamespacedName, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("API token secret %s not found in namespace %s", secretName, secretNamespace)
		}
		return fmt.Errorf("failed to get API token secret: %w", err)
	}

	// Check if the secret contains the specified key
	if _, exists := secret.Data[secretKey]; !exists {
		return fmt.Errorf("API token secret %s does not contain key %s", secretName, secretKey)
	}

	return nil
}

// setClusterReadyCondition sets the Ready condition on the cluster issuer
func (r *DigicloudClusterIssuerReconciler) setClusterReadyCondition(issuer *digicloudv1alpha1.DigicloudClusterIssuer, reason, message string) {
	status := cmmeta.ConditionTrue
	if reason == "Failed" {
		status = cmmeta.ConditionFalse
	}

	// Find existing condition
	for i, condition := range issuer.Status.Conditions {
		if condition.Type == cmapi.IssuerConditionReady {
			issuer.Status.Conditions[i].Status = status
			issuer.Status.Conditions[i].Reason = reason
			issuer.Status.Conditions[i].Message = message
			now := metav1.Now()
			issuer.Status.Conditions[i].LastTransitionTime = &now
			return
		}
	}

	// Add new condition if not found
	now := metav1.Now()
	issuer.Status.Conditions = append(issuer.Status.Conditions, cmapi.IssuerCondition{
		Type:               cmapi.IssuerConditionReady,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: &now,
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *DigicloudClusterIssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&digicloudv1alpha1.DigicloudClusterIssuer{}).
		Complete(r)
}

// DigicloudSigner implements the cert-manager issuer-lib signer interface
type DigicloudSigner struct {
	issuerSpec      digicloudv1alpha1.DigicloudIssuerProvisioner
	secretNamespace string
	client          client.Client
}

// NewDigicloudSigner creates a new Digicloud signer
func NewDigicloudSigner(client client.Client, issuerSpec digicloudv1alpha1.DigicloudIssuerProvisioner, secretNamespace string) *DigicloudSigner {
	return &DigicloudSigner{
		issuerSpec:      issuerSpec,
		secretNamespace: secretNamespace,
		client:          client,
	}
}

// Sign signs a certificate request using the Digicloud DNS provider for DNS01 challenges
func (s *DigicloudSigner) Sign(ctx context.Context, cr signer.CertificateRequestObject, issuerObj client.Object) (signer.PEMBundle, error) {
	logger := log.FromContext(ctx)

	// Get the API token from the secret
	apiToken, namespace, err := s.getAPIToken(ctx, issuerObj)
	if err != nil {
		return signer.PEMBundle{}, fmt.Errorf("failed to get API token: %w", err)
	}

	// TODO: Get the namespace from the issuer configuration
	digicloudNamespace := namespace // This should be configured in the issuer spec

	// Create the DNS provider
	_ = dnsprovider.NewDigicloudProvider(
		s.issuerSpec.APIBaseURL,
		apiToken,
		digicloudNamespace,
		s.getTTL(),
	)

	logger.Info("Digicloud signer created successfully")

	// TODO: Implement actual certificate signing logic using ACME with DNS01 challenges
	// This would involve:
	// 1. Creating an ACME client
	// 2. Registering the Digicloud DNS provider for DNS01 challenges
	// 3. Requesting a certificate from the ACME server
	// 4. Returning the signed certificate

	// For now, return an error indicating this is not yet implemented
	return signer.PEMBundle{}, fmt.Errorf("certificate signing not yet implemented")
}

// getAPIToken retrieves the API token from the Kubernetes secret
func (s *DigicloudSigner) getAPIToken(ctx context.Context, issuerObj client.Object) (string, string, error) {
	secretName := s.issuerSpec.APITokenSecretRef.Name
	secretKey := s.issuerSpec.APITokenSecretRef.Key

	var secretNamespace string
	if s.secretNamespace != "" {
		secretNamespace = s.secretNamespace
	} else {
		// For namespaced issuers, use the issuer's namespace
		secretNamespace = issuerObj.GetNamespace()
	}

	var secret corev1.Secret
	secretNamespacedName := types.NamespacedName{
		Name:      secretName,
		Namespace: secretNamespace,
	}

	if err := s.client.Get(ctx, secretNamespacedName, &secret); err != nil {
		return "", "", fmt.Errorf("failed to get secret %s/%s: %w", secretNamespace, secretName, err)
	}

	apiTokenBytes, exists := secret.Data[secretKey]
	if !exists {
		return "", "", fmt.Errorf("secret %s/%s does not contain key %s", secretNamespace, secretName, secretKey)
	}

	// Look for namespace key in the secret, otherwise use a default
	namespaceBytes, exists := secret.Data["namespace"]
	namespace := "default"
	if exists {
		namespace = string(namespaceBytes)
	}

	return string(apiTokenBytes), namespace, nil
}

// getTTL returns the TTL for DNS records
func (s *DigicloudSigner) getTTL() int {
	if s.issuerSpec.TTL != nil {
		return *s.issuerSpec.TTL
	}
	return 300 // Default TTL
}
