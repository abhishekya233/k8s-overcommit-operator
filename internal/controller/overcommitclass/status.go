// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"os"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	certmanager "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OvercommitClassReconciler) updateResourcesStatus(ctx context.Context, overcommitClass *overcommit.OvercommitClass) error {
	logger := log.FromContext(ctx)

	// Resources
	readyStatus := make(map[string]overcommit.ResourceStatus)

	// Deployment
	deployName := overcommitClass.Name + "-overcommit-webhook"
	deploy := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: deployName, Namespace: os.Getenv("POD_NAMESPACE")}, deploy)
	if err == nil {
		readyStatus["deployment"] = overcommit.ResourceStatus{Name: overcommitClass.Name + "-webhook-deployment", Ready: true}
	} else {
		readyStatus["deployment"] = overcommit.ResourceStatus{Name: overcommitClass.Name + "-webhook-deployment", Ready: false}
	}

	// Service
	svcName := overcommitClass.Name + "-webhook-service"
	svc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: svcName, Namespace: os.Getenv("POD_NAMESPACE")}, svc)
	if err == nil {
		readyStatus["service"] = overcommit.ResourceStatus{Name: svcName, Ready: true}
	} else {
		readyStatus["service"] = overcommit.ResourceStatus{Name: svcName, Ready: false}
	}

	// Certificate
	certName := overcommitClass.Name + "-webhook-certificate"
	cert := &certmanager.Certificate{}
	err = r.Get(ctx, types.NamespacedName{Name: certName, Namespace: os.Getenv("POD_NAMESPACE")}, cert)
	if err == nil {
		readyStatus["certificate"] = overcommit.ResourceStatus{Name: certName, Ready: true}
	} else {
		readyStatus["certificate"] = overcommit.ResourceStatus{Name: certName, Ready: false}
	}

	// Webhook Configuration
	webhookName := overcommitClass.Name + "-overcommit-webhook"
	webhook := &admissionv1.MutatingWebhookConfiguration{}
	err = r.Get(ctx, client.ObjectKey{Name: webhookName}, webhook)
	if err == nil {
		readyStatus["webhook"] = overcommit.ResourceStatus{Name: webhookName, Ready: true}
	} else {
		readyStatus["webhook"] = overcommit.ResourceStatus{Name: webhookName, Ready: false}
	}

	// Convert map values to a slice
	resources := make([]overcommit.ResourceStatus, 0, len(readyStatus)) // Pre-allocate slice
	for _, status := range readyStatus {
		resources = append(resources, status)
	}
	overcommitClass.Status.Resources = resources
	if updateErr := r.Status().Update(ctx, overcommitClass); updateErr != nil {
		logger.Error(updateErr, "Failed to update OvercommitClass status")
		// For resource conflicts, don't fail but log and continue
		if !apierrors.IsConflict(updateErr) {
			return updateErr
		}
		logger.Info("Resource conflict detected during status update, continuing with condition update")
	}

	// Conditions
	allReady := true
	for _, res := range overcommitClass.Status.Resources {
		if !res.Ready {
			allReady = false
			break
		}
	}

	condition := metav1.Condition{
		Type:    "ResourcesReady",
		Status:  metav1.ConditionTrue,
		Reason:  "AllResourcesReady",
		Message: "All managed resources are ready",
	}

	if !allReady {
		condition.Status = metav1.ConditionFalse
		condition.Reason = "ResourcesNotReady"
		condition.Message = "Some resources are not ready"
	}

	// Update or add the condition
	setCondition(&overcommitClass.Status, condition)

	// Update status in the API
	if err := r.Status().Update(ctx, overcommitClass); err != nil {
		logger.Error(err, "Failed to update status with conditions")
		// For resource conflicts, don't fail but log
		if !apierrors.IsConflict(err) {
			return err
		}
		logger.Info("Resource conflict detected during condition update")
	}

	return nil
}
