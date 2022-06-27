package controller

import (
	"context"
	"time"

	"github.com/acorn-io/baaah"
	"github.com/acorn-io/baaah/pkg/merr"
	"github.com/acorn-io/baaah/pkg/restconfig"
	"github.com/sirupsen/logrus"
	"github.com/traefik/hub-agent-kubernetes/pkg/crd/api/hub/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Start(ctx context.Context) error {
	scheme, err := newScheme()
	if err != nil {
		return err
	}

	if err := waitForCRD(ctx, scheme); err != nil {
		return err
	}

	router, err := baaah.DefaultRouter(scheme)
	if err != nil {
		return err
	}

	sel := klabels.SelectorFromSet(map[string]string{
		AcornManaged: "true",
	})

	router.Type(&networkingv1.Ingress{}).Selector(sel).HandlerFunc(CreateEdgeService)
	router.Type(&v1alpha1.EdgeIngress{}).Selector(sel).HandlerFunc(UpdateIngressWithDomain)

	return router.Start(ctx)
}

func newScheme() (*runtime.Scheme, error) {
	var (
		errs   []error
		scheme = runtime.NewScheme()
	)

	errs = append(errs, networkingv1.AddToScheme(scheme))
	errs = append(errs, v1alpha1.AddToScheme(scheme))
	errs = append(errs, v1.AddToScheme(scheme))
	return scheme, merr.NewErrors(errs...)
}

func waitForCRD(ctx context.Context, scheme *runtime.Scheme) error {
	cfg, err := restconfig.New(scheme)
	if err != nil {
		return err
	}
	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return err
	}

	for {
		crd := &v1.CustomResourceDefinition{}
		err := c.Get(ctx, client.ObjectKey{Name: "edgeingresses.hub.traefik.io"}, crd)
		if apierrors.IsNotFound(err) {
			logrus.Info("waiting for traefik hub agent to be installed")
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
				continue
			}
		} else if err != nil {
			return err
		}
		break
	}

	return nil
}
