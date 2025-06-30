// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"time"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"

	resources "github.com/InditexTech/k8s-overcommit-operator/internal/resources"
	"github.com/InditexTech/k8s-overcommit-operator/internal/utils"
	certmanager "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// OvercommitClassReconciler reconciles a OvercommitClass object
type OvercommitClassReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=overcommit.inditex.dev,resources=overcommitclasses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=overcommit.inditex.dev,resources=overcommitclasses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=overcommit.inditex.dev,resources=overcommitclasses/finalizers,verbs=update
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OvercommitClass object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile

// SetupWithManager sets up the controller with the Manager.
func (r *OvercommitClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&overcommit.OvercommitClass{}).
		Named("OvercommitClass").
		Complete(r)
}

func (r *OvercommitClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	label, err := utils.GetOvercommitLabel(ctx, r.Client)
	if err != nil {
		logger.Error(err, "Failed to get Overcommit label")
		return ctrl.Result{}, err
	}

	overcommitClass := &overcommit.OvercommitClass{}

	err = r.Get(ctx, req.NamespacedName, overcommitClass)
	if err != nil {
		logger.Info("Deleting resources for the class", "name", req.Name)
		err := ensureResourceDeleted(ctx, r.Client, &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: req.Name + "-webhook-deployment", Namespace: "k8s-overcommit"}})
		if err != nil {
			logger.Error(err, "Failed to delete Deployment")
		}
		err = ensureResourceDeleted(ctx, r.Client, &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: req.Name + "-webhook-service", Namespace: "k8s-overcommit"}})
		if err != nil {
			logger.Error(err, "Failed to delete Service")
		}
		err = ensureResourceDeleted(ctx, r.Client, &certmanager.Certificate{ObjectMeta: metav1.ObjectMeta{Name: req.Name + "-webhook-certificate", Namespace: "k8s-overcommit"}})
		if err != nil {
			logger.Error(err, "Failed to delete Certificate")
		}
		err = ensureResourceDeleted(ctx, r.Client, &admissionv1.MutatingWebhookConfiguration{ObjectMeta: metav1.ObjectMeta{Name: req.Name + "overcommit-webhook", Namespace: "k8s-overcommit"}})
		if err != nil {
			logger.Error(err, "Failed to delete MutatingWebhookConfiguration")
		}
		if getTotalClasses(ctx, r.Client) != nil {
			logger.Error(err, "Failed to update metrics")
		}
		return ctrl.Result{}, err
	}
	// Check if the OvercommitClass has the correct owner reference
	overcommitResource, err := utils.GetOvercommit(ctx, r.Client)
	if err != nil {
		logger.Error(err, "Failed to get Overcommit")
		return ctrl.Result{}, err
	}

	needsOwnerUpdate := false
	if len(overcommitClass.OwnerReferences) == 0 {
		needsOwnerUpdate = true
	} else {
		// Check if the current owner reference is correct
		hasCorrectOwner := false
		for _, ownerRef := range overcommitClass.OwnerReferences {
			if ownerRef.UID == overcommitResource.UID && ownerRef.Kind == "Overcommit" {
				hasCorrectOwner = true
				break
			}
		}
		if !hasCorrectOwner {
			needsOwnerUpdate = true
		}
	}

	if needsOwnerUpdate {
		logger.Info("Setting ControllerReference for OvercommitClass", "name", overcommitClass.Name)
		err = controllerutil.SetControllerReference(&overcommitResource, overcommitClass, r.Scheme)
		if err != nil {
			logger.Error(err, "Failed to set ControllerReference for OvercommitClass")
			return ctrl.Result{}, err
		}

		// Update the OvercommitClass with the new owner reference
		err = r.Update(ctx, overcommitClass)
		if err != nil {
			logger.Error(err, "Failed to update OvercommitClass with ControllerReference")
			return ctrl.Result{}, err
		}
		logger.Info("ControllerReference updated, requeuing reconciliation")
		return ctrl.Result{}, nil
	}

	logger.Info("Reconciling resources for the class", "name", overcommitClass)
	deployment := resources.CreateDeployment(*overcommitClass)
	service := resources.CreateService(overcommitClass.Name)
	certificate := resources.CreateCertificate(overcommitClass.Name, *service)
	webhookConfig := resources.CreateMutatingWebhookConfiguration(*overcommitClass, *service, *certificate, label)

	// Use CreateOrUpdate for each resource
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		// Update the deployment spec if needed
		updatedDeployment := resources.CreateDeployment(*overcommitClass)
		deployment.Spec = updatedDeployment.Spec
		return controllerutil.SetControllerReference(overcommitClass, deployment, r.Scheme)
	})
	if err != nil {
		logger.Error(err, "Failed to create or update Deployment")
		return ctrl.Result{}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		// Update the service spec if needed
		updatedService := resources.CreateService(overcommitClass.Name)
		service.Spec = updatedService.Spec
		return controllerutil.SetControllerReference(overcommitClass, service, r.Scheme)
	})
	if err != nil {
		logger.Error(err, "Failed to create or update Service")
		return ctrl.Result{}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, certificate, func() error {
		// Update the certificate spec if needed
		updatedCertificate := resources.CreateCertificate(overcommitClass.Name, *service)
		certificate.Spec = updatedCertificate.Spec
		return controllerutil.SetControllerReference(overcommitClass, certificate, r.Scheme)
	})
	if err != nil {
		logger.Error(err, "Failed to create or update Certificate")
		return ctrl.Result{}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, webhookConfig, func() error {
		// Update the webhook configuration spec if needed
		updatedWebhookConfig := resources.CreateMutatingWebhookConfiguration(*overcommitClass, *service, *certificate, label)
		webhookConfig.Webhooks = updatedWebhookConfig.Webhooks
		return controllerutil.SetControllerReference(overcommitClass, webhookConfig, r.Scheme)
	})
	if err != nil {
		logger.Error(err, "Failed to create or update MutatingWebhookConfiguration")
		return ctrl.Result{}, err
	}

	if getTotalClasses(ctx, r.Client) != nil {
		logger.Error(err, "Failed to update metrics")
		return ctrl.Result{}, err
	}

	// Update the status of the resources
	if err := r.updateResourcesStatus(ctx, overcommitClass); err != nil {
		logger.Error(err, "Error updating resource status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		RequeueAfter: 10 * time.Second,
	}, nil
}
