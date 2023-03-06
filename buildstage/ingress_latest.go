package buildstage

import (
	"context"
	"github.com/InnKeeperDevOps/operator/api/v1alpha1"
	"github.com/InnKeeperDevOps/operator/buildstage/changelog"
	log "github.com/sirupsen/logrus"
	"reflect"
)

func (b *BuildStage) LatestIngress(ctx context.Context) bool {
	name := b.Deploy.Namespace + "_" + b.Deploy.Name
	b.Changes = changelog.New()
	b.Deletes = changelog.New()
	entriesLeft := []string{}
	for _, statusIngressEntry := range b.Deploy.Spec.Ingress {
		entriesLeft = append(entriesLeft, statusIngressEntry.Name)
	}
	if b.Deploy.Status.Ingress == nil {
		for _, ingressEntry := range b.Deploy.Spec.Ingress {
			log.Debug(name + ": INGRESS NOT CREATED")
			b.Changes.Add(ingressEntry.Name, "*")
		}
	} else {
		ingressEntries := map[string]*v1alpha1.Ingress{}
		for _, ingressEntry := range b.Deploy.Spec.Ingress {
			ingressEntries[ingressEntry.Name] = ingressEntry
		}
		for _, statusIngressEntry := range b.Deploy.Status.Ingress {
			entriesLeft = remove(entriesLeft, statusIngressEntry.Name)
			if ingressEntries[statusIngressEntry.Name] == nil {
				log.Debug(name + ": INGRESS ENTRY MISSING DIFF, NEED TO DELETE")
				b.Deletes.Add(statusIngressEntry.Name, "delete")
			} else {
				base := ingressEntries[statusIngressEntry.Name]
				if statusIngressEntry.Domain != base.Domain {
					log.Debug(name + ": INGRESS DOMAIN DIFF")
					b.Changes.Add(statusIngressEntry.Name, "domain")
				}
				if !reflect.DeepEqual(statusIngressEntry.Gateway, base.Gateway) {
					log.Debug(name + ": INGRESS GATEWAY DIFF")
					b.Changes.Add(statusIngressEntry.Name, "gateway")
				}
				if statusIngressEntry.Path != base.Path {
					log.Debug(name + ": INGRESS PATH DIFF")
					b.Changes.Add(statusIngressEntry.Name, "path")
				}
				if statusIngressEntry.Port != base.Port {
					log.Debug(name + ": INGRESS PORT DIFF")
					b.Changes.Add(statusIngressEntry.Name, "port")
				}
				if statusIngressEntry.Headers != nil {
					for headerKey, headerVal := range base.Headers {
						if headerVal != base.Headers[headerKey] {
							log.Debug(name + ": INGRESS HEADER DIFF")
							b.Changes.Add(statusIngressEntry.Name, "header["+headerKey+"]")
						}
					}
				}
			}
		}
	}
	for _, entry := range entriesLeft {
		_, ok := b.Changes.Get()[entry]
		if !ok {
			log.Debug(name + ": INGRESS ENTRY NOT CREATED")
			b.Changes.Add(entry, "*")
		}
	}
	return b.Changes.Count() == 0
}
