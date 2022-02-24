/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var chainnodelog = logf.Log.WithName("chainnode-resource")

func (r *ChainNode) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-citacloud-rivtower-com-v1-chainnode,mutating=true,failurePolicy=fail,sideEffects=None,groups=citacloud.rivtower.com,resources=chainnodes,verbs=create;update,versions=v1,name=mchainnode.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &ChainNode{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ChainNode) Default() {
	chainnodelog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-citacloud-rivtower-com-v1-chainnode,mutating=false,failurePolicy=fail,sideEffects=None,groups=citacloud.rivtower.com,resources=chainnodes,verbs=create;update,versions=v1,name=vchainnode.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ChainNode{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ChainNode) ValidateCreate() error {
	chainnodelog.Info("validate create", "name", r.Name)
	err := r.validate()
	if err != nil {
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ChainNode) ValidateUpdate(old runtime.Object) error {
	chainnodelog.Info("validate update", "name", r.Name)
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ChainNode) ValidateDelete() error {
	chainnodelog.Info("validate delete", "name", r.Name)
	return nil
}

func (r *ChainNode) validate() error {
	account := &Account{}
	err := cli.Get(context.Background(), types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Spec.Account,
	}, account)
	if err != nil {
		return err
	}
	if r.Spec.Status != "" {
		return fmt.Errorf("cann't set ChainNode.Spec.Status field")
	}
	return nil
}
