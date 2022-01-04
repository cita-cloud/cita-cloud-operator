package controllers

import (
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IsEqual check two object is equal.
func IsEqual(obj1, obj2 interface{}) bool {
	return equality.Semantic.DeepEqual(obj1, obj2)
}

// DiffObject
func DiffObject(old, new client.Object) ([]byte, error) {
	return client.MergeFrom(old).Data(new)
}
