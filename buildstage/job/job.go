package job

import (
	uuid "github.com/uuid6/uuid6go-proto"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Job struct {
	env          []corev1.EnvVar
	jobName      string
	name         string
	image        string
	volumes      []corev1.Volume
	volumeMounts []corev1.VolumeMount
	retries      int32
	batchJob     *batchv1.Job
	client       client.Client
}

type JobState struct {
	status *batchv1.JobStatus
	job    *Job
}

type JobPod struct {
	pod *corev1.Pod
	job *Job
}

func NewJob(client client.Client, name string, image string, env []corev1.EnvVar, volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, retries int32) *Job {
	var gen uuid.UUIDv7Generator
	return &Job{
		env:          env,
		name:         name,
		jobName:      gen.Next().ToString(),
		image:        image,
		retries:      retries,
		volumes:      volumes,
		volumeMounts: volumeMounts,
		client:       client,
	}
}
