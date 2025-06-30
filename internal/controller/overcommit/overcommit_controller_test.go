// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"os"
	"time"

	overcommitv1 "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Overcommit Controller", func() {
	const (
		OvercommitName      = "cluster"
		OvercommitNamespace = "k8s-overcommit"
		timeout             = time.Second * 10
		interval            = time.Millisecond * 250
	)

	Context("When creating Overcommit CR", func() {
		It("Should create related resources", func() {
			By("Creating a new Overcommit CR")
			ctx := context.Background()
			overcommit := &overcommitv1.Overcommit{
				ObjectMeta: metav1.ObjectMeta{
					Name:      OvercommitName,
					Namespace: OvercommitNamespace,
				},
				Spec: overcommitv1.OvercommitSpec{
					Labels: map[string]string{
						"app": "test-app",
					},
					Annotations: map[string]string{
						"app": "test-app",
					},
					OvercommitLabel: "test-overcommit",
				},
			}
			Expect(k8sClient.Create(ctx, overcommit)).Should(Succeed())
			// Create the namespace
			Expect(k8sClient.Create(ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: os.Getenv("POD_NAMESPACE"),
				},
			}))

			By("Expecting Issuer to be created")
			issuer := &certmanagerv1.Issuer{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: "k8s-overcommit-issuer", Namespace: OvercommitNamespace}, issuer)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Expecting OvercommitClassValidating Deployment to be created")
			deploy := &appsv1.Deployment{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: "k8s-overcommit-class-validating-webhook", Namespace: OvercommitNamespace}, deploy)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Expecting OvercommitClassValidating Service to be created")
			svc := &corev1.Service{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: "k8s-overcommit-class-validating-webhook-service", Namespace: OvercommitNamespace}, svc)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Expecting OvercommitClassValidating WebhookConfiguration to be created")
			webhook := &admissionv1.ValidatingWebhookConfiguration{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: "k8s-overcommit-class-validating-webhook", Namespace: ""}, webhook)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Expecting OvercommitClass Controller Deployment to be created")
			occontroller := &appsv1.Deployment{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: "k8s-overcommit-overcommitclass-controller", Namespace: OvercommitNamespace}, occontroller)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Expecting Pod Validating Deployment to be created")
			podWebhookDeploy := &appsv1.Deployment{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: "k8s-overcommit-pod-validating-webhook", Namespace: OvercommitNamespace}, podWebhookDeploy)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Expecting Pod Validating Service to be created")
			podWebhookSvc := &corev1.Service{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: "k8s-overcommit-pod-validating-webhook-service", Namespace: OvercommitNamespace}, podWebhookSvc)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Expecting Pod ValidatingWebhookConfiguration to be created")
			podWebhookConfig := &admissionv1.ValidatingWebhookConfiguration{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: "k8s-overcommit-pod-validating-webhook", Namespace: ""}, podWebhookConfig)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, overcommit)).Should(Succeed())
		})
	})
})
