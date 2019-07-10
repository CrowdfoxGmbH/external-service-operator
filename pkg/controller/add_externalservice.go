package controller

import (
	"github.com/CrowdfoxGmbH/external-service-operator/pkg/controller/externalservice"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, externalservice.Add)
}
