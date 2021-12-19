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
	"reflect"

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

// ChainConfigReconciler reconciles a ChainConfig object
type ChainConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=citacloud.rivtower.com,resources=chainconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=citacloud.rivtower.com,resources=chainconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=citacloud.rivtower.com,resources=chainconfigs/finalizers,verbs=update

func (r *ChainConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info(fmt.Sprintf("chainconfig %s in reconcile", req.NamespacedName))

	chainConfig := &citacloudv1.ChainConfig{}

	if err := r.Get(ctx, req.NamespacedName, chainConfig); err != nil {
		logger.Info("chain  " + req.Name + " has been deleted")
		logger.Info(fmt.Sprintf("the chain %s has been deleted", req.NamespacedName))
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
	var adminAddress string
	adminConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: GetAdminAccountName(chainConfig.Name), Namespace: chainConfig.Namespace}, adminConfigMap)
	if err != nil && errors.IsNotFound(err) {
		keyId, adminAddress, err := cc.CreateAccount(chainConfig.Spec.KmsPassword)
		if err != nil {
			return ctrl.Result{}, err
		}
		adminConfigMap.ObjectMeta = metav1.ObjectMeta{
			Name:      GetAdminAccountName(chainConfig.Name),
			Namespace: chainConfig.Namespace,
			Labels:    LabelsForChain(chainConfig.Name),
		}
		adminConfigMap.Data = map[string]string{
			"keyId":   keyId,
			"address": adminAddress,
		}

		kmsDb, err := cc.ReadKmsDb(adminAddress)
		if err != nil {
			return ctrl.Result{}, err
		}
		adminConfigMap.BinaryData = map[string][]byte{
			"kms.db": kmsDb,
		}

		// set ownerReference
		_ = ctrl.SetControllerReference(chainConfig, adminConfigMap, r.Scheme)
		// create chain secret
		err = r.Create(ctx, adminConfigMap)
		if err != nil {
			return ctrl.Result{}, err
		}
		logger.Info(fmt.Sprintf("chain config configmap %s/%s created", chainConfig.Namespace, chainConfig.Name))
		// requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "failed to get admin address configmap")
		return ctrl.Result{}, err
	} else {
		adminAddress = adminConfigMap.Data["address"]
	}

	if chainConfig.Spec.AdminAddress == "" {
		chainConfig.Spec.AdminAddress = adminAddress
		if err := r.Update(ctx, chainConfig); err != nil {
			logger.Error(err, "update chain config admin address error")
			return ctrl.Result{}, err
		}
		logger.Info("set chain config admin address success")
		return ctrl.Result{Requeue: true}, nil
	}

	if chainConfig.Spec.EnableTLS {
		chainSecret := &corev1.Secret{}
		err = r.Get(ctx, types.NamespacedName{Name: chainConfig.Name, Namespace: chainConfig.Namespace}, chainSecret)
		if err != nil && errors.IsNotFound(err) {
			// not found
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
				Name:      chainConfig.Name,
				Namespace: chainConfig.Namespace,
				Labels:    LabelsForChain(chainConfig.Name),
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
			logger.Info(fmt.Sprintf("chain config secret %s/%s created", chainConfig.Namespace, chainConfig.Name))
			// requeue
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			logger.Error(err, "failed to get chain secret")
			return ctrl.Result{}, err
		}
	}

	var updateStatusFlag bool
	// list all chain's node
	chainNodeList := &citacloudv1.ChainNodeList{}
	opts := []client.ListOption{
		client.InNamespace(chainConfig.Namespace),
		client.MatchingFields{"spec.chainName": chainConfig.Name},
	}
	if err := r.List(ctx, chainNodeList, opts...); err != nil {
		return ctrl.Result{}, err
	}

	// set validators
	if updateValidators(chainConfig, chainNodeList.Items) {
		if err := r.Update(ctx, chainConfig); err != nil {
			logger.Error(err, "update chain config validator error")
			return ctrl.Result{}, err
		}
		logger.Info("set chain config validator success")
		return ctrl.Result{Requeue: true}, nil
	}

	nodeInfoMap := make(map[string]citacloudv1.NodeInfo, 0)
	for _, nl := range chainNodeList.Items {
		nodeInfoMap[nl.Name] = nl.Spec.NodeInfo
	}
	if !reflect.DeepEqual(chainConfig.Status.NodeInfoMap, nodeInfoMap) {
		chainConfig.Status.NodeInfoMap = nodeInfoMap
		updateStatusFlag = true
	}
	if updateStatusFlag {
		err := r.Status().Update(ctx, chainConfig)
		if err != nil {
			logger.Error(err, "failed to update chain config status")
			return ctrl.Result{}, err
		}
		logger.Info("update chain config status successful")
	}

	return ctrl.Result{}, nil
}

func updateValidators(config *citacloudv1.ChainConfig, chainNodes []citacloudv1.ChainNode) bool {
	var updateFlag bool
	if len(config.Spec.Validators) != len(chainNodes) {
		tmp := make([]string, 0)
		for _, node := range chainNodes {
			if node.Spec.Type == citacloudv1.Consensus {
				tmp = append(tmp, node.Spec.Address)
			}
		}
		config.Spec.Validators = tmp
		updateFlag = true
	}
	return updateFlag
}

// SetupWithManager sets up the controller with the Manager.
func (r *ChainConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&citacloudv1.ChainConfig{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&citacloudv1.ChainNode{}).
		Complete(r)
}
