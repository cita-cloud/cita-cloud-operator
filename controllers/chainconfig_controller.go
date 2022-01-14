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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	citacloudv1 "github.com/cita-cloud/cita-cloud-operator/api/v1"
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

	oldChainConfig := chainConfig.DeepCopy()
	if err := r.SetDefaultSpec(chainConfig); err != nil {
		return ctrl.Result{}, err
	}
	if !IsEqual(oldChainConfig.Spec, chainConfig.Spec) {
		diff, _ := DiffObject(oldChainConfig, chainConfig)
		logger.Info("ChainConfig setDefault: " + string(diff))
		return ctrl.Result{}, r.Update(ctx, chainConfig)
	}

	//updated, err := r.SetDefaultSpec(ctx, chainConfig)
	//if updated || err != nil {
	//	return ctrl.Result{}, err
	//}

	updated, err := r.SetDefaultStatus(ctx, chainConfig)
	if updated || err != nil {
		return ctrl.Result{}, err
	}

	if chainConfig.Spec.Action != chainConfig.Status.Status {
		chainConfig.Status.Status = chainConfig.Spec.Action
		err := r.Status().Update(ctx, chainConfig)
		if err != nil {
			logger.Error(err, fmt.Sprintf("update chain config status [%s] failed", chainConfig.Spec.Action))
			return ctrl.Result{}, err
		}
		logger.Info(fmt.Sprintf("update chain config status [%s] success", chainConfig.Spec.Action))
		return ctrl.Result{}, nil
	}

	if chainConfig.Status.Status == citacloudv1.Publicizing {
		adminAccountList := &citacloudv1.AccountList{}
		adminAccountOpts := []client.ListOption{
			client.InNamespace(chainConfig.Namespace),
			client.MatchingFields{"spec.chain": chainConfig.Name},
			client.MatchingFields{"spec.role": string(citacloudv1.Admin)},
		}
		if err := r.List(ctx, adminAccountList, adminAccountOpts...); err != nil {
			return ctrl.Result{}, err
		}
		if len(adminAccountList.Items) > 1 {
			err := fmt.Errorf("admin account count > 1")
			logger.Error(err, "admin account count > 1")
			return ctrl.Result{}, err
		}

		var updateFlag bool
		var aai citacloudv1.AdminAccountInfo
		if len(adminAccountList.Items) == 1 {
			aai.Address = adminAccountList.Items[0].Status.Address
			aai.Name = adminAccountList.Items[0].Name
		}
		if !reflect.DeepEqual(chainConfig.Status.AdminAccount, aai) {
			chainConfig.Status.AdminAccount = &aai
			updateFlag = true
		}

		// 查询共识账户
		ConsensusAccountList := &citacloudv1.AccountList{}
		ConsensusAccountOpts := []client.ListOption{
			client.InNamespace(chainConfig.Namespace),
			client.MatchingFields{"spec.chain": chainConfig.Name},
			client.MatchingFields{"spec.role": string(citacloudv1.Consensus)},
		}
		if err := r.List(ctx, ConsensusAccountList, ConsensusAccountOpts...); err != nil {
			return ctrl.Result{}, err
		}
		vaiMap := make(map[string]citacloudv1.ValidatorAccountInfo, 0)
		for _, acc := range ConsensusAccountList.Items {
			vai := citacloudv1.ValidatorAccountInfo{Address: acc.Status.Address}
			vaiMap[acc.Name] = vai
		}

		if !reflect.DeepEqual(chainConfig.Status.ValidatorAccountMap, vaiMap) {
			chainConfig.Status.ValidatorAccountMap = vaiMap
			updateFlag = true
		}
		if updateFlag {
			err := r.Status().Update(ctx, chainConfig)
			if err != nil {
				logger.Error(err, "failed to update chain config status")
				return ctrl.Result{}, err
			}
			logger.Info("update chain config status successful")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
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
	nodeInfoMap := make(map[string]citacloudv1.NodeInfo, 0)
	for _, nl := range chainNodeList.Items {
		// set node status
		nl.Spec.NodeInfo.Status = nl.Status.Status
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

func (r *ChainConfigReconciler) SetDefaultStatus(ctx context.Context, chainConfig *citacloudv1.ChainConfig) (bool, error) {
	logger := log.FromContext(ctx)
	if chainConfig.Status.Status == "" {
		chainConfig.Status.Status = citacloudv1.Publicizing
		err := r.Client.Status().Update(ctx, chainConfig)
		if err != nil {
			logger.Error(err, fmt.Sprintf("set chain config default status [%s] failed", chainConfig.Status.Status))
			return false, err
		}
		logger.Info(fmt.Sprintf("set chain config default status [%s] success", chainConfig.Status.Status))
		return true, nil
	}
	return false, nil
}

func (r *ChainConfigReconciler) SetDefaultSpec(chainConfig *citacloudv1.ChainConfig) error {
	//logger := log.FromContext(ctx)

	if chainConfig.Spec.Action == "" {
		chainConfig.Spec.Action = citacloudv1.Publicizing
	}

	if chainConfig.Spec.PullPolicy == "" {
		chainConfig.Spec.PullPolicy = corev1.PullIfNotPresent
	}

	// set default images
	if chainConfig.Spec.NetworkImage == "" {
		if chainConfig.Spec.EnableTLS {
			chainConfig.Spec.NetworkImage = "citacloud/network_tls:v6.3.0"
		} else {
			chainConfig.Spec.NetworkImage = "citacloud/network_p2p:v6.3.0"
		}
	}
	if chainConfig.Spec.ConsensusImage == "" {
		if chainConfig.Spec.ConsensusType == citacloudv1.Raft {
			chainConfig.Spec.ConsensusImage = "citacloud/consensus_raft:v6.3.0"
		} else if chainConfig.Spec.ConsensusType == citacloudv1.BFT {
			chainConfig.Spec.ConsensusImage = "citacloud/consensus_bft:v6.3.0"
		} else {
			return fmt.Errorf("mismatched consensus type")
		}
	}
	if chainConfig.Spec.ExecutorImage == "" {
		chainConfig.Spec.ExecutorImage = "citacloud/executor_evm:v6.3.0"
	}
	if chainConfig.Spec.StorageImage == "" {
		chainConfig.Spec.StorageImage = "citacloud/storage_rocksdb:v6.3.0"
	}
	if chainConfig.Spec.ControllerImage == "" {
		chainConfig.Spec.ControllerImage = "citacloud/controller:v6.3.0"
	}
	if chainConfig.Spec.KmsImage == "" {
		chainConfig.Spec.KmsImage = "citacloud/kms_sm:v6.3.0"
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ChainConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&citacloudv1.ChainConfig{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&citacloudv1.ChainNode{}).
		Owns(&citacloudv1.Account{}).
		Complete(r)
}
