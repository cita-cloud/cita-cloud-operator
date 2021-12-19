package controllers

import (
	"testing"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsEqual(t *testing.T) {
	sts1 := v1.StatefulSet{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "sts1",
		},
		Spec:       v1.StatefulSetSpec{},
		Status:     v1.StatefulSetStatus{},
	}
	sts2 := &v1.StatefulSet{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "sts2",
		},
		Spec:       v1.StatefulSetSpec{},
		Status:     v1.StatefulSetStatus{},
	}
	sts3 := v1.StatefulSet{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "sts1",
		},
		Spec:       v1.StatefulSetSpec{},
		Status:     v1.StatefulSetStatus{},
	}
	type args struct {
		obj1 interface{}
		obj2 interface{}
	}
	var tests []struct {
		name string
		args args
		want bool
	}
	tests = append(tests, struct {
		name string
		args args
		want bool
	}{name: "my-test1", args: struct {
		obj1 interface{}
		obj2 interface{}
	}{obj1: sts1, obj2: sts2}, want: false})
	tests = append(tests, struct {
		name string
		args args
		want bool
	}{name: "my-test2", args: struct {
		obj1 interface{}
		obj2 interface{}
	}{obj1: sts1, obj2: sts3}, want: true})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEqual(tt.args.obj1, tt.args.obj2); got != tt.want {
				t.Errorf("IsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}