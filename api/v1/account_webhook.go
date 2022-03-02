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
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var accountlog = logf.Log.WithName("account-resource")
var cli client.Client

func (r *Account) SetupWebhookWithManager(mgr ctrl.Manager) error {
	// todo: ugly code
	cli = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-citacloud-rivtower-com-v1-account,mutating=true,failurePolicy=fail,sideEffects=None,groups=citacloud.rivtower.com,resources=accounts,verbs=create;update,versions=v1,name=maccount.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Account{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Account) Default() {
	accountlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-citacloud-rivtower-com-v1-account,mutating=false,failurePolicy=fail,sideEffects=None,groups=citacloud.rivtower.com,resources=accounts,verbs=create;update,versions=v1,name=vaccount.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Account{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Account) ValidateCreate() error {
	accountlog.Info("validate create", "name", r.Name)
	err := r.validate()
	if err != nil {
		return err
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Account) ValidateUpdate(old runtime.Object) error {
	accountlog.Info("validate update", "name", r.Name)
	err := r.validate()
	if err != nil {
		return err
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Account) ValidateDelete() error {
	accountlog.Info("validate delete", "name", r.Name)
	return nil
}

func (r *Account) validate() error {
	chainConfig := &ChainConfig{}
	err := cli.Get(context.Background(), types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Spec.Chain,
	}, chainConfig)
	if err != nil {
		return err
	}
	if r.Spec.Role != Admin && r.Spec.Domain == "" {
		return fmt.Errorf("domain field cann't be empty")
	}
	return nil
}
