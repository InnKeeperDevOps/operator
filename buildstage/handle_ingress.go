package buildstage

import (
	"context"
	"errors"
	"github.com/InnKeeperDevOps/operator/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

func (b *BuildStage) DeleteIngress(ctx context.Context, ingress *v1alpha1.Ingress) error {
	name := b.Deploy.Namespace + "_" + b.Deploy.Name
	if ingress == nil {
		return errors.New(name + ": Ingress not set?")
	}
	log.Debug(name + ": Deleting service " + ingress.Name)
	policy := metav1.DeletePropagationForeground
	err := b.Delete(ctx, &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingress.Name,
			Namespace: b.Deploy.Namespace,
		},
	}, &client.DeleteOptions{PropagationPolicy: &policy})
	if err != nil {
		log.Error(name + ": " + err.Error())
		return err
	}
	log.Debug(name + ": Deleting http route " + ingress.Name)
	err = b.Delete(ctx, &v1beta1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingress.Name,
			Namespace: b.Deploy.Namespace,
		},
	}, &client.DeleteOptions{PropagationPolicy: &policy})
	if err != nil {
		log.Error(name + ": " + err.Error())
		return err
	}
	return nil
}

func (b *BuildStage) HandleIngress(ctx context.Context, ingress *v1alpha1.Ingress) error {
	name := b.Deploy.Namespace + "_" + b.Deploy.Name
	if ingress == nil {
		return errors.New("Ingress not set?")
	}
	log.Debug(name + ": Saving changes for ingress " + ingress.Name)
	headers := []v1beta1.HTTPHeaderMatch{}
	for key, val := range ingress.Headers {
		headers = append(headers, v1beta1.HTTPHeaderMatch{
			Name:  v1beta1.HTTPHeaderName(key),
			Value: val,
		})
	}
	namespace := v1beta1.Namespace(b.Deploy.Namespace)

	gateways := []v1beta1.ParentReference{}
	for _, gateway := range ingress.Gateway {
		gateways = append(gateways, v1beta1.ParentReference{
			Namespace: &namespace,
			Name:      v1beta1.ObjectName(gateway),
		})
	}
	service := v1.Service{}
	err := b.Get(ctx, client.ObjectKey{
		Name:      ingress.Name,
		Namespace: b.Deploy.Namespace,
	}, &service)
	if err != nil {
		log.Error(name + ": " + ingress.Name + ": " + err.Error())
		log.Debug(name + ": Creating new service " + ingress.Name)
		service = v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ingress.Name,
				Namespace: b.Deploy.Namespace,
			},
			Spec: v1.ServiceSpec{
				Type: v1.ServiceTypeClusterIP,
				Ports: []v1.ServicePort{
					v1.ServicePort{
						Name: ingress.Name,
						Port: int32(ingress.Port),
					},
				},
				Selector: map[string]string{
					"app-connector": b.Deploy.Name + "_" + b.Deploy.Namespace,
				},
			},
		}
		err = b.Create(ctx, &service)
		if err != nil {
			log.Error(name + ": " + ingress.Name + ": " + err.Error())
			return err
		}
	} else {
		log.Debug(name + ": Updating existing service " + ingress.Name)
		service.Spec = v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: ingress.Name,
					Port: int32(ingress.Port),
				},
			},
			Selector: map[string]string{
				"app-connector": b.Deploy.Name + "_" + b.Deploy.Namespace,
			},
		}
		err = b.Update(ctx, &service)
		if err != nil {
			log.Error(name + ": " + ingress.Name + ": " + err.Error())
			return err
		}
	}

	port := v1beta1.PortNumber(ingress.Port)

	prefix := v1beta1.PathMatchPathPrefix

	route := v1beta1.HTTPRoute{}
	err = b.Client.Get(ctx, client.ObjectKey{
		Namespace: b.Deploy.Spec.Deploy.Namespace,
		Name:      ingress.Name,
	}, &route)
	if err != nil {
		log.Debug(name + ": Creating new ingress " + ingress.Name)
		log.Error(name + ": " + ingress.Name + ": " + err.Error())
		route = v1beta1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ingress.Name,
				Namespace: b.Deploy.Namespace,
			},
			Spec: v1beta1.HTTPRouteSpec{
				CommonRouteSpec: v1beta1.CommonRouteSpec{
					ParentRefs: gateways,
				},
				Hostnames: []v1beta1.Hostname{
					v1beta1.Hostname(ingress.Domain),
				},
				Rules: []v1beta1.HTTPRouteRule{
					v1beta1.HTTPRouteRule{
						Matches: []v1beta1.HTTPRouteMatch{
							v1beta1.HTTPRouteMatch{
								Path: &v1beta1.HTTPPathMatch{
									Type:  &prefix,
									Value: &ingress.Path,
								},
								Headers: headers,
							},
						},
						BackendRefs: []v1beta1.HTTPBackendRef{
							v1beta1.HTTPBackendRef{
								BackendRef: v1beta1.BackendRef{
									BackendObjectReference: v1beta1.BackendObjectReference{
										Name:      v1beta1.ObjectName(ingress.Name),
										Namespace: &namespace,
										Port:      &port,
									},
								},
							},
						},
					},
				},
			},
		}
		err = b.Create(ctx, &route)
		if err != nil {
			log.Debug(name + ": " + ingress.Name + ": " + err.Error())
			return err
		}
	} else {
		log.Debug(name + ": Updating existing ingress " + ingress.Name)
		route.Spec = v1beta1.HTTPRouteSpec{
			CommonRouteSpec: v1beta1.CommonRouteSpec{
				ParentRefs: gateways,
			},
			Hostnames: []v1beta1.Hostname{
				v1beta1.Hostname(ingress.Domain),
			},
			Rules: []v1beta1.HTTPRouteRule{
				v1beta1.HTTPRouteRule{
					Matches: []v1beta1.HTTPRouteMatch{
						v1beta1.HTTPRouteMatch{
							Path: &v1beta1.HTTPPathMatch{
								Type:  &prefix,
								Value: &ingress.Path,
							},
							Headers: headers,
						},
					},
					BackendRefs: []v1beta1.HTTPBackendRef{
						v1beta1.HTTPBackendRef{
							BackendRef: v1beta1.BackendRef{
								BackendObjectReference: v1beta1.BackendObjectReference{
									Name:      v1beta1.ObjectName(ingress.Name),
									Namespace: &namespace,
									Port:      &port,
								},
							},
						},
					},
				},
			},
		}
		err = b.Update(ctx, &route)
		if err != nil {
			log.Error(name + ": " + ingress.Name + ": " + err.Error())
			return err
		}
	}
	return nil
}
