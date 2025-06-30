// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package v1alphav1

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	v1 "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
)

var (
	cfg          *rest.Config
	testEnv      *envtest.Environment
	scheme       = runtime.NewScheme()
	k8sClient    client.Client
	recorder     record.EventRecorder
	overcommitCR = &v1.Overcommit{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: v1.OvercommitSpec{
			OvercommitLabel: "inditex.com/overcommit-class",
		},
	}
)

func TestWebhookSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pod Mutating Webhook Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("POD_NAME", "webhook-server-mock")
	os.Setenv("POD_NAMESPACE", "webhook-server-mock-namespace")
	os.Setenv("OVERCOMMIT_CLASS_NAME", "default-overcommitclass")
	os.Setenv("LABEL_OVERCOMMIT_CLASS", "inditex.com/overcommit-class")

	log.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "..", "config", "crd", "bases"),
		},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	By("adding corev1 to scheme")
	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	By("adding OvercommitClass CRD to scheme")
	err = v1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	By("creating Kubernetes client")
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("creating a test overcommit")
	err = k8sClient.Create(context.TODO(), overcommitCR)
	Expect(err).NotTo(HaveOccurred())

	By("creating a default OvercommitClass resource")
	overcommitClass := &v1.OvercommitClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default-overcommitclass",
		},
		Spec: v1.OvercommitClassSpec{
			CpuOvercommit:      0.5,
			MemoryOvercommit:   0.5,
			ExcludedNamespaces: "kube-system",
			IsDefault:          false,
		},
	}

	err = k8sClient.Create(context.TODO(), overcommitClass)
	Expect(err).NotTo(HaveOccurred())

	By("initializing EventRecorder")
	mgr, err := manager.New(cfg, manager.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	recorder = mgr.GetEventRecorderFor("k8s-overcommit-operator")
	Expect(recorder).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	os.Unsetenv("POD_NAME")
	os.Unsetenv("POD_NAMESPACE")
	os.Unsetenv("OVERCOMMIT_CLASS_NAME")
	os.Unsetenv("LABEL_OVERCOMMIT_CLASS")

	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
