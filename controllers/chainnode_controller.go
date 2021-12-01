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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	citacloudv1 "github.com/cita-cloud/cita-cloud-operator/api/v1"
	cmd "github.com/cita-cloud/cita-cloud-operator/pkg/exec"
)

// ChainNodeReconciler reconciles a ChainNode object
type ChainNodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=citacloud.rivtower.com,resources=chainnodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=citacloud.rivtower.com,resources=chainnodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=citacloud.rivtower.com,resources=chainnodes/finalizers,verbs=update

func (r *ChainNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info(fmt.Sprintf("chainnode %s in reconcile", req.NamespacedName))

	chainNode := &citacloudv1.ChainNode{}
	if err := r.Get(ctx, req.NamespacedName, chainNode); err != nil {
		logger.Info(fmt.Sprintf("the chainnode %s has been deleted", req.NamespacedName))
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	chainConfig := &citacloudv1.ChainConfig{}

	if err := r.Get(ctx, types.NamespacedName{Name: chainNode.Spec.ChainName, Namespace: chainNode.Namespace}, chainConfig); err != nil {
		logger.Info(fmt.Sprintf("the chainconfig %s/%s has been deleted", req.Namespace, chainNode.Spec.ChainName))
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// init chain and init chain config
	cc := cmd.NewCloudConfig(chainConfig.Name, ".")
	if !cc.Exist() {
		err := cc.Init(chainConfig.Spec.Id)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// check admin account
	var nodeAddress string
	accountConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-%s-admin", chainConfig.Name, chainNode.Name), Namespace: chainNode.Namespace}, accountConfigMap)
	if err != nil && errors.IsNotFound(err) {
		keyId, nodeAddress, err := cc.CreateAccount(chainNode.Spec.KmsPassword)
		if err != nil {
			return ctrl.Result{}, err
		}
		accountConfigMap.ObjectMeta = metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-admin", chainConfig.Name, chainNode.Name),
			Namespace: chainNode.Namespace,
			Labels:    labelsForChain(chainNode.Name),
		}
		accountConfigMap.Data = map[string]string{
			"keyId":   keyId,
			"address": nodeAddress,
		}
		kmsDb, err := cc.ReadKmsDb(nodeAddress)
		if err != nil {
			return ctrl.Result{}, err
		}
		accountConfigMap.BinaryData = map[string][]byte{
			"kms.db": kmsDb,
		}

		// set ownerReference
		_ = ctrl.SetControllerReference(chainNode, accountConfigMap, r.Scheme)
		// create chain secret
		err = r.Create(ctx, accountConfigMap)
		if err != nil {
			return ctrl.Result{}, err
		}
		// requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "failed to get node address configmap")
		return ctrl.Result{}, err
	} else {
		nodeAddress = accountConfigMap.Data["address"]
	}

	if chainConfig.Spec.EnableTLS {
		nodeSecret := &corev1.Secret{}
		err := r.Get(ctx, types.NamespacedName{Name: chainNode.Name, Namespace: chainNode.Namespace}, nodeSecret)
		if err != nil && errors.IsNotFound(err) {
			// not found
			// 1. check ca secret
			chainSecret := &corev1.Secret{}
			err = r.Get(ctx, types.NamespacedName{Name: chainNode.Spec.ChainName, Namespace: chainNode.Namespace}, chainSecret)
			if err != nil && errors.IsNotFound(err) {
				// 1.1 ca secret not exist
				var cert, key []byte
				if !cc.Exist() {
					// 1.1.1 dir not exist, we will create dir and create ca secret
					err = cc.Init(chainConfig.Spec.Id)
					if err != nil {
						return ctrl.Result{}, err
					}
					cert, key, err = cc.CreateCaAndRead()
					if err != nil {
						return ctrl.Result{}, err
					}
				} else {
					// 1.1.2 dir exist, we will create ca secret
					// if any file not exist, will return error
					cert, key, err = cc.ReadCa()
				}
				if err != nil {
					return ctrl.Result{}, err
				}
				chainSecret.ObjectMeta = metav1.ObjectMeta{
					Name:      chainNode.Spec.ChainName,
					Namespace: chainNode.Namespace,
				}
				chainSecret.Data = map[string][]byte{
					"cert": cert,
					"key":  key,
				}
				// set ownerReference
				_ = ctrl.SetControllerReference(chainConfig, chainSecret, r.Scheme)
				// create chain secret
				err = r.Create(ctx, chainSecret)
				if err != nil {
					return ctrl.Result{}, err
				}
				// requeue
				return ctrl.Result{Requeue: true}, nil
			} else if err != nil {
				// 1.2 get ca secret error
				return ctrl.Result{}, err
			} else {
				// 1.3 ca secret exist
				// 1.3.1 dir not exist, we will recover
				if !cc.Exist() {
					err = cc.WriteCaCert(chainSecret.Data["cert"])
					if err != nil {
						return ctrl.Result{}, err
					}
					err = cc.WriteCaKey(chainSecret.Data["key"])
					if err != nil {
						return ctrl.Result{}, err
					}
				}
			}
			// this situation gen csr directly
			csr, key, cert, err := cc.CreateSignCsrAndRead(chainNode.Spec.Domain)
			if err != nil {
				return ctrl.Result{}, err
			}
			nodeSecret.ObjectMeta = metav1.ObjectMeta{
				Name:      chainNode.Name,
				Namespace: chainNode.Namespace,
			}
			nodeSecret.Data = map[string][]byte{
				"csr":  csr,
				"key":  key,
				"cert": cert,
			}
			// set ownerReference
			_ = ctrl.SetControllerReference(chainNode, nodeSecret, r.Scheme)
			// create node secret
			err = r.Create(ctx, nodeSecret)
			if err != nil {
				return ctrl.Result{}, err
			}
			// requeue
			return ctrl.Result{Requeue: true}, nil
		}
	}

	// update status
	if chainNode.Status.Address != nodeAddress {
		chainNode.Status.Address = nodeAddress
		err := r.Status().Update(ctx, chainNode)
		if err != nil {
			logger.Error(err, "Failed to update chain node status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ChainNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&citacloudv1.ChainNode{}).
		Complete(r)
}
