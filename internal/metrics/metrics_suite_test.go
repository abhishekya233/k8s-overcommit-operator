// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

type MetricsTestSuite struct {
	suite.Suite
}

func (suite *MetricsTestSuite) SetupTest() {
	// Reset metrics registry before each test
	metrics.Registry = prometheus.NewRegistry()
}

func (suite *MetricsTestSuite) TestK8sOvercommitOperatorPodsRequestedTotal() {
	K8sOvercommitOperatorPodsRequestedTotal.WithLabelValues("test").Inc()
	count := testutil.ToFloat64(K8sOvercommitOperatorPodsRequestedTotal.WithLabelValues("test"))
	assert.Equal(suite.T(), 1.0, count)
}

func (suite *MetricsTestSuite) TestK8sOvercommitOperatorMutatedPodsTotal() {
	K8sOvercommitOperatorMutatedPodsTotal.WithLabelValues("test").Inc()
	count := testutil.ToFloat64(K8sOvercommitOperatorMutatedPodsTotal.WithLabelValues("test"))
	assert.Equal(suite.T(), 1.0, count)
}

func (suite *MetricsTestSuite) TestK8sOvercommitOperatorPodsNotMutatedTotal() {
	K8sOvercommitOperatorPodsNotMutatedTotal.WithLabelValues("test-generateName", "namespace", "test-type", "test-reason").Inc()
	count := testutil.ToFloat64(K8sOvercommitOperatorPodsNotMutatedTotal.WithLabelValues("test-generateName", "namespace", "test-type", "test-reason"))
	assert.Equal(suite.T(), 1.0, count)
}

func TestMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(MetricsTestSuite))
}
