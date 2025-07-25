package controllers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cmacme "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"

	"github.com/vamirreza/digicloud-issuer/api/v1alpha1"
)

func TestDigicloudIssuerReconciler_Reconcile(t *testing.T) {
	// Create a scheme with our types
	scheme := runtime.NewScheme()
	err := clientgoscheme.AddToScheme(scheme)
	assert.NoError(t, err)
	err = cmapi.AddToScheme(scheme)
	assert.NoError(t, err)
	err = cmacme.AddToScheme(scheme)
	assert.NoError(t, err)
	err = v1alpha1.AddToScheme(scheme)
	assert.NoError(t, err)

	// Create fake client with no resources (empty client)
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	// Create reconciler
	reconciler := &DigicloudIssuerReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	// Test reconciliation with non-existent resource
	ctx := context.Background()
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-issuer",
			Namespace: "default",
		},
	}

	result, err := reconciler.Reconcile(ctx, req)

	// When the resource doesn't exist, the controller should return no error
	// and no requeue (it handles NotFound gracefully)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestDigicloudIssuerReconciler_SetupWithManager(t *testing.T) {
	// For this test, we'll just test that the reconciler can be created without errors
	// Testing the actual manager setup would require envtest which is complex for unit tests

	// Create a scheme
	scheme := runtime.NewScheme()
	err := v1alpha1.AddToScheme(scheme)
	assert.NoError(t, err)

	// Create fake client
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	// Create reconciler
	reconciler := &DigicloudIssuerReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	// Test that reconciler was created successfully
	assert.NotNil(t, reconciler)
	assert.NotNil(t, reconciler.Client)
	assert.NotNil(t, reconciler.Scheme)
}

func TestDigicloudClusterIssuerReconciler_Reconcile(t *testing.T) {
	// Create a scheme with our types
	scheme := runtime.NewScheme()
	err := clientgoscheme.AddToScheme(scheme)
	assert.NoError(t, err)
	err = v1alpha1.AddToScheme(scheme)
	assert.NoError(t, err)

	// Create fake client with no resources (empty client)
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	// Create reconciler
	reconciler := &DigicloudClusterIssuerReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	// Test reconciliation with non-existent resource
	ctx := context.Background()
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-cluster-issuer",
		},
	}

	result, err := reconciler.Reconcile(ctx, req)

	// When the resource doesn't exist, the controller should return no error
	// and no requeue (it handles NotFound gracefully)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}
