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

func (r *ReconcileExternalService) reconcileEndpoints(instance *esov1alpha1.ExternalService, reqLogger logr.Logger) (reconcile.Result, error) {

	endpoint := CreateEndpointsCr(instance)
	if err := controllerutil.SetControllerReference(instance, endpoint, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	found := &corev1.Endpoints{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Endpoint", "Pod.Namespace", instance.Namespace, "Pod.Name", instance.Name)
		err = r.client.Create(context.TODO(), endpoint)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Service created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	reconciledEndpoint, changed := mergeEndpointWithExternalServiceDef(instance, found)
	if !changed {
		reqLogger.Info("Skip reconcile: Endpoint already exists", "Endpoint.Namespace", found.Namespace, "Endpoint.Name", found.Name)
		return reconcile.Result{}, nil
	}

	reqLogger.Info("Specs changed. Trying to update Endpoint", "Endpoint.Namespace", found.Namespace, "Endpoint.Name", found.Name)

	reconciledEndpoint.ObjectMeta.ResourceVersion = found.ObjectMeta.ResourceVersion
	err = r.client.Update(context.TODO(), reconciledEndpoint)

	if err == nil {
		reqLogger.Info("Updated Endpoint", "Endpoint.Namespace", found.Namespace, "Endpoint.Name", found.Name)
	}

	return reconcile.Result{
		RequeueAfter: 2 * time.Second,
		Requeue:      true,
	}, err
}

func filterRemovedIps(externalService *esov1alpha1.ExternalService, addresses []corev1.EndpointAddress) []corev1.EndpointAddress {
	filteredList := []corev1.EndpointAddress{}
	for _, address := range addresses {
		for _, ip := range externalService.Spec.Ips {
			if address.IP == ip {
				filteredList = append(filteredList, address)
			}
		}
	}

	return filteredList
}

func addMissingIps(readyAddresses []corev1.EndpointAddress, notReadyAddresses []corev1.EndpointAddress, ips []string) []corev1.EndpointAddress {
	newNotReadyAddresses := notReadyAddresses

	for _, ip := range ips {
		found := false
		for _, address := range readyAddresses {
			if address.IP == ip {
				found = true
				break
			}
		}

		if !found {
			for _, address := range notReadyAddresses {
				if address.IP == ip {
					found = true
					break
				}
			}
		}

		// If Ips are not known to the Endpoints yet, add them to NotReady to prevent unready
		// Services getting traffic
		if !found {
			newNotReadyAddresses = append(newNotReadyAddresses, corev1.EndpointAddress{IP: ip})
		}
	}

	return newNotReadyAddresses
}

func mergeEndpointWithExternalServiceDef(externalService *esov1alpha1.ExternalService, endpoint *corev1.Endpoints) (mergedEndpoint *corev1.Endpoints, changed bool) {
	mergedEndpoint = endpoint.DeepCopy()

	newReadyAddresses := filterRemovedIps(externalService, endpoint.Subsets[0].Addresses)
	newNotReadyAddresses := filterRemovedIps(externalService, endpoint.Subsets[0].NotReadyAddresses)

	// Then add new IPs
	mergedEndpoint.Subsets[0].Addresses = newReadyAddresses
	mergedEndpoint.Subsets[0].NotReadyAddresses = addMissingIps(newReadyAddresses, newNotReadyAddresses, externalService.Spec.Ips)
	mergedEndpoint.Subsets[0].Ports[0].Port = externalService.Spec.Port

	return mergedEndpoint, !equalIgnoreReady(mergedEndpoint, endpoint)
}

func equalIgnoreReady(a *corev1.Endpoints, b *corev1.Endpoints) bool {
	if !reflect.DeepEqual(a.ObjectMeta, b.ObjectMeta) {
		return false
	}

	if a.Subsets[0].Ports[0] != b.Subsets[0].Ports[0] {
		return false
	}

	// This check ignores the difference between Ready and not Ready
	aAddresses := a.Subsets[0].Addresses
	aAddresses = append(aAddresses, a.Subsets[0].NotReadyAddresses...)

	bAddresses := b.Subsets[0].Addresses
	bAddresses = append(bAddresses, b.Subsets[0].NotReadyAddresses...)

	//check if b has every ip a has
	for _, aAddress := range aAddresses {
		found := false
		for _, bAddress := range bAddresses {
			if aAddress == bAddress {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	// check if a has every ip b has
	for _, bAddress := range bAddresses {
		found := false
		for _, aAddress := range aAddresses {
			if aAddress == bAddress {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func CreateEndpointsCr(i *esov1alpha1.ExternalService) *corev1.Endpoints {

	labels := map[string]string{
		"app":         i.Name,
		"serviceType": "external",
	}

	endpointAddresses := []corev1.EndpointAddress{}
	for _, ip := range i.Spec.Ips {
		endpointAddresses = append(endpointAddresses, corev1.EndpointAddress{
			IP: ip,
		})
	}

	return &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      i.Name,
			Namespace: i.Namespace,
			Labels:    labels,
		},
		Subsets: []corev1.EndpointSubset{
			corev1.EndpointSubset{
				NotReadyAddresses: endpointAddresses,
				Ports: []corev1.EndpointPort{
					corev1.EndpointPort{
						Port: i.Spec.Port,
					},
				},
			},
		},
	}

}
