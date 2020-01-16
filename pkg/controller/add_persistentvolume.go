package controller

import (
	"github.com/sstarcher/kube-ebs-tagger/pkg/controller/persistentvolume"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, persistentvolume.Add)
}
