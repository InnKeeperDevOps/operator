package job

import (
	"context"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (j *Job) StartJob(ctx context.Context) {

	securityContext := corev1.PodSecurityContext{
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeUnconfined,
		},
	}
	true := true

	j.batchJob = &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      j.jobName,
			Namespace: "git-builder",
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &j.retries,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      j.jobName,
					Namespace: "git-builder",
					Labels: map[string]string{
						"gitCheck": j.jobName,
					},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &securityContext,
					Volumes:         j.volumes,
					Containers: []corev1.Container{{
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: &true,
							Privileged:               &true,
						},
						Name:            j.jobName,
						Image:           j.image,
						ImagePullPolicy: corev1.PullAlways,
						VolumeMounts:    j.volumeMounts,
						Env:             j.env,
					}},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
	j.client.Create(ctx, j.batchJob)
}
