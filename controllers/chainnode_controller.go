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
	"time"

	citacloudv1 "github.com/cita-cloud/cita-cloud-operator/api/v1"
	cmd "github.com/cita-cloud/cita-cloud-operator/pkg/exec"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ChainNodeReconciler reconciles a ChainNode object
type ChainNodeReconciler struct {
	client.Client
	Scheme                *runtime.Scheme
	updateConfigFlag      bool
	updateStatefulSetFlag bool
	startUpdateTime       time.Time
	cmd                   cmd.Cmd
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

	if chainNode.Spec.Action == citacloudv1.NodeInitialize {
		oldChainNode := chainNode.DeepCopy()
		if err := r.SetDefaultSpec(ctx, chainConfig, chainNode); err != nil {
			return ctrl.Result{}, err
		}
		if !IsEqual(oldChainNode.Spec, chainNode.Spec) {
			diff, _ := DiffObject(oldChainNode, chainNode)
			logger.Info("SetDefault: " + string(diff))
			return ctrl.Result{}, r.Update(ctx, chainNode)
		}

		updated, err := r.SetDefaultStatus(ctx, chainNode)
		if updated || err != nil {
			return ctrl.Result{}, err
		}
	} else if chainNode.Spec.Action == citacloudv1.NodeCreate {
		// create all resource
		if err := r.ReconcileAllRecourse(ctx, chainConfig, chainNode); err != nil {
			return ctrl.Result{}, err
		}
		// sync status
		if err := r.SyncStatus(ctx, chainNode); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ChainNodeReconciler) SetDefaultSpec(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) error {
	logger := log.FromContext(ctx)

	// set cloud config init
	r.cmd = cmd.NewCloudConfig(chainConfig.Name, ".")
	if !r.cmd.Exist() {
		err := r.cmd.Init(chainConfig.Spec.Id)
		if err != nil {
			return err
		}
	}

	if chainNode.Spec.PullPolicy == "" {
		chainNode.Spec.PullPolicy = corev1.PullIfNotPresent
	}

	// set default images
	if chainNode.Spec.NetworkImage == "" {
		if chainConfig.Spec.EnableTLS {
			chainNode.Spec.NetworkImage = "citacloud/network_tls:v6.3.0"
		} else {
			chainNode.Spec.NetworkImage = "citacloud/network_p2p:v6.3.0"
		}
	}
	if chainNode.Spec.ConsensusImage == "" {
		if chainConfig.Spec.ConsensusType == citacloudv1.Raft {
			chainNode.Spec.ConsensusImage = "citacloud/consensus_raft:v6.3.0"
		} else if chainConfig.Spec.ConsensusType == citacloudv1.BFT {
			chainNode.Spec.ConsensusImage = "citacloud/consensus_bft:v6.3.0"
		} else {
			return fmt.Errorf("mismatched consensus type")
		}
	}
	if chainNode.Spec.ExecutorImage == "" {
		chainNode.Spec.ExecutorImage = "citacloud/executor_evm:v6.3.0"
	}
	if chainNode.Spec.StorageImage == "" {
		chainNode.Spec.StorageImage = "citacloud/storage_rocksdb:v6.3.0"
	}
	if chainNode.Spec.ControllerImage == "" {
		chainNode.Spec.ControllerImage = "citacloud/controller:v6.3.0"
	}
	if chainNode.Spec.KmsImage == "" {
		chainNode.Spec.KmsImage = "citacloud/kms_sm:v6.3.0"
	}

	//if chainNode.Spec.Address == "" {
	//	// init chain and init chain config
	//	cc := cmd.NewCloudConfig(chainConfig.Name, ".")
	//	if !cc.Exist() {
	//		err := cc.Init(chainConfig.Spec.Id)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//
	//	// check node account
	//	accountConfigMap := &corev1.ConfigMap{}
	//	err := r.Get(ctx, types.NamespacedName{Name: GetNodeAccountName(chainConfig.Name, chainNode.Name), Namespace: chainNode.Namespace}, accountConfigMap)
	//	if err != nil && errors.IsNotFound(err) {
	//
	//		keyId, nodeAddress, err := cc.CreateAccount(chainNode.Spec.KmsPassword)
	//		if err != nil {
	//			return err
	//		}
	//		accountConfigMap.ObjectMeta = metav1.ObjectMeta{
	//			Name:      GetNodeAccountName(chainConfig.Name, chainNode.Name),
	//			Namespace: chainNode.Namespace,
	//			Labels:    LabelsForChain(chainNode.Name),
	//		}
	//		accountConfigMap.Data = map[string]string{
	//			"keyId":   keyId,
	//			"address": nodeAddress,
	//		}
	//		kmsDb, err := cc.ReadKmsDb(nodeAddress)
	//		if err != nil {
	//			return err
	//		}
	//		accountConfigMap.BinaryData = map[string][]byte{
	//			"kms.db": kmsDb,
	//		}
	//
	//		// set ownerReference
	//		_ = ctrl.SetControllerReference(chainNode, accountConfigMap, r.Scheme)
	//		// create chain secret
	//		err = r.Create(ctx, accountConfigMap)
	//		if err != nil {
	//			return err
	//		}
	//		chainNode.Spec.Address = nodeAddress
	//		// requeue
	//		//return ctrl.Result{Requeue: true}, nil
	//
	//	} else if err != nil {
	//		logger.Error(err, "failed to get node address configmap")
	//		return err
	//	} else {
	//		chainNode.Spec.Address = accountConfigMap.Data["address"]
	//	}
	//}

	if chainNode.Spec.Address == "" {
		// ensure node account
		nodeAddress, err := r.ensureNodeAccount(ctx, chainConfig, chainNode)
		if err != nil {
			return err
		}
		// set chainNode.spec.address
		chainNode.Spec.Address = nodeAddress
	}

	// ensure node secret
	if err := r.ensureNodeCertAndKey(ctx, chainConfig, chainNode); err != nil {
		return err
	}

	// set ownerReference
	// todo: 需要生成账号后再进行绑定？
	if len(chainNode.OwnerReferences) == 0 {
		if err := ctrl.SetControllerReference(chainConfig, chainNode, r.Scheme); err != nil {
			logger.Error(err, "set chain node controller reference failed")
			return err
		}
		logger.Info("set chain node controller reference success")
	}
	return nil
}

func (r *ChainNodeReconciler) SetDefaultStatus(ctx context.Context, chainNode *citacloudv1.ChainNode) (bool, error) {
	logger := log.FromContext(ctx)
	if chainNode.Status.Status == "" {
		chainNode.Status.Status = citacloudv1.NodeInitialized
		err := r.Client.Status().Update(ctx, chainNode)
		if err != nil {
			return false, err
		}
		logger.Info("set chain node default status success")
		return true, nil
	}
	return false, nil
}

func (r *ChainNodeReconciler) ensureNodeCertAndKey(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) error {
	logger := log.FromContext(ctx)

	nodeCertAndKeySecret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: GetNodeCertAndKeySecretName(chainConfig.Name, chainNode.Name), Namespace: chainNode.Namespace}, nodeCertAndKeySecret)
	if err != nil && errors.IsNotFound(err) {
		csr, key, cert, err := r.cmd.CreateSignCsrAndRead(chainNode.Spec.Domain)
		if err != nil {
			return nil
		}
		nodeCertAndKeySecret.ObjectMeta = metav1.ObjectMeta{
			Name:      GetNodeCertAndKeySecretName(chainConfig.Name, chainNode.Name),
			Namespace: chainNode.Namespace,
			Labels:    LabelsForNode(chainConfig.Name, chainNode.Name),
		}
		nodeCertAndKeySecret.Data = map[string][]byte{
			NodeCert: cert,
			NodeCsr:  csr,
			NodeKey:  key,
		}
		// set ownerReference
		_ = ctrl.SetControllerReference(chainNode, nodeCertAndKeySecret, r.Scheme)
		// create chain secret
		err = r.Create(ctx, nodeCertAndKeySecret)
		if err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("chain node cert-key secret %s/%s created", chainNode.Namespace, chainNode.Name))
		return nil
	} else if err != nil {
		logger.Error(err, "failed to get node secret")
		return err
	}
	return nil
}

