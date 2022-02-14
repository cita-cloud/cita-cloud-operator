/*
 * Copyright Rivtower Technologies LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package controllers

import (
	"testing"
	"time"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsEqual(t *testing.T) {
	sts1 := v1.StatefulSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "sts1",
		},
		Spec:   v1.StatefulSetSpec{},
		Status: v1.StatefulSetStatus{},
	}
	sts2 := &v1.StatefulSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "sts2",
		},
		Spec:   v1.StatefulSetSpec{},
		Status: v1.StatefulSetStatus{},
	}
	sts3 := v1.StatefulSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "sts1",
		},
		Spec:   v1.StatefulSetSpec{},
		Status: v1.StatefulSetStatus{},
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

func TestTime(t *testing.T) {
	a := time.Now()
	time.Sleep(time.Duration(3) * time.Second)
	if time.Since(a).Seconds() >= 10 {
		t.Log(">10")
	}
	t.Log("<10")
}
