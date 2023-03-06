package buildstage

import (
	"context"
	"github.com/InnKeeperDevOps/operator/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

func (b *BuildStage) Deployer(ctx context.Context) (ctrl.Result, error) {
	name := b.Deploy.Namespace + "_" + b.Deploy.Name
	log.Debug(name + ": In Deployer Stage")
	namespace := "default"
	if b.Deploy.Spec.Deploy != nil {
		if b.Deploy.Spec.Deploy.Namespace != "" {
			namespace = b.Deploy.Spec.Deploy.Namespace
		}
		log.Debug(name + ": Deploying to namespace: " + namespace)
		err := b.Create(ctx, &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		})
		if err != nil {
			log.Debug(name + ": " + err.Error())
		}
		if b.Deploy.Spec.Publish.Secret != "" {
			err = b.PullSecret(ctx, namespace)
			if err != nil {
				log.Error(name + ": " + err.Error())
				return ctrl.Result{RequeueAfter: time.Second * 20}, err
			}
		}

		switch {
		case b.Deploy.Spec.Deploy.Deployment != nil:
			log.Debug(name + ": Handling " + b.Deploy.Name + " as deployment")
			err = b.HandleDeployment(ctx)
			break
		case b.Deploy.Spec.Deploy.CronJob != nil:
			log.Debug(name + ": Handling " + b.Deploy.Name + " as cronjob")
			err = b.HandleCronJob(ctx)
			break
		case b.Deploy.Spec.Deploy.DaemonSet != nil:
			log.Debug(name + ": Handling " + b.Deploy.Name + " as daemonset")
			err = b.HandleDaemonSet(ctx)
			break
		default:
			log.Debug(name + ": Handling " + b.Deploy.Name + " as simple deployment")
			err = b.HandleSimpleDeployment(ctx)
			break
		}

		if err != nil {
			log.Error(name + ": " + err.Error())
			return ctrl.Result{RequeueAfter: time.Second * 20}, nil
		} else {
			b.Deploy.Status.Deployed = &v1alpha1.Deployed{
				Pod:      "",
				Git:      b.Deploy.Status.Built.Git,
				Complete: true,
				Registry: b.Deploy.Status.Built.Registry,
			}
			err = b.Status().Update(ctx, b.Deploy)
			if err != nil {
				log.Error(name + ": " + err.Error())
			}
		}
	}

	return ctrl.Result{RequeueAfter: time.Second * 20}, nil
}
