package externalservice

import (
	esov1alpha1 "github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/CrowdfoxGmbH/external-service-operator/pkg/testutils"
	"testing"
)

func TestCreateEndpointsCr(t *testing.T) {

	endpoint := CreateEndpointsCr(getTestExternalServiceCR())

	//check name is set
	testutils.ExpectEqStr(endpoint.Name, "TestService", t)
	testutils.ExpectEqStr(endpoint.Namespace, "external-services", t)

	testutils.ExpectEqStr(endpoint.Subsets[0].NotReadyAddresses[0].IP, "10.0.100.10", t)
	testutils.ExpectEqStr(endpoint.Subsets[0].NotReadyAddresses[1].IP, "10.0.100.11", t)
	testutils.ExpectEqStr(endpoint.Subsets[0].NotReadyAddresses[2].IP, "10.0.100.12", t)

	testutils.ExpectEqInt(endpoint.Subsets[0].Ports[0].Port, 80, t)

}

func TestMergeEndpointWithExternalServiceDef_IpCameAvailable(t *testing.T) {
	oldEndpoint := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
		},
		Subsets: []corev1.EndpointSubset{
			corev1.EndpointSubset{
				NotReadyAddresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.12",
					},
				},
				Addresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.10",
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

	externalService := &esov1alpha1.ExternalService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
			Annotations: map[string]string{
				"foo.bar": "testvalue",
			},
		},

		Spec: esov1alpha1.ExternalServiceSpec{
			Port: 80,
			Ips: []string{
				"10.0.102.10",
				"10.0.102.12",
			},
			Hosts: []esov1alpha1.ExternalServiceHostPath{
				esov1alpha1.ExternalServiceHostPath{
					Host: "subdomain.example.com",
				},
				esov1alpha1.ExternalServiceHostPath{
					Host: "another.domain.com",
					Path: "/foo",
				},
			},
		},
	}
	actualEndpoint, changed := mergeEndpointWithExternalServiceDef(externalService, oldEndpoint)
	testutils.ExpectEqualEndpoints(*actualEndpoint, *oldEndpoint, t)
	testutils.ExpectFalse(changed, t)
}

func TestMergeEndpointWithExternalServiceDef_AddIp(t *testing.T) {
	oldEndpoint := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
		},
		Subsets: []corev1.EndpointSubset{
			corev1.EndpointSubset{
				NotReadyAddresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.11",
					},
				},
				Addresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.10",
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

	expectedEndpoint := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
		},
		Subsets: []corev1.EndpointSubset{
			corev1.EndpointSubset{
				NotReadyAddresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.11",
					},
					corev1.EndpointAddress{
						IP: "10.0.102.12",
					},
				},
				Addresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.10",
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

	externalService := &esov1alpha1.ExternalService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
			Annotations: map[string]string{
				"foo.bar": "testvalue",
			},
		},

		Spec: esov1alpha1.ExternalServiceSpec{
			Port: 80,
			Ips: []string{
				"10.0.102.10",
				"10.0.102.11",
				"10.0.102.12",
			},
			Hosts: []esov1alpha1.ExternalServiceHostPath{
				esov1alpha1.ExternalServiceHostPath{
					Host: "subdomain.example.com",
				},
				esov1alpha1.ExternalServiceHostPath{
					Host: "another.domain.com",
					Path: "/foo",
				},
			},
		},
	}

	actualEndpoint, changed := mergeEndpointWithExternalServiceDef(externalService, oldEndpoint)
	testutils.ExpectEqualEndpoints(*actualEndpoint, *expectedEndpoint, t)
	testutils.ExpectTrue(changed, t)
}

func TestMergeEndpointWithExternalServiceDef_RemoveIp(t *testing.T) {
	oldEndpoint := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
		},
		Subsets: []corev1.EndpointSubset{
			corev1.EndpointSubset{
				NotReadyAddresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.11",
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

	expectedEndpoint := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
		},
		Subsets: []corev1.EndpointSubset{
			corev1.EndpointSubset{
				NotReadyAddresses: []corev1.EndpointAddress{},
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

	externalService := &esov1alpha1.ExternalService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
			Annotations: map[string]string{
				"foo.bar": "testvalue",
			},
		},

		Spec: esov1alpha1.ExternalServiceSpec{
			Port: 80,
			Ips: []string{
				"10.0.102.10",
				"10.0.102.12",
			},
			Hosts: []esov1alpha1.ExternalServiceHostPath{
				esov1alpha1.ExternalServiceHostPath{
					Host: "subdomain.example.com",
				},
				esov1alpha1.ExternalServiceHostPath{
					Host: "another.domain.com",
					Path: "/foo",
				},
			},
		},
	}

	actualEndpoint, changed := mergeEndpointWithExternalServiceDef(externalService, oldEndpoint)
	testutils.ExpectEqualEndpoints(*actualEndpoint, *expectedEndpoint, t)
	testutils.ExpectTrue(changed, t)
}

func TestMergeEndpointWithExternalServiceDef_changePort(t *testing.T) {
	externalService := getTestExternalServiceCR()
	endpoint := CreateEndpointsCr(externalService)

	externalService.Spec.Port = 9090
	expectedEndpoint := CreateEndpointsCr(externalService)

	actualEndpoint, _ := mergeEndpointWithExternalServiceDef(externalService, endpoint)
	testutils.ExpectEqInt(actualEndpoint.Subsets[0].Ports[0].Port, expectedEndpoint.Subsets[0].Ports[0].Port, t)
}

func TestEqualIgnoreReadyExactSame(t *testing.T) {
	oldEndpoint := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
		},
		Subsets: []corev1.EndpointSubset{
			corev1.EndpointSubset{
				NotReadyAddresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.12",
					},
				},
				Addresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.10",
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

	newEndpoint := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
		},
		Subsets: []corev1.EndpointSubset{
			corev1.EndpointSubset{
				NotReadyAddresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.12",
					},
				},
				Addresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.10",
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

	equals := equalIgnoreReady(oldEndpoint, newEndpoint)
	testutils.ExpectTrue(equals, t)
}

func TestEqualIgnoreReadyIpsBecameAllReady(t *testing.T) {
	oldEndpoint := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
		},
		Subsets: []corev1.EndpointSubset{
			corev1.EndpointSubset{
				NotReadyAddresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.12",
					},
				},
				Addresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.10",
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

	newEndpoint := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
		},
		Subsets: []corev1.EndpointSubset{
			corev1.EndpointSubset{
				NotReadyAddresses: []corev1.EndpointAddress{},
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

	equals := equalIgnoreReady(oldEndpoint, newEndpoint)
	testutils.ExpectTrue(equals, t)
}

func TestEqualIgnoreReadyIpChanged(t *testing.T) {
	oldEndpoint := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
		},
		Subsets: []corev1.EndpointSubset{
			corev1.EndpointSubset{
				NotReadyAddresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.12",
					},
				},
				Addresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.10",
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

	newEndpoint := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
		},
		Subsets: []corev1.EndpointSubset{
			corev1.EndpointSubset{
				NotReadyAddresses: []corev1.EndpointAddress{},
				Addresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: "10.0.102.11",
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

	equals := equalIgnoreReady(oldEndpoint, newEndpoint)
	testutils.ExpectFalse(equals, t)
}
