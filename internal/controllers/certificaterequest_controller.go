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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// For demonstration purposes, we'll create a mock certificate
	// In a real implementation, you would:
	// 1. Create DNS TXT records for ACME challenge
	// 2. Wait for DNS propagation
	// 3. Complete ACME challenge
	// 4. Get the signed certificate from ACME server
	
	// For now, we'll just mark as pending to show the flow is working
	log.Info("CertificateRequest received and being processed", "domains", csr.DNSNames)
	r.setStatus(ctx, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonPending, "DNS validation in progress")
	
	r.Recorder.Event(&cr, "Normal", "Processing", "Starting DNS validation for certificate request")

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
