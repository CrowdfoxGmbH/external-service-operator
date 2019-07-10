package testutils

import (
	esov1alpha1 "github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func InitFakeClient(objs ...runtime.Object) client.Client {
	dummy := esov1alpha1.ExternalService{}

	s := scheme.Scheme
	s.AddKnownTypes(esov1alpha1.SchemeGroupVersion, &dummy)

	//I hate it when somebody uses globals instead ob requiring values via arguments
	return fake.NewFakeClient(objs...)
}
