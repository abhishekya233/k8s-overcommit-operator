// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package v1alphav1

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("OvercommitClass Webhook", func() {
	var validator *OvercommitClassValidator

	BeforeEach(func() {
		validator = &OvercommitClassValidator{}
		validator.InjectClient(k8sClient)
	})

	AfterEach(func() {
		// Clean up all resources created during the test
		By("Cleaning up OvercommitClass resources")
		err := k8sClient.DeleteAllOf(context.TODO(), &OvercommitClass{})
		Expect(err).NotTo(HaveOccurred())
	})

	Context("ValidateCreate", func() {
		It("Should pass validation for a valid OvercommitClass", func() {
			overcommitClass := &OvercommitClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-overcommitclass",
				},
				Spec: OvercommitClassSpec{
					CpuOvercommit:      0.5,
					MemoryOvercommit:   0.5,
					ExcludedNamespaces: "kube-system",
					IsDefault:          false,
				},
			}

			warnings, err := validator.ValidateCreate(context.TODO(), overcommitClass)
			Expect(warnings).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should fail validation for invalid CPU overcommit", func() {
			overcommitClass := &OvercommitClass{
				Spec: OvercommitClassSpec{
					CpuOvercommit:      -0.5, // Invalid value
					MemoryOvercommit:   0.5,
					ExcludedNamespaces: "kube-system",
					IsDefault:          false,
				},
			}

			warnings, err := validator.ValidateCreate(context.TODO(), overcommitClass)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cpuOvercommit must be greater than 0 and equal or lower than 1, failed creating  class"))
		})

		It("Should fail validation for invalid memory overcommit", func() {
			overcommitClass := &OvercommitClass{
				Spec: OvercommitClassSpec{
					CpuOvercommit:      0.5,
					MemoryOvercommit:   -0.5, // Invalid value
					ExcludedNamespaces: "kube-system",
					IsDefault:          false,
				},
			}

			warnings, err := validator.ValidateCreate(context.TODO(), overcommitClass)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("memoryOvercommit must be greater than 0 and equal or lower than 1, failed creating  class"))
		})

		It("Should fail validation for invalid excluded namespaces", func() {
			overcommitClass := &OvercommitClass{
				Spec: OvercommitClassSpec{
					CpuOvercommit:      0.5,
					MemoryOvercommit:   0.5,
					ExcludedNamespaces: ".**./*/-*./../kube-system,invalid-namespace", // Invalid value
					IsDefault:          false,
				},
			}

			warnings, err := validator.ValidateCreate(context.TODO(), overcommitClass)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("regex"))
		})

	})

	Context("ValidateUpdate", func() {
		It("Should pass validation for a valid update", func() {
			oldOvercommitClass := &OvercommitClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-overcommitclass",
				},
				Spec: OvercommitClassSpec{
					CpuOvercommit:      0.5,
					MemoryOvercommit:   0.5,
					ExcludedNamespaces: "kube-system",
					IsDefault:          false,
				},
			}

			newOvercommitClass := &OvercommitClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-overcommitclass",
				},
				Spec: OvercommitClassSpec{
					CpuOvercommit:      0.7,
					MemoryOvercommit:   0.7,
					ExcludedNamespaces: "kube-system",
					IsDefault:          false,
				},
			}

			warnings, err := validator.ValidateUpdate(context.TODO(), oldOvercommitClass, newOvercommitClass)
			Expect(warnings).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should fail validation for invalid memory overcommit in update", func() {
			oldOvercommitClass := &OvercommitClass{
				Spec: OvercommitClassSpec{
					CpuOvercommit:      0.5,
					MemoryOvercommit:   0.5,
					ExcludedNamespaces: "kube-system",
				},
			}

			newOvercommitClass := &OvercommitClass{
				Spec: OvercommitClassSpec{
					CpuOvercommit:      0.7,
					MemoryOvercommit:   -0.7, // Invalid value
					ExcludedNamespaces: "kube-system",
				},
			}

			warnings, err := validator.ValidateUpdate(context.TODO(), oldOvercommitClass, newOvercommitClass)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("memoryOvercommit must be greater than 0 and equal or lower than 1, failed creating  class"))
		})

		It("Should fail validation for invalid memory overcommit in update", func() {
			oldOvercommitClass := &OvercommitClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: OvercommitClassSpec{
					CpuOvercommit:      0.5,
					MemoryOvercommit:   0.5,
					ExcludedNamespaces: "kube-system",
				},
			}

			newOvercommitClass := &OvercommitClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: OvercommitClassSpec{
					CpuOvercommit:      0.7,
					MemoryOvercommit:   -0.7, // Invalid value
					ExcludedNamespaces: "kube-system",
				},
			}

			warnings, err := validator.ValidateUpdate(context.TODO(), oldOvercommitClass, newOvercommitClass)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("memoryOvercommit must be greater than 0 and equal or lower than 1, failed creating test class"))
		})
	})

	Context("ValidateDelete", func() {
		It("Should pass validation for delete", func() {
			overcommitClass := &OvercommitClass{
				Spec: OvercommitClassSpec{
					CpuOvercommit:      0.5,
					MemoryOvercommit:   0.5,
					ExcludedNamespaces: "kube-system",
				},
			}

			warnings, err := validator.ValidateDelete(context.TODO(), overcommitClass)
			Expect(warnings).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