func (r *ChainNodeReconciler) ensureNodeAccount(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) (string, error) {
	logger := log.FromContext(ctx)

	accountConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: GetNodeAccountName(chainConfig.Name, chainNode.Name), Namespace: chainNode.Namespace}, accountConfigMap)
	if err != nil && errors.IsNotFound(err) {

		keyId, nodeAddress, err := r.cmd.CreateAccount(chainNode.Spec.KmsPassword)
		if err != nil {
			return "", err
		}
		accountConfigMap.ObjectMeta = metav1.ObjectMeta{
			Name:      GetNodeAccountName(chainConfig.Name, chainNode.Name),
			Namespace: chainNode.Namespace,
			Labels:    LabelsForChain(chainNode.Name),
		}
		accountConfigMap.Data = map[string]string{
			"keyId":   keyId,
			"address": nodeAddress,
		}
		kmsDb, err := r.cmd.ReadKmsDb(nodeAddress)
		if err != nil {
			return "", err
		}
		accountConfigMap.BinaryData = map[string][]byte{
			"kms.db": kmsDb,
		}

		// set ownerReference
		_ = ctrl.SetControllerReference(chainNode, accountConfigMap, r.Scheme)
		// create chain secret
		err = r.Create(ctx, accountConfigMap)
		if err != nil {
			return "", err
		}
		return nodeAddress, nil

	} else if err != nil {
		logger.Error(err, "failed to get node address configmap")
		return "", err
	} else {
		return accountConfigMap.Data["address"], nil
	}
}

func (r *ChainNodeReconciler) ReconcileAllRecourse(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) error {
	var err error
	// reconcile service
	if err := r.ReconcileService(ctx, chainConfig, chainNode); err != nil {
		return err
	}
	// reconcile node config
	r.updateConfigFlag, err = r.ReconcileConfigMap(ctx, chainConfig, chainNode)
	if err != nil {
		return err
	}
	// reconcile log config
	r.updateConfigFlag, err = r.ReconcileLogConfigMap(ctx, chainConfig, chainNode)
	if err != nil {
		return err
	}
	// reconcile statefulset
	r.updateStatefulSetFlag, err = r.ReconcileStatefulSet(ctx, chainConfig, chainNode)
	if err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ChainNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&citacloudv1.ChainNode{}).
		Owns(&appsv1.StatefulSet{}, builder.WithPredicates(r.statefulSetPredicates())).
		//Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}

func (r *ChainNodeReconciler) statefulSetPredicates() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(event event.UpdateEvent) bool {
			curSet := event.ObjectNew.(*appsv1.StatefulSet)
			oldSet := event.ObjectOld.(*appsv1.StatefulSet)
			if reflect.DeepEqual(curSet.Status, oldSet.Status) {
				return false
			}
			return true
		},
	}
}
