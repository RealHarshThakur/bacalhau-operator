package controllers

import (
	jobv1 "bacalhau/api/v1"
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *JobReconciler) reconcileJobCM(ctx context.Context, job *jobv1.Job, bjobID string) (*string, error) {
	l := log.FromContext(ctx)
	cm := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      job.Name,
		Namespace: job.Namespace,
	}, cm)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			l.Error(err, "unable to fetch ConfigMap")
			return nil, err
		}
		cm = r.generateConfigmap(*job, bjobID)
		err := controllerutil.SetControllerReference(job, cm, r.Scheme)
		if err != nil {
			l.Error(err, "could not set owner reference")
			return nil, err
		}

		err = r.Create(ctx, cm)
		if err != nil {
			l.Error(err, "unable to create ConfigMap")
			return nil, err
		}

	}
	id := cm.Data[jobID]
	return &id, nil
}

func (r *JobReconciler) generateConfigmap(job jobv1.Job, bjobID string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.Name,
			Namespace: job.Namespace,
		},
		Immutable: pointerTo(true),
		Data: map[string]string{
			jobID: bjobID,
		},
	}
}
