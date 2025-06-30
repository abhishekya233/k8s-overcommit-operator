// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package overcommit

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
)

var (
	cfg           *rest.Config
	k8sClient     client.Client
	testEnv       *envtest.Environment
	recorder      record.EventRecorder
	scheme        = runtime.NewScheme()
	testNamespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
			Labels: map[string]string{
				"inditex.com/overcommit-class": "test-class",
			},
		},
	}
	testPod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"inditex.com/overcommit-class": "test-class",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "test-container",
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
					Image: "nginx",
				},
			},
		},
	}
	testOvercommit = &overcommit.Overcommit{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: overcommit.OvercommitSpec{
			OvercommitLabel: "inditex.com/overcommit-class",
		},
	}
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Overcommit Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("OVERCOMMIT_CLASS_NAME", "test-class")
	os.Setenv("LABEL_OVERCOMMIT_CLASS", "inditex.com/overcommit-class")

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			"../../config/crd/bases",
		},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	By("adding corev1 to scheme")
	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	By("creating Kubernetes client")
	err = overcommit.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("initializing EventRecorder")
	mgr, err := manager.New(cfg, manager.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	recorder = mgr.GetEventRecorderFor("k8s-overcommit-operator")
	Expect(recorder).NotTo(BeNil())

	By("creating a test overcommit")
	err = k8sClient.Create(context.TODO(), testOvercommit)
	Expect(err).NotTo(HaveOccurred())

	By("creating a default OvercommitClass resource")
	overcommitClass := &overcommit.OvercommitClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-class",
		},
		Spec: overcommit.OvercommitClassSpec{
			CpuOvercommit:      0.5,
			MemoryOvercommit:   0.5,
			ExcludedNamespaces: "kube-system",
			IsDefault:          true,
		},
	}

	By("creating a test namespace")
	err = k8sClient.Create(context.TODO(), testNamespace)
	Expect(err).NotTo(HaveOccurred())

	By("creating a test pod")
	err = k8sClient.Create(context.TODO(), testPod)
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Create(context.TODO(), overcommitClass)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	os.Unsetenv("OVERCOMMIT_CLASS_NAME")
	os.Unsetenv("LABEL_OVERCOMMIT_CLASS")
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
