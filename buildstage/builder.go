package buildstage

import (
	"context"
	"github.com/InnKeeperDevOps/operator/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"time"
)

func (b *BuildStage) Builder(ctx context.Context) (ctrl.Result, error) {
	name := b.Deploy.Namespace + "_" + b.Deploy.Name
	jobExists := &batchv1.Job{}
	err := b.Client.Get(ctx, types.NamespacedName{Name: b.Deploy.GetBuilderName(), Namespace: "git-builder"}, jobExists)
	if err != nil {
		log.Debug(name + ": " + err.Error())
		err = b.Client.Create(ctx, b.createBuildJob(b.Deploy))

		if err != nil {
			log.Debug(name + ": Error Builder: " + err.Error())
		}
		if b.Deploy.Status.Built != nil {
			b.Deploy.Status.Built.Complete = false
			b.Status().Update(ctx, b.Deploy)
		}
		return ctrl.Result{RequeueAfter: time.Second * 25}, nil
	} else {
		if jobExists.Status.Succeeded == 0 {
			log.Debug(name + ": Waiting for image to build. [" + b.Deploy.Name + "]")
			return ctrl.Result{RequeueAfter: time.Second * 25}, nil
		}

		// continue to next stage.
		log.Debug(name + ": Builder completed making image")
		podList := &v1.PodList{}
		selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
			MatchLabels: map[string]string{"buildFor": b.Deploy.GetBuilderName()},
		})
		err = b.Client.List(ctx, podList, client.MatchingLabelsSelector{
			Selector: selector,
		})
		if err == nil {
			if len(podList.Items) > 0 {
				pod := podList.Items[0]
				logStr := b.getPodLogs(ctx, pod)
				branch := match(logStr, "branch:([a-zA-Z0-9/]+)")
				hash := match(logStr, "hash:(.*?),")
				author := match(logStr, "author:(.*?),")
				date, _ := strconv.Atoi(match(logStr, "date:([0-9]+)"))
				b.Deploy.Status.Built = &v1alpha1.Built{
					Registry: v1alpha1.RegistryStatus{
						Tag:  b.Deploy.Spec.Publish.Tag,
						Host: b.Deploy.Spec.Publish.Host,
					},
					Git: v1alpha1.GitStatus{
						Commit: hash,
						Branch: branch,
						Date:   date,
						Author: author,
					},
					Complete: true,
				}
				policy := metav1.DeletePropagationForeground
				err = b.Delete(ctx, jobExists, &client.DeleteOptions{PropagationPolicy: &policy})
				if err != nil {
					log.Error(name + ": " + err.Error())
				}
				err := b.Status().Update(ctx, b.Deploy)
				if err != nil {
					log.Error(name + ": " + err.Error())
					return ctrl.Result{}, err
				}
				return ctrl.Result{RequeueAfter: time.Second * 5}, nil
			}
		} else {
			log.Error(name + ": " + err.Error())
		}

	}

	return ctrl.Result{RequeueAfter: time.Second * 20}, nil
}
