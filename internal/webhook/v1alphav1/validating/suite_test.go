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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
)

var (
	cfg            *rest.Config
	k8sClient      client.Client
	testEnv        *envtest.Environment
	testOvercommit = &overcommit.Overcommit{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: overcommit.OvercommitSpec{
			OvercommitLabel: "inditex.com/overcommit-class",
		},
	}
)

func TestWebhookSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validating Webhook Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("OVERCOMMIT_CLASS_NAME", "default-overcommitclass")
	os.Setenv("LABEL_OVERCOMMIT_CLASS", "inditex.com/overcommit-class")
	log.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("starting test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "..", "config", "crd", "bases"),
		},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())
	err = overcommit.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("creating a test overcommit")
	err = k8sClient.Create(context.TODO(), testOvercommit)
	Expect(err).NotTo(HaveOccurred())

	By("creating a default OvercommitClass resource")
	overcommitClass := &overcommit.OvercommitClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default-overcommitclass",
		},
		Spec: overcommit.OvercommitClassSpec{
			CpuOvercommit:      0.5,
			MemoryOvercommit:   0.5,
			ExcludedNamespaces: "kube-system",
			IsDefault:          true,
		},
	}

	err = k8sClient.Create(context.TODO(), overcommitClass)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	os.Unsetenv("OVERCOMMIT_CLASS_NAME")
	os.Unsetenv("LABEL_OVERCOMMIT_CLASS")

	By("stopping test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
