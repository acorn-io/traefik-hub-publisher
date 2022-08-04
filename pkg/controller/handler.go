package controller

import (
	"strings"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/traefik/hub-agent-kubernetes/pkg/crd/api/hub/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getACP(req router.Request) (*v1alpha1.EdgeIngressACP, error) {
	acp := &v1alpha1.AccessControlPolicy{}
	err := req.Get(acp, "", "acorn")
	if apierrors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &v1alpha1.EdgeIngressACP{
		Name: "acorn",
	}, nil
}

func UpdateIngressWithDomain(req router.Request, resp router.Response) error {
	edgeservice := req.Object.(*v1alpha1.EdgeIngress)
	if edgeservice.Status.URL == "" {
		return nil
	}

	ingress := &networkingv1.Ingress{}
	err := req.Get(ingress, edgeservice.Namespace, strings.TrimSuffix(edgeservice.Name, "-th"))
	if apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	if ingress.Annotations[AcornPublishURL] != edgeservice.Status.URL {
		if ingress.Annotations == nil {
			ingress.Annotations = map[string]string{}
		}
		ingress.Annotations[AcornPublishURL] = edgeservice.Status.URL
		return req.Client.Update(req.Ctx, ingress)
	}

	return nil
}

func CreateEdgeService(req router.Request, resp router.Response) error {
	ingress := req.Object.(*networkingv1.Ingress)

	acp, err := getACP(req)
	if err != nil {
		return err
	}

	for _, rule := range ingress.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service == nil {
				continue
			}
			resp.Objects(&v1alpha1.EdgeIngress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ingress.Name + "-th",
					Namespace: ingress.Namespace,
					Labels: map[string]string{
						AcornManaged: "true",
					},
				},
				Spec: v1alpha1.EdgeIngressSpec{
					ACP: acp,
					Service: v1alpha1.EdgeIngressService{
						Name: path.Backend.Service.Name,
						Port: int(path.Backend.Service.Port.Number),
					},
				},
			})
			return nil
		}
	}

	return nil
}
