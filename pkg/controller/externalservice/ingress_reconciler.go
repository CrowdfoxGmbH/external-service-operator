package externalservice

import (
	"context"
	"reflect"
	"time"

	esov1alpha1 "github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1"
	extv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileExternalService) reconcileIngress(instance *esov1alpha1.ExternalService, reqLogger logr.Logger) (reconcile.Result, error) {

	key := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}

	// Check if Ingress should exist
	if len(instance.Spec.Hosts) <= 0 {
		// Check if Ingress exists
		reqLogger.V(1).Info("No Host definitions found. Skip Creating Ingress")
		found := &extv1.Ingress{}
		err := r.client.Get(context.TODO(), key, found)

		if err == nil {
			err := r.client.Delete(context.TODO(), found)
			reqLogger.V(1).Info("Found existing old Ingress. Deleted Resource")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	ingress := createIngressCr(instance)
	if err := controllerutil.SetControllerReference(instance, ingress, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &extv1.Ingress{}
	err := r.client.Get(context.TODO(), key, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Ingress", "Pod.Namespace", instance.Namespace, "Pod.Name", instance.Name)
		err = r.client.Create(context.TODO(), ingress)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	newIngress := createIngressCr(instance)
	if !reflect.DeepEqual(*newIngress, *found) {
		reqLogger.Info("Specs changed. Trying to update Ingress", "namespace", found.Namespace, "ingress", found.Name)
		newIngress.ObjectMeta.ResourceVersion = found.ObjectMeta.ResourceVersion

		if err := controllerutil.SetControllerReference(instance, newIngress, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		if err = r.client.Update(context.TODO(), newIngress); err == nil {
			reqLogger.Info("Updated Ingress", "namespace", found.Namespace, "ingress", found.Name)
		}

		return reconcile.Result{
			RequeueAfter: time.Second,
			Requeue:      true,
		}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Ingress already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)

	return reconcile.Result{}, nil
}

func createIngressCr(i *esov1alpha1.ExternalService) *extv1.Ingress {
	labels := map[string]string{
		"app":         i.Name,
		"serviceType": "external",
	}

	ingressrules := []extv1.IngressRule{}
	for _, hostpath := range i.Spec.Hosts {
		ingressrules = append(ingressrules, extv1.IngressRule{
			Host: hostpath.Host,
			IngressRuleValue: extv1.IngressRuleValue{
				HTTP: &extv1.HTTPIngressRuleValue{
					Paths: []extv1.HTTPIngressPath{
						extv1.HTTPIngressPath{
							Path: hostpath.Path,
							Backend: extv1.IngressBackend{
								ServiceName: i.Name,
								ServicePort: intstr.FromInt(int(i.Spec.Port)),
							},
						},
					},
				},
			},
		})
	}

	return &extv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        i.Name,
			Namespace:   i.Namespace,
			Labels:      labels,
			Annotations: i.Annotations,
		},
		Spec: extv1.IngressSpec{
			Rules: ingressrules,
		},
	}
}
