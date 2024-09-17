/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	certsv1 "k8c.io/kubermatic-cert-manager-operator/api/v1"
)

// CertificateReconciler reconciles a Certificate object
type CertificateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=certs.k8c.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=certs.k8c.io,resources=certificates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=certs.k8c.io,resources=certificates/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

func (r *CertificateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := logf.FromContext(ctx, "Request.Name", req.Name)
	reqLogger.Info("Reconciling Certificate")

	//fetch the certificate CR
	config := &certsv1.Certificate{}
	err := r.Get(ctx, req.NamespacedName, config)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// request object not found, could have been deleted after reconcile request
			return reconcile.Result{}, nil
		}
		// error reading the object
		reqLogger.Error(err, "Failed to get Certificate")
		return reconcile.Result{}, err
	}

	//handling finalizer
	if finished, err := r.HandleFinalizer(ctx, config); err != nil {
		reqLogger.Error(err, "Failed to handle finalizer")
		return reconcile.Result{}, err
	} else if finished {
		reqLogger.Info("Finalizer handling complete")
		return reconcile.Result{}, nil
	}

	secret := &corev1.Secret{}
	err = r.certificateManager(ctx, config, secret)
	if err != nil {
		reqLogger.Error(err, "Failed to reconcile Certificate")
		r.setConditionStatus(config, metav1.ConditionTrue, TypeReconcileError, ReconcileFailed, err.Error())
		updatedErr := r.Status().Update(ctx, config)
		if updatedErr != nil {
			reqLogger.Error(updatedErr, "Failed to update status")
		}
		return reconcile.Result{}, err
	}

	err = r.setCertificateStatus(ctx, config, secret)
	if err != nil {
		reqLogger.Error(err, "Failed to update certificate status")

	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CertificateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&certsv1.Certificate{}).
		Complete(r)
}

func (r *CertificateReconciler) HandleFinalizer(ctx context.Context, config *certsv1.Certificate) (finished bool, err error) {
	reqLogger := logf.FromContext(ctx)
	reqLogger.Info("Handling finalizer for certificate")
	//examine DeletionTimestamp to determine is object is under deletion
	if config.ObjectMeta.DeletionTimestamp.IsZero() {
		//object is not deleted if it doesn't have finalizer,
		//then lets add the finalizer and update
		if !controllerutil.ContainsFinalizer(config, finalizerName) {
			reqLogger.Info("Adding finalizer for certificate")
			controllerutil.AddFinalizer(config, finalizerName)
			if err := r.Update(ctx, config); err != nil {
				reqLogger.Error(err, "Failed to update certificate with finalizer")
				return false, err
			}
		}

	} else {
		reqLogger.Info("Certificate being deleted")
		if controllerutil.ContainsFinalizer(config, finalizerName) {
			reqLogger.Info("Removing finalizer for certificate")
			controllerutil.RemoveFinalizer(config, finalizerName)
			if err := r.Update(ctx, config); err != nil {
				reqLogger.Error(err, "Failed to update certificate by removing finalizer")
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}
