/*
Copyright 2025 Digicloud.

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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	digicloudv1alpha1 "github.com/vamirreza/digicloud-issuer/api/v1alpha1"
)

// CertificateRequestReconciler reconciles a CertificateRequest object
type CertificateRequestReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=cert-manager.io,resources=certificaterequests,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificaterequests/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CertificateRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Get the CertificateRequest
	var cr cmapi.CertificateRequest
	if err := r.Get(ctx, req.NamespacedName, &cr); err != nil {
		log.Error(err, "unable to fetch CertificateRequest")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if this is a DigicloudIssuer or DigicloudClusterIssuer request
	if !r.isDigicloudIssuer(cr.Spec.IssuerRef) {
		log.V(1).Info("CertificateRequest is not for a Digicloud issuer, ignoring")
		return ctrl.Result{}, nil
	}

	// Check if already processed
	if cr.Status.Certificate != nil {
		log.V(1).Info("CertificateRequest is already complete")
		return ctrl.Result{}, nil
	}

	// Check if failed
	if r.hasFailedCondition(&cr) {
		log.V(1).Info("CertificateRequest has failed")
		return ctrl.Result{}, nil
	}

	log.Info("Processing CertificateRequest", "name", cr.Name, "namespace", cr.Namespace)

	// Get the issuer
	_, err := r.getIssuer(ctx, cr.Spec.IssuerRef, cr.Namespace)
	if err != nil {
		log.Error(err, "failed to get issuer")
		r.setStatus(ctx, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonPending, "Failed to get issuer: "+err.Error())
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Parse the CSR
	csr, err := r.parseCSR(cr.Spec.Request)
	if err != nil {
		log.Error(err, "failed to parse CSR")
		r.setStatus(ctx, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonFailed, "Failed to parse CSR: "+err.Error())
		return ctrl.Result{}, nil
	}

	// Check if we've already processed this request successfully
	if r.isAlreadyIssued(&cr) {
		log.Info("Certificate already issued, skipping")
		return ctrl.Result{}, nil
	}

	// For demonstration purposes, we'll create a mock certificate
	// In a real implementation, you would:
	// 1. Create DNS TXT records for ACME challenge using Digicloud API
	// 2. Wait for DNS propagation
	// 3. Complete ACME challenge
	// 4. Get the signed certificate from ACME server

	log.Info("CertificateRequest received and being processed", "domains", csr.DNSNames)

	// Check if this is the first time we're processing this request
	if !r.hasProcessingCondition(&cr) {
		log.Info("Starting DNS validation process")
		r.setStatus(ctx, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonPending, "DNS validation in progress")
		r.Recorder.Event(&cr, "Normal", "Processing", "Starting DNS validation for certificate request")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Simulate DNS validation completion after some time
	// In a real implementation, you would check with Digicloud API
	if r.shouldCompleteValidation(&cr) {
		log.Info("DNS validation completed, issuing certificate")

		// Generate a mock certificate for testing
		cert, err := r.generateMockCertificate(csr)
		if err != nil {
			log.Error(err, "failed to generate mock certificate")
			r.setStatus(ctx, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonFailed, "Failed to generate certificate: "+err.Error())
			return ctrl.Result{}, nil
		}

		// Set the certificate in the status
		cr.Status.Certificate = cert
		r.setStatus(ctx, &cr, cmmeta.ConditionTrue, cmapi.CertificateRequestReasonIssued, "Certificate issued successfully")
		r.Recorder.Event(&cr, "Normal", "Issued", "Certificate issued successfully")

		return ctrl.Result{}, nil
	}

	// Continue waiting for DNS validation
	log.Info("DNS validation still in progress")
	r.setStatus(ctx, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonPending, "DNS validation in progress")

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// isDigicloudIssuer checks if the issuer reference is for a Digicloud issuer
func (r *CertificateRequestReconciler) isDigicloudIssuer(ref cmmeta.ObjectReference) bool {
	return ref.Group == digicloudv1alpha1.GroupVersion.Group &&
		(ref.Kind == "DigicloudIssuer" || ref.Kind == "DigicloudClusterIssuer")
}

// hasFailedCondition checks if the CertificateRequest has a failed condition
func (r *CertificateRequestReconciler) hasFailedCondition(cr *cmapi.CertificateRequest) bool {
	for _, condition := range cr.Status.Conditions {
		if condition.Type == cmapi.CertificateRequestConditionReady &&
			condition.Status == cmmeta.ConditionFalse &&
			condition.Reason == cmapi.CertificateRequestReasonFailed {
			return true
		}
	}
	return false
}

// getIssuer retrieves the DigicloudIssuer or DigicloudClusterIssuer
func (r *CertificateRequestReconciler) getIssuer(ctx context.Context, ref cmmeta.ObjectReference, namespace string) (client.Object, error) {
	if ref.Kind == "DigicloudIssuer" {
		var issuer digicloudv1alpha1.DigicloudIssuer
		err := r.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: namespace}, &issuer)
		return &issuer, err
	} else if ref.Kind == "DigicloudClusterIssuer" {
		var issuer digicloudv1alpha1.DigicloudClusterIssuer
		err := r.Get(ctx, types.NamespacedName{Name: ref.Name}, &issuer)
		return &issuer, err
	}
	return nil, fmt.Errorf("unknown issuer kind: %s", ref.Kind)
}

// parseCSR parses the certificate signing request
func (r *CertificateRequestReconciler) parseCSR(data []byte) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate request: %w", err)
	}

	return csr, nil
}

// setStatus updates the CertificateRequest status
func (r *CertificateRequestReconciler) setStatus(ctx context.Context, cr *cmapi.CertificateRequest, status cmmeta.ConditionStatus, reason, message string) {
	now := metav1.Now()

	// Find existing condition or create new one
	var condition *cmapi.CertificateRequestCondition
	for i := range cr.Status.Conditions {
		if cr.Status.Conditions[i].Type == cmapi.CertificateRequestConditionReady {
			condition = &cr.Status.Conditions[i]
			break
		}
	}

	if condition == nil {
		// Add new condition
		cr.Status.Conditions = append(cr.Status.Conditions, cmapi.CertificateRequestCondition{
			Type:               cmapi.CertificateRequestConditionReady,
			Status:             status,
			Reason:             reason,
			Message:            message,
			LastTransitionTime: &now,
		})
	} else {
		// Update existing condition
		condition.Status = status
		condition.Reason = reason
		condition.Message = message
		condition.LastTransitionTime = &now
	}

	if err := r.Status().Update(ctx, cr); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed to update CertificateRequest status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CertificateRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cmapi.CertificateRequest{}).
		Complete(r)
}

// isAlreadyIssued checks if the CertificateRequest has already been issued
func (r *CertificateRequestReconciler) isAlreadyIssued(cr *cmapi.CertificateRequest) bool {
	for _, condition := range cr.Status.Conditions {
		if condition.Type == cmapi.CertificateRequestConditionReady &&
			condition.Status == cmmeta.ConditionTrue &&
			condition.Reason == cmapi.CertificateRequestReasonIssued {
			return true
		}
	}
	return len(cr.Status.Certificate) > 0
}

// hasProcessingCondition checks if the CertificateRequest has a processing condition
func (r *CertificateRequestReconciler) hasProcessingCondition(cr *cmapi.CertificateRequest) bool {
	for _, condition := range cr.Status.Conditions {
		if condition.Type == cmapi.CertificateRequestConditionReady &&
			condition.Reason == cmapi.CertificateRequestReasonPending {
			return true
		}
	}
	return false
}

// shouldCompleteValidation determines if enough time has passed to complete validation
// In a real implementation, this would check with the Digicloud API
func (r *CertificateRequestReconciler) shouldCompleteValidation(cr *cmapi.CertificateRequest) bool {
	// For testing, complete validation after 2 minutes
	for _, condition := range cr.Status.Conditions {
		if condition.Type == cmapi.CertificateRequestConditionReady &&
			condition.Reason == cmapi.CertificateRequestReasonPending &&
			condition.LastTransitionTime != nil {
			elapsed := time.Since(condition.LastTransitionTime.Time)
			return elapsed > 30*time.Second
		}
	}
	return false
}

// generateMockCertificate creates a mock certificate for testing purposes
// In a real implementation, this would be replaced with actual certificate from ACME server
func (r *CertificateRequestReconciler) generateMockCertificate(csr *x509.CertificateRequest) ([]byte, error) {
	// This is a mock implementation for testing
	// In reality, you would get the certificate from your ACME provider

	mockCert := `-----BEGIN CERTIFICATE-----
MIIDQTCCAimgAwIBAgITBmyfz5m/jAo54vB4ikPmljZbyjANBgkqhkiG9w0BAQsF
ADA5MQswCQYDVQQGEwJVUzEPMA0GA1UEChMGQW1hem9uMRkwFwYDVQQDExBBbWF6
b24gUm9vdCBDQSAxMB4XDTE1MDUyNjAwMDAwMFoXDTM4MDExNzAwMDAwMFowOTEL
MAkGA1UEBhMCVVMxDzANBgNVBAoTBkFtYXpvbjEZMBcGA1UEAxMQQW1hem9uIFJv
b3QgQ0EgMTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALJ4gHHKeNXj
ca9HgFB0fW7Y14h29Jlo91ghYPl0hAEvrAIthtOgQ3pOsqTQNroBvo3bSMgHFzZM
9O6II8c+6zf1tRn4SWiw3te5djgdYZ6k/oI2peVKVuRF4fn9tBb6dNqcmzU5L/qw
IFAGbHrQgLKm+a/sRxmPUDgH3KKHOVj4utWp+UhnMJbulHheb4mjUcAwhmahRWa6
VOujw5H5SNz/0egwLX0tdHA114gk957EWW67c4cX8jJGKLhD+rcdqsq08p8kDi1L
93FcXmn/6pUCyziKrlA4b9v7LWIbxcceVOF34GfID5yHI9Y/QCB/IIDEgEw+OyQm
jgSubJrIqg0CAwEAAaNCMEAwDwYDVR0TAQH/BAUwAwEB/zAOBgNVHQ8BAf8EBAMC
AYYwHQYDVR0OBBYEFIQYzIU07LwMlJQuCFmcx7IQTgoIMA0GCSqGSIb3DQEBCwUA
A4IBAQCY8jdaQZChGsV2USggNiMOruYou6r4lK5IpDB/G/wkjUu0yKGX9rbxenDI
U5PMCCjjmCXPI6T53iHTfIuJruydjsw2hUwsqdnlQkOYjPRi7vV+BwlEEPWmJNrA
VA8NvJsH4jfGZz8xTFdJcCQ5YNVWOa1Fs0d5MFRe1YOJZnFfJwStMVDjcJXpJPRf
AXhiCxCKrWX8f9KACF37CfFT0PVn9rYI5jh5kHPvHPe2Sw5qF/kKUGwOFNn6XwUx
JNjaMjIGZPgJVCB0hhGsXRBCdEZOlJuUTp7xt9bPlRi5JrKx8YOC8XBM2HTwZt1u
mFHZ9rZO8P1oSGOB0XDFQF6WHTzD
-----END CERTIFICATE-----`

	return []byte(mockCert), nil
}
