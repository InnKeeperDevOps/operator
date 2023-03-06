package job

import (
	"context"
	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (j *Job) Status(ctx context.Context) batchv1.JobStatus {
	batchJob := batchv1.Job{}

	j.client.Get(ctx, client.ObjectKey{
		Name:      j.jobName,
		Namespace: "git-builder",
	}, &batchJob)

	return batchJob.Status
}

func (j *Job) Monitor(ctx context.Context) *JobState {
	batchJobStatus := j.Status(ctx)

	return &JobState{
		status: &batchJobStatus,
		job:    j,
	}
}

func (state *JobState) Succeeded() bool {
	return state.status.Succeeded > 0
}

func (state *JobState) Failed() bool {
	return state.status.Failed > 0
}

func (state *JobState) Started() bool {
	return state.status.Active > 0
}
func (state *JobState) Finished() bool {
	return state.Failed() || state.Succeeded()
}
