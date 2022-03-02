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
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var chainconfiglog = logf.Log.WithName("chainconfig-resource")

func (r *ChainConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-citacloud-rivtower-com-v1-chainconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=citacloud.rivtower.com,resources=chainconfigs,verbs=create;update,versions=v1,name=mchainconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &ChainConfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ChainConfig) Default() {
	chainconfiglog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-citacloud-rivtower-com-v1-chainconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=citacloud.rivtower.com,resources=chainconfigs,verbs=create;update,versions=v1,name=vchainconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ChainConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ChainConfig) ValidateCreate() error {
	chainconfiglog.Info("validate create", "name", r.Name)
	if r.Spec.Action != Publicizing {
		return fmt.Errorf("ChainConfig.Spec.Action should be set to Publicizing at the time of creation")
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ChainConfig) ValidateUpdate(old runtime.Object) error {
	chainconfiglog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ChainConfig) ValidateDelete() error {
	chainconfiglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
