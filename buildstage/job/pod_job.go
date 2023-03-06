package job

import (
	"context"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (j *Job) Pod(ctx context.Context) *JobPod {
	log.Debug(j.name + ": " + j.jobName + ", retrieving job pod")
	podList := &v1.PodList{}
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{"gitCheck": j.jobName},
	})
	err = j.client.List(ctx, podList, client.MatchingLabelsSelector{
		Selector: selector,
	})
	if err == nil {
		if len(podList.Items) > 0 {
			pod := podList.Items[0]
			return &JobPod{
				pod: &pod,
				job: j,
			}
		} else {
			log.Debug(j.name + ": No pods found")
		}
	}
	return nil
}
