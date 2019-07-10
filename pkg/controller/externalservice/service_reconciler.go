package externalservice

import (
	"context"
	"reflect"
	"time"

	esov1alpha1 "github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileExternalService) reconcileService(instance *esov1alpha1.ExternalService, reqLogger logr.Logger) (reconcile.Result, error) {

	service := createServiceCr(instance)
	if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "namespace", instance.Namespace, "service", instance.Name)
		err = r.client.Create(context.TODO(), service)
		if err != nil {
			return reconcile.Result{}, err
		}

		//Service  created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	newService := createServiceCr(instance)
	if !reflect.DeepEqual(*newService, *found) {
		reqLogger.Info("Specs Changed for Service. Trying to update", "namespace", found.Namespace, "service", found.Name)
		newService.ObjectMeta.ResourceVersion = found.ObjectMeta.ResourceVersion

		if err := controllerutil.SetControllerReference(instance, newService, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		if err = r.client.Update(context.TODO(), newService); err == nil {
			reqLogger.Info("Updated Service", "namespace", found.Namespace, "service", found.Name)
		}

		return reconcile.Result{
			RequeueAfter: time.Second,
			Requeue:      true,
		}, err
	}
	reqLogger.Info("Skip reconcile: Service already exists", "namespace", found.Namespace, "service", found.Name)

	return reconcile.Result{}, nil
}

func createServiceCr(i *esov1alpha1.ExternalService) *corev1.Service {
	labels := map[string]string{
		"app":         i.Name,
		"serviceType": "external",
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      i.Name,
			Namespace: i.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Port: i.Spec.Port,
				},
			},
			ClusterIP: "None",
			Type:      corev1.ServiceTypeClusterIP,
		},
	}
}
