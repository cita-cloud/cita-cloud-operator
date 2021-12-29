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

	updated, err := r.SetDefaultSpec(ctx, chainConfig)
	if updated || err != nil {
		return ctrl.Result{}, err
	}

	updated, err = r.SetDefaultStatus(ctx, chainConfig)
	if updated || err != nil {
		return ctrl.Result{}, err
	}

	// init chain and init chain config
	//cc := cmd.NewCloudConfig(chainConfig.Name, ".")
	//if !cc.Exist() {
	//	logger.Info(fmt.Sprintf("config dir %s/%s not found, we will init it", ".", chainConfig.Name))
	//	err := cc.Init(chainConfig.Spec.Id)
	//	if err != nil {
	//		return ctrl.Result{}, err
	//	}
	//}

	//if chainConfig.Spec.EnableTLS {
	//	caSecret := &corev1.Secret{}
	//	err := r.Get(ctx, types.NamespacedName{Name: GetCaSecretName(chainConfig.Name), Namespace: chainConfig.Namespace}, caSecret)
	//	if err != nil && errors.IsNotFound(err) {
	//		// not found
	//		// 1.1 ca secret not exist
	//		var cert, key []byte
	//		if !cc.Exist() {
	//			logger.Info(fmt.Sprintf("config dir %s/%s not found, we will init it", ".", chainConfig.Name))
	//			// 1.1.1 dir not exist, we will create dir and create ca secret
	//			err = cc.Init(chainConfig.Spec.Id)
	//			if err != nil {
	//				return ctrl.Result{}, err
	//			}
	//			cert, key, err = cc.CreateCaAndRead()
	//			if err != nil {
	//				return ctrl.Result{}, err
	//			}
	//		} else {
	//			// 1.1.2 dir exist, we will create ca secret
	//			// if any file not exist, will return error
	//			cert, key, err = cc.CreateCaAndRead()
	//		}
	//		if err != nil {
	//			return ctrl.Result{}, err
	//		}
	//		caSecret.ObjectMeta = metav1.ObjectMeta{
	//			Name:      GetCaSecretName(chainConfig.Name),
	//			Namespace: chainConfig.Namespace,
	//			Labels:    LabelsForChain(chainConfig.Name),
	//		}
	//		caSecret.Data = map[string][]byte{
	//			CaCert: cert,
	//			CaKey:  key,
	//		}
	//		// set ownerReference
	//		_ = ctrl.SetControllerReference(chainConfig, caSecret, r.Scheme)
	//		// create chain secret
	//		err = r.Create(ctx, caSecret)
	//		if err != nil {
	//			return ctrl.Result{}, err
	//		}
	//		logger.Info(fmt.Sprintf("chain config ca secret %s/%s created", chainConfig.Namespace, chainConfig.Name))
	//		// requeue
	//		//return ctrl.Result{Requeue: true}, nil
	//	} else if err != nil {
	//		logger.Error(err, "failed to get chain ca secret")
	//		return ctrl.Result{}, err
	//	}
	//}
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
			chainConfig.Status.AdminAccount = aai
			updateFlag = true
		}

		// 查询共识账户
		OrdinaryAccountList := &citacloudv1.AccountList{}
		OrdinaryAccountOpts := []client.ListOption{
			client.InNamespace(chainConfig.Namespace),
			client.MatchingFields{"spec.chain": chainConfig.Name},
			client.MatchingFields{"spec.role": string(citacloudv1.Ordinary)},
		}
		if err := r.List(ctx, OrdinaryAccountList, OrdinaryAccountOpts...); err != nil {
			return ctrl.Result{}, err
		}
		vaiMap := make(map[string]citacloudv1.ValidatorAccountInfo, 0)
		for _, acc := range OrdinaryAccountList.Items {
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
	}

	//var updateStatusFlag bool
	//// list all chain's node
	//chainNodeList := &citacloudv1.ChainNodeList{}
	//opts := []client.ListOption{
	//	client.InNamespace(chainConfig.Namespace),
	//	client.MatchingFields{"spec.chainName": chainConfig.Name},
	//}
	//if err := r.List(ctx, chainNodeList, opts...); err != nil {
	//	return ctrl.Result{}, err
	//}
	//
	//// set validators
	//if updateValidators(chainConfig, chainNodeList.Items) {
	//	if err := r.Update(ctx, chainConfig); err != nil {
	//		logger.Error(err, "update chain config validator error")
	//		return ctrl.Result{}, err
	//	}
	//	logger.Info("set chain config validator success")
	//	return ctrl.Result{RequeueAfter: time.Duration(3) * time.Second}, nil
	//}
	//
	//nodeInfoMap := make(map[string]citacloudv1.NodeInfo, 0)
	//for _, nl := range chainNodeList.Items {
	//	// set node status
	//	nl.Spec.NodeInfo.Status = nl.Status.Status
	//	nodeInfoMap[nl.Name] = nl.Spec.NodeInfo
	//}
	//if !reflect.DeepEqual(chainConfig.Status.NodeInfoMap, nodeInfoMap) {
	//	chainConfig.Status.NodeInfoMap = nodeInfoMap
	//	updateStatusFlag = true
	//}
	//if updateStatusFlag {
	//	err := r.Status().Update(ctx, chainConfig)
	//	if err != nil {
	//		logger.Error(err, "failed to update chain config status")
	//		return ctrl.Result{}, err
	//	}
	//	logger.Info("update chain config status successful")
	//}

	return ctrl.Result{}, nil
}

//func updateValidators(config *citacloudv1.ChainConfig, chainNodes []citacloudv1.ChainNode) bool {
//	var updateFlag bool
//	if len(config.Spec.Validators) != len(chainNodes) {
//		tmp := make([]string, 0)
//		for _, node := range chainNodes {
//			if node.Spec.Type == citacloudv1.Consensus {
//				tmp = append(tmp, node.Spec.Address)
//			}
//		}
//		config.Spec.Validators = tmp
//		updateFlag = true
//	}
//	return updateFlag
//}

func (r *ChainConfigReconciler) SetDefaultStatus(ctx context.Context, chainConfig *citacloudv1.ChainConfig) (bool, error) {
	logger := log.FromContext(ctx)
	if chainConfig.Status.Status == "" {
		chainConfig.Status.Status = citacloudv1.Publicizing
		err := r.Client.Status().Update(ctx, chainConfig)
		if err != nil {
			logger.Error(err, "set chain config default status failed")
			return false, err
		}
		logger.Info("set chain config default status success")
		return true, nil
	}
	return false, nil
}

func (r *ChainConfigReconciler) SetDefaultSpec(ctx context.Context, chainConfig *citacloudv1.ChainConfig) (bool, error) {
	logger := log.FromContext(ctx)
	if chainConfig.Spec.Action == "" {
		chainConfig.Spec.Action = citacloudv1.Publicizing
		err := r.Client.Update(ctx, chainConfig)
		if err != nil {
			logger.Error(err, "set chain config default action value failed")
			return false, err
		}
		logger.Info("set chain config default action value success")
		return true, nil
	}
	return false, nil
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
