package testutils

import (
	esov1alpha1 "github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/intstr"
)

func CreateDefaultTestProbe() corev1.Probe {
	return CreateTestProbe(1, 5, 3, 3, 3, corev1.URISchemeHTTP, 80, "/")
}

func CreateTestProbe(delay int32, timeout int32, period int32, success int32, failure int32, scheme corev1.URIScheme, port int32, path string) corev1.Probe {

	return corev1.Probe{
		InitialDelaySeconds: delay,
		TimeoutSeconds:      timeout,
		PeriodSeconds:       period,
		SuccessThreshold:    success,
		FailureThreshold:    failure,
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   path,
				Scheme: scheme,
				Port:   intstr.FromInt(int(port)),
			},
		},
	}

}

func CreateDefaultExternalService() *esov1alpha1.ExternalService {
	hosts := CreateDefaultExternalHostPaths()

	return CreateExternalService("TestService", "external-services", []string{"10.0.102.10", "10.0.102.12", "10.0.102.14"}, 8080, hosts)
}

func CreateDefaultExternalHostPaths() []esov1alpha1.ExternalServiceHostPath {
	return []esov1alpha1.ExternalServiceHostPath{
		esov1alpha1.ExternalServiceHostPath{
			Host: "sub.example.com",
			Path: "/",
		},
	}
}

func CreateExternalService(name string, namespace string, ips []string, port int, hosts []esov1alpha1.ExternalServiceHostPath) *esov1alpha1.ExternalService {

	return &esov1alpha1.ExternalService{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{},
		},

		Spec: esov1alpha1.ExternalServiceSpec{
			Port:           int32(port),
			Ips:            ips,
			Hosts:          hosts,
			ReadinessProbe: corev1.Probe{},
		},
	}
}

func CreateDefaultEndpoint() *corev1.Endpoints {
	return &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
		},
		Subsets: []corev1.EndpointSubset{
			corev1.EndpointSubset{
				NotReadyAddresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.14",
					},
					corev1.EndpointAddress{
						IP: "10.0.102.16",
					},
				},
				Addresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.10",
					},
					corev1.EndpointAddress{
						IP: "10.0.102.12",
					},
				},
				Ports: []corev1.EndpointPort{
					corev1.EndpointPort{
						Port: 80,
					},
				},
			},
		},
	}
}
