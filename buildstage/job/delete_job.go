package job

import (
	"context"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (j *Job) Delete(ctx context.Context) error {
	policy := metav1.DeletePropagationForeground
	log.Debug(j.name + ": Deleting job " + j.jobName)
	return j.client.Delete(ctx, j.batchJob, &client.DeleteOptions{PropagationPolicy: &policy})
}
