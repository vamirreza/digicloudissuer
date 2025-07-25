package integration

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	cmacme "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"

	"github.com/vamirreza/digicloud-issuer/api/v1alpha1"
	"github.com/vamirreza/digicloud-issuer/internal/controllers"
)

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Test Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = v1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = cmapi.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = cmacme.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	// Start the manager
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
		Metrics: server.Options{
			BindAddress: "0",
		},
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.DigicloudIssuerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.DigicloudClusterIssuerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("DigicloudIssuer", func() {
	var (
		namespace  string
		issuerName string
		issuer     *v1alpha1.DigicloudIssuer
	)

	BeforeEach(func() {
		namespace = "default"
		issuerName = "test-issuer-" + time.Now().Format("20060102150405")

		issuer = &v1alpha1.DigicloudIssuer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      issuerName,
				Namespace: namespace,
			},
			Spec: v1alpha1.DigicloudIssuerSpec{
				Provisioner: v1alpha1.DigicloudIssuerProvisioner{
					APIBaseURL: "https://api.digicloud.ir",
					APITokenSecretRef: v1alpha1.SecretKeySelector{
						Name: "api-key-secret",
						Key:  "api-key",
					},
				},
			},
		}
	})

	Context("When creating a DigicloudIssuer", func() {
		It("Should be created successfully", func() {
			Expect(k8sClient.Create(ctx, issuer)).Should(Succeed())

			// Verify the issuer was created
			createdIssuer := &v1alpha1.DigicloudIssuer{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(issuer), createdIssuer)
			}, time.Second*10, time.Millisecond*250).Should(Succeed())

			Expect(createdIssuer.Spec.Provisioner.APIBaseURL).To(Equal("https://api.digicloud.ir"))
			Expect(createdIssuer.Spec.Provisioner.APITokenSecretRef.Name).To(Equal("api-key-secret"))
		})

		It("Should be deletable", func() {
			Expect(k8sClient.Create(ctx, issuer)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, issuer)).Should(Succeed())

			// Verify the issuer was deleted
			deletedIssuer := &v1alpha1.DigicloudIssuer{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(issuer), deletedIssuer)
				return err != nil
			}, time.Second*10, time.Millisecond*250).Should(BeTrue())
		})
	})
})

var _ = Describe("DigicloudClusterIssuer", func() {
	var (
		clusterIssuerName string
		clusterIssuer     *v1alpha1.DigicloudClusterIssuer
	)

	BeforeEach(func() {
		clusterIssuerName = "test-cluster-issuer-" + time.Now().Format("20060102150405")

		clusterIssuer = &v1alpha1.DigicloudClusterIssuer{
			ObjectMeta: metav1.ObjectMeta{
				Name: clusterIssuerName,
			},
			Spec: v1alpha1.DigicloudClusterIssuerSpec{
				Provisioner: v1alpha1.DigicloudIssuerProvisioner{
					APIBaseURL: "https://api.digicloud.ir",
					APITokenSecretRef: v1alpha1.SecretKeySelector{
						Name: "api-key-secret",
						Key:  "api-key",
					},
				},
			},
		}
	})

	Context("When creating a DigicloudClusterIssuer", func() {
		It("Should be created successfully", func() {
			Expect(k8sClient.Create(ctx, clusterIssuer)).Should(Succeed())

			// Verify the cluster issuer was created
			createdClusterIssuer := &v1alpha1.DigicloudClusterIssuer{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(clusterIssuer), createdClusterIssuer)
			}, time.Second*10, time.Millisecond*250).Should(Succeed())

			Expect(createdClusterIssuer.Spec.Provisioner.APIBaseURL).To(Equal("https://api.digicloud.ir"))
			Expect(createdClusterIssuer.Spec.Provisioner.APITokenSecretRef.Name).To(Equal("api-key-secret"))
		})

		AfterEach(func() {
			// Clean up the cluster issuer
			Expect(k8sClient.Delete(ctx, clusterIssuer)).Should(Succeed())
		})
	})
})
