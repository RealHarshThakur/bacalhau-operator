package controllers

import (
	jobv1 "bacalhau/api/v1"
	"context"
	"fmt"

	bdocker "github.com/bacalhau-project/bacalhau/cmd/cli/docker"
	"github.com/bacalhau-project/bacalhau/pkg/bacerrors"
	jobutils "github.com/bacalhau-project/bacalhau/pkg/job"
	"github.com/bacalhau-project/bacalhau/pkg/model"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func generateBacalhauJob(ctx context.Context, job *jobv1.Job) (*model.Job, error) {
	l := log.FromContext(ctx)

	opts := bdocker.NewDockerRunOptions()
	cmdArgs := make([]string, 0, len(job.Spec.Entrypoint)+1)
	cmdArgs = append(cmdArgs, job.Spec.Image)
	cmdArgs = append(cmdArgs, job.Spec.Entrypoint...)
	bjob, err := bdocker.CreateJob(ctx, cmdArgs, opts)
	if err != nil {
		l.Error(err, "unable to create job")
		return nil, err
	}

	if err := jobutils.VerifyJob(ctx, bjob); err != nil {
		if _, ok := err.(*bacerrors.ImageNotFound); ok {
			return nil, fmt.Errorf("docker image '%s' not found in the registry, or needs authorization", bjob.Spec.Docker.Image)
		}
		return nil, fmt.Errorf("verifying job: %s", err)
	}
	return bjob, nil
}
