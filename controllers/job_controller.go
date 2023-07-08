/*
Copyright 2023.

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

package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	jobv1 "bacalhau/api/v1"

	"github.com/bacalhau-project/bacalhau/cmd/util"
	"github.com/bacalhau-project/bacalhau/pkg/requester/publicapi"
	"github.com/bacalhau-project/bacalhau/pkg/system"
)

const (
	jobID = "jobID"
)

// JobReconciler reconciles a Job object
type JobReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	BCleanupManager *system.CleanupManager
	Bclient         *publicapi.RequesterAPIClient
}

//+kubebuilder:rbac:groups=bacalhau.org.bacalhau.org,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bacalhau.org.bacalhau.org,resources=jobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bacalhau.org.bacalhau.org,resources=jobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *JobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	ctx = context.WithValue(ctx, util.SystemManagerKey, r.BCleanupManager)

	l.Info("Reconciling", "job", req.NamespacedName)
	job := &jobv1.Job{}
	err := r.Get(ctx, req.NamespacedName, job)
	if err != nil {
		l.Error(err, "unable to fetch Job")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	jobCM := &corev1.ConfigMap{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      job.Name,
		Namespace: job.Namespace,
	}, jobCM)
	if err != nil && !k8sErrors.IsNotFound(err) {
		l.Error(err, "unable to fetch ConfigMap")
		return ctrl.Result{}, err
	}

	bjobID := jobCM.Data[jobID]
	if err != nil && k8sErrors.IsNotFound(err) {
		bjob, err := generateBacalhauJob(ctx, job)
		if err != nil {
			l.Error(err, "unable to generate job")
			return ctrl.Result{}, err
		}
		executingJob, err := r.Bclient.Submit(ctx, bjob)
		if err != nil {
			l.Error(err, "unable to execute job")
			return ctrl.Result{}, err
		}

		// We create an immutable Configmap to store the jobID. This is going to be used to be manage the lifecyle of the Bacalhau job
		id, err := r.reconcileJobCM(ctx, job, executingJob.Metadata.ID)
		if err != nil {
			if k8sErrors.IsAlreadyExists(err) {
				// should never occur, but if it does, we requeue
				return ctrl.Result{Requeue: true}, err
			}
			_, err = r.Bclient.Cancel(ctx, executingJob.Metadata.ID, "failed to reconcile job configmap")
			if err != nil {
				// Not sure we can do much here, but we should log it
				l.Error(err, "unable to cancel job")
				return ctrl.Result{}, err
			}
			l.Error(err, "unable to reconcile job configmap")
			return ctrl.Result{}, err
		}

		bjobID = *id
	}

	if job.Status.JobID != bjobID {
		job.Status.JobID = bjobID
		err = r.Status().Update(ctx, job)
		if err != nil {

			l.Error(err, "unable to update job status")
			return ctrl.Result{}, err
		}
	}

	l.Info("Reconciled successfully", "job", req.NamespacedName, "jobID", bjobID)

	return ctrl.Result{}, nil
}

// PointerTo converts passed value to pointer
func pointerTo[T any](value T) *T {
	return &value
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&jobv1.Job{}).
		Complete(r)
}
