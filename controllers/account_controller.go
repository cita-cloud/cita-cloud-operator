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

package controllers

import (
	"context"
	"fmt"

	cmd "github.com/cita-cloud/cita-cloud-operator/pkg/exec"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	citacloudv1 "github.com/cita-cloud/cita-cloud-operator/api/v1"
)

// AccountReconciler reconciles a Account object
type AccountReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=citacloud.rivtower.com,resources=accounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=citacloud.rivtower.com,resources=accounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=citacloud.rivtower.com,resources=accounts/finalizers,verbs=update

func (r *AccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info(fmt.Sprintf("account %s in reconcile", req.NamespacedName))

	account := &citacloudv1.Account{}
	if err := r.Get(ctx, req.NamespacedName, account); err != nil {
		logger.Info("account  " + req.Name + " has been deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// find the chain
	chainConfig := &citacloudv1.ChainConfig{}
	if err := r.Get(ctx, types.NamespacedName{Name: account.Spec.Chain, Namespace: account.Namespace}, chainConfig); err != nil {
		logger.Info(fmt.Sprintf("the chainconfig %s/%s has been deleted", req.Namespace, account.Spec.Chain))
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// init chain and init chain config
	cc := cmd.NewCloudConfig(chainConfig.Name, ".")
	if !cc.Exist() {
		logger.Info(fmt.Sprintf("config dir %s/%s not found, we will init it", ".", chainConfig.Name))
		err := cc.Init(chainConfig.Spec.Id)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// check account
	var accountAddress string
	accountConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: GetAccountConfigmap(account.Name), Namespace: account.Namespace}, accountConfigMap)
	if err != nil && errors.IsNotFound(err) {
		keyId, address, err := cc.CreateAccount(account.Spec.KmsPassword)
		accountAddress = address
		if err != nil {
			return ctrl.Result{}, err
		}

		accountConfigMap.ObjectMeta = metav1.ObjectMeta{
			Name:      GetAccountConfigmap(account.Name),
			Namespace: account.Namespace,
			Labels:    LabelsForChain(chainConfig.Name),
		}
		accountConfigMap.Data = map[string]string{
			"keyId":   keyId,
			"address": accountAddress,
		}

		kmsDb, err := cc.ReadKmsDb(accountAddress)
		if err != nil {
			return ctrl.Result{}, err
		}
		accountConfigMap.BinaryData = map[string][]byte{
			"kms.db": kmsDb,
		}

		// set ownerReference
		_ = ctrl.SetControllerReference(account, accountConfigMap, r.Scheme)
		// create chain secret
		err = r.Create(ctx, accountConfigMap)
		if err != nil {
			return ctrl.Result{}, err
		}
		logger.Info(fmt.Sprintf("account configmap %s/%s created", account.Name, account.Namespace))
	} else if err != nil {
		logger.Error(err, "failed to get account address configmap")
		return ctrl.Result{}, err
	} else {
		accountAddress = accountConfigMap.Data["address"]
	}

	// Generate certificate directly
	var caCertResult []byte
	if account.Spec.Role == citacloudv1.Admin {
		//var certResult []byte
		caSecret := &corev1.Secret{}
		err = r.Get(ctx, types.NamespacedName{Name: GetCaSecretName(chainConfig.Name), Namespace: chainConfig.Namespace}, caSecret)
		if err != nil && errors.IsNotFound(err) {
			cert, key, err := cc.CreateCaAndRead()
			caCertResult = cert
			if err != nil {
				return ctrl.Result{}, err
			}
			caSecret.ObjectMeta = metav1.ObjectMeta{
				Name:      GetCaSecretName(chainConfig.Name),
				Namespace: chainConfig.Namespace,
				Labels:    LabelsForChain(chainConfig.Name),
			}
			caSecret.Data = map[string][]byte{
				CaCert: cert,
				CaKey:  key,
			}
			// set ownerReference
			_ = ctrl.SetControllerReference(account, caSecret, r.Scheme)
			// create ca secret
			err = r.Create(ctx, caSecret)
			if err != nil {
				return ctrl.Result{}, err
			}
			logger.Info(fmt.Sprintf("ca secret %s/%s created", chainConfig.Namespace, chainConfig.Name))
		} else if err != nil {
			logger.Error(err, "failed to get chain ca secret")
			return ctrl.Result{}, err
		} else {
			caCertResult = caSecret.Data[CaCert]
		}

	} else {
		// find ca secret
		caSecret := &corev1.Secret{}
		err = r.Get(ctx, types.NamespacedName{Name: GetCaSecretName(chainConfig.Name), Namespace: chainConfig.Namespace}, caSecret)
		if err != nil {
			logger.Error(err, "get ca secret error when create account cert and key")
			return ctrl.Result{}, err
		}

		accountCertAndKeySecret := &corev1.Secret{}
		err := r.Get(ctx, types.NamespacedName{Name: GetAccountCertAndKeySecretName(account.Name), Namespace: account.Namespace}, accountCertAndKeySecret)
		if err != nil && errors.IsNotFound(err) {
			csr, key, cert, err := cc.CreateSignCsrAndRead(account.Spec.Domain)
			if err != nil {
				return ctrl.Result{}, err
			}
			accountCertAndKeySecret.ObjectMeta = metav1.ObjectMeta{
				Name:      GetAccountCertAndKeySecretName(account.Name),
				Namespace: account.Namespace,
				Labels:    LabelsForChain(chainConfig.Name),
			}
			accountCertAndKeySecret.Data = map[string][]byte{
				NodeCert: cert,
				NodeCsr:  csr,
				NodeKey:  key,
			}
			// set ownerReference
			_ = ctrl.SetControllerReference(account, accountCertAndKeySecret, r.Scheme)
			// create chain secret
			err = r.Create(ctx, accountCertAndKeySecret)
			if err != nil {
				return ctrl.Result{}, err
			}
			logger.Info(fmt.Sprintf("cert-key secret %s/%s created", account.Namespace, account.Name))
		} else if err != nil {
			logger.Error(err, "failed to get account cert and key secret")
			return ctrl.Result{}, err
		}
	}

	if account.Status.Cert != string(caCertResult) || account.Status.Address != accountAddress {
		account.Status.Cert = string(caCertResult)
		account.Status.Address = accountAddress
		err = r.Client.Status().Update(ctx, account)
		if err != nil {
			logger.Error(err, "update account status error")
			return ctrl.Result{}, err
		}
		logger.Info("update account status success")
		return ctrl.Result{Requeue: true}, nil
	}

	// bind chain config resource
	if len(account.OwnerReferences) == 0 {
		if err := ctrl.SetControllerReference(chainConfig, account, r.Scheme); err != nil {
			logger.Error(err, "set account controller reference failed")
			return ctrl.Result{}, err
		}
		err = r.Update(ctx, account)
		if err != nil {
			logger.Error(err, "set account controller reference failed")
			return ctrl.Result{}, err
		}
		logger.Info("set account controller reference success")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&citacloudv1.Account{}).
		Complete(r)
}
