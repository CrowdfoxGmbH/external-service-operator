// +build !ignore_autogenerated

// Code generated by openapi-gen. DO NOT EDIT.

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1.ExternalService":       schema_pkg_apis_eso_v1alpha1_ExternalService(ref),
		"github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1.ExternalServiceSpec":   schema_pkg_apis_eso_v1alpha1_ExternalServiceSpec(ref),
		"github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1.ExternalServiceStatus": schema_pkg_apis_eso_v1alpha1_ExternalServiceStatus(ref),
	}
}

func schema_pkg_apis_eso_v1alpha1_ExternalService(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ExternalService is the Schema for the externalservices API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1.ExternalServiceSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1.ExternalServiceStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1.ExternalServiceSpec", "github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1.ExternalServiceStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_eso_v1alpha1_ExternalServiceSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ExternalServiceSpec defines the desired state of ExternalService",
				Properties:  map[string]spec.Schema{},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_eso_v1alpha1_ExternalServiceStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ExternalServiceStatus defines the observed state of ExternalService",
				Properties:  map[string]spec.Schema{},
			},
		},
		Dependencies: []string{},
	}
}
