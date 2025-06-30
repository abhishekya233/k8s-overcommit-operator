// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package overcommit

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Overcommit Functions", func() {

	Describe("getNamespaceOvercommit", func() {
		It("should return the correct overcommit values from the namespace", func() {
			cpuOvercommit, memoryOvercommit := getNamespaceOvercommit(context.TODO(), testPod, k8sClient, "inditex.com/overcommit-class", "ownerName", "ownerKind")
			Expect(cpuOvercommit).To(Equal(0.5))
			Expect(memoryOvercommit).To(Equal(0.5))
		})
	})

	Describe("checkOvercommitType", func() {
		It("should return the correct overcommit values from the pod", func() {
			cpuOvercommit, memoryOvercommit := checkOvercommitType(context.TODO(), *testPod, k8sClient)
			Expect(cpuOvercommit).To(Equal(0.5))
			Expect(memoryOvercommit).To(Equal(0.5))
		})

		It("should fallback to namespace overcommit values if pod label is missing", func() {
			delete(testPod.Labels, "LABEL_OVERCOMMIT_CLASS")
			err := k8sClient.Update(context.TODO(), testPod)
			Expect(err).NotTo(HaveOccurred())

			cpuOvercommit, memoryOvercommit := checkOvercommitType(context.TODO(), *testPod, k8sClient)
			Expect(cpuOvercommit).To(Equal(0.5))
			Expect(memoryOvercommit).To(Equal(0.5))
		})
	})
})
