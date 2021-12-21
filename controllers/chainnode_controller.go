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
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
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

	if chainNode.Spec.PullPolicy == "" {
		chainNode.Spec.PullPolicy = corev1.PullIfNotPresent
	}

	// set default images
	if chainNode.Spec.NetworkImage == "" {
		chainNode.Spec.NetworkImage = "citacloud/network_p2p:v6.3.0"
	}
	if chainNode.Spec.ConsensusImage == "" {
		chainNode.Spec.ConsensusImage = "citacloud/consensus_raft:v6.3.0"
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

	if chainNode.Spec.Address == "" {
		// init chain and init chain config
		cc := cmd.NewCloudConfig(chainConfig.Name, ".")
		if !cc.Exist() {
			err := cc.Init(chainConfig.Spec.Id)
			if err != nil {
				return err
			}
		}

		// check node account
		accountConfigMap := &corev1.ConfigMap{}
		err := r.Get(ctx, types.NamespacedName{Name: GetNodeAccountName(chainConfig.Name, chainNode.Name), Namespace: chainNode.Namespace}, accountConfigMap)
		if err != nil && errors.IsNotFound(err) {

			keyId, nodeAddress, err := cc.CreateAccount(chainNode.Spec.KmsPassword)
			if err != nil {
				return err
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
			kmsDb, err := cc.ReadKmsDb(nodeAddress)
			if err != nil {
				return err
			}
			accountConfigMap.BinaryData = map[string][]byte{
				"kms.db": kmsDb,
			}

			// set ownerReference
			_ = ctrl.SetControllerReference(chainNode, accountConfigMap, r.Scheme)
			// create chain secret
			err = r.Create(ctx, accountConfigMap)
			if err != nil {
				return err
			}
			chainNode.Spec.Address = nodeAddress
			// requeue
			//return ctrl.Result{Requeue: true}, nil

		} else if err != nil {
			logger.Error(err, "failed to get node address configmap")
			return err
		} else {
			chainNode.Spec.Address = accountConfigMap.Data["address"]
		}
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

//+kubebuilder:rbac:groups=citacloud.rivtower.com,resources=chainnodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=citacloud.rivtower.com,resources=chainnodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=citacloud.rivtower.com,resources=chainnodes/finalizers,verbs=update

func (r *ChainNodeReconciler) Reconcile1(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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
		return r.EnsureInit(ctx, logger, chainConfig, chainNode)
	} else if chainNode.Spec.Action == citacloudv1.NodeCreate {
		return r.EnsureCreate(ctx, logger, chainConfig, chainNode)
	} else {
		return ctrl.Result{}, fmt.Errorf("mismatched node action")
	}
}

func (r *ChainNodeReconciler) EnsureInit(ctx context.Context, logger logr.Logger,
	chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) (ctrl.Result, error) {

	// init chain and init chain config
	cc := cmd.NewCloudConfig(chainConfig.Name, ".")
	if !cc.Exist() {
		err := cc.Init(chainConfig.Spec.Id)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// check node account
	var nodeAddress string
	accountConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-%s-account", chainConfig.Name, chainNode.Name), Namespace: chainNode.Namespace}, accountConfigMap)
	if err != nil && errors.IsNotFound(err) {
		keyId, nodeAddress, err := cc.CreateAccount(chainNode.Spec.KmsPassword)
		if err != nil {
			return ctrl.Result{}, err
		}
		accountConfigMap.ObjectMeta = metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-account", chainConfig.Name, chainNode.Name),
			Namespace: chainNode.Namespace,
			Labels:    LabelsForChain(chainNode.Name),
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

	if chainNode.Spec.Address == "" {
		chainNode.Spec.Address = nodeAddress
		if err := r.Update(ctx, chainNode); err != nil {
			logger.Error(err, "update chain node address error")
			return ctrl.Result{}, err
		}
		logger.Info("set chain node address success")
		return ctrl.Result{Requeue: true}, nil
	}

	// set ownerReference
	if len(chainNode.OwnerReferences) == 0 {
		if err := ctrl.SetControllerReference(chainConfig, chainNode, r.Scheme); err != nil {
			logger.Error(err, "set chain node controller reference failed")
			return ctrl.Result{}, err
		}
		if err := r.Update(ctx, chainNode); err != nil {
			logger.Error(err, "update chain node controller reference failed")
			return ctrl.Result{}, err
		}
		logger.Info("set chain node controller reference success")
		return ctrl.Result{Requeue: true}, nil
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
	// set initialized status
	chainNode.Status.Status = citacloudv1.NodeInitialized
	err = r.Status().Update(ctx, chainNode)
	if err != nil {
		logger.Error(err, fmt.Sprintf("failed to update chain node status[%s]", citacloudv1.NodeInitialized))
		return ctrl.Result{}, err
	}
	logger.Info(fmt.Sprintf("update chain node status[%s] successful", citacloudv1.NodeInitialized))
	return ctrl.Result{}, nil
}

func (r *ChainNodeReconciler) EnsureCreate(ctx context.Context, logger logr.Logger,
	chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) (ctrl.Result, error) {
	// create clusterIP Service
	// check: apt-get update && apt-get install dnsutils && nslookup my-chainconfig-my-node-1-cluster-ip
	clusterIPService := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-%s-cluster-ip", chainConfig.Name, chainNode.Name), Namespace: chainNode.Namespace}, clusterIPService)
	if err != nil && errors.IsNotFound(err) {
		clusterIPService.ObjectMeta = metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-cluster-ip", chainConfig.Name, chainNode.Name),
			Namespace: chainNode.Namespace,
			Labels:    LabelsForChain(chainNode.Name),
		}
		TargetPort := intstr.FromInt(NetworkPort)
		clusterIPService.Spec = corev1.ServiceSpec{
			Selector: LabelsForChain(chainNode.Name),
			Ports: []corev1.ServicePort{
				{
					Name:       "network",
					Port:       40000,
					TargetPort: TargetPort,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		}
		// set ownerReference
		_ = ctrl.SetControllerReference(chainNode, clusterIPService, r.Scheme)
		// create chain secret
		err = r.Create(ctx, clusterIPService)
		if err != nil {
			logger.Error(err, "failed to create chain node clusterIP service")
			return ctrl.Result{}, err
		}
		// requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "failed to get chain node clusterIP service")
		return ctrl.Result{}, err
	}

	// create ChainNode's ConfigMap
	nodeConfigMap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-%s-config", chainConfig.Name, chainNode.Name), Namespace: chainNode.Namespace}, nodeConfigMap)
	if err != nil && errors.IsNotFound(err) {
		nodeConfigMap.ObjectMeta = metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-config", chainConfig.Name, chainNode.Name),
			Namespace: chainNode.Namespace,
			Labels:    LabelsForChain(chainNode.Name),
		}

		cnService := NewChainNodeService(chainConfig, chainNode)

		nodeConfigMap.Data = map[string]string{
			NodeConfigFile: cnService.GenerateNodeConfig(),
		}
		// set ownerReference
		_ = ctrl.SetControllerReference(chainNode, nodeConfigMap, r.Scheme)
		// create chain secret
		err = r.Create(ctx, nodeConfigMap)
		if err != nil {
			logger.Error(err, "failed to create chain node configmap")
			return ctrl.Result{}, err
		}
		// requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "failed to get chain node configmap")
		return ctrl.Result{}, err
	}

	logConfigMap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-%s-log-config", chainConfig.Name, chainNode.Name), Namespace: chainNode.Namespace}, logConfigMap)
	if err != nil && errors.IsNotFound(err) {
		logConfigMap.ObjectMeta = metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-log-config", chainConfig.Name, chainNode.Name),
			Namespace: chainNode.Namespace,
			Labels:    LabelsForChain(chainNode.Name),
		}

		cnService := NewChainNodeService(chainConfig, chainNode)
		logConfigMap.Data = map[string]string{
			ControllerLogConfigFile: cnService.GenerateControllerLogConfig(),
			ExecutorLogConfigFile:   cnService.GenerateExecutorLogConfig(),
			KmsLogConfigFile:        cnService.GenerateKmsLogConfig(),
			NetworkLogConfigFile:    cnService.GenerateNetworkLogConfig(),
			StorageLogConfigFile:    cnService.GenerateStorageLogConfig(),
		}
		// set ownerReference
		_ = ctrl.SetControllerReference(chainNode, logConfigMap, r.Scheme)
		// create chain secret
		err = r.Create(ctx, logConfigMap)
		if err != nil {
			logger.Error(err, "failed to create chain node log configmap")
			return ctrl.Result{}, err
		}
		// requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "failed to get chain node log configmap")
		return ctrl.Result{}, err
	}

	err = r.Get(ctx, types.NamespacedName{Name: chainNode.Name, Namespace: chainNode.Namespace}, &appsv1.StatefulSet{})
	if err != nil && errors.IsNotFound(err) {
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      chainNode.Name,
				Namespace: chainNode.Namespace,
				Labels:    LabelsForChain(chainNode.Name),
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: pointer.Int32(1),
				UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
					Type: appsv1.RollingUpdateStatefulSetStrategyType,
				},
				PodManagementPolicy: appsv1.ParallelPodManagement,
				Selector: &metav1.LabelSelector{
					MatchLabels: LabelsForChain(chainNode.Name),
				},
				VolumeClaimTemplates: generatePVC(chainNode),
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: LabelsForChain(chainNode.Name),
					},
					Spec: corev1.PodSpec{
						ShareProcessNamespace: pointer.Bool(true),
						InitContainers: []corev1.Container{
							{
								Name:            "init",
								Image:           "busybox:stable",
								ImagePullPolicy: "Always",
								Command:         []string{"/bin/sh"},
								Args:            []string{"-c", fmt.Sprintf("cp %s/kms.db %s", AccountVolumeMountPath, DataVolumeMountPath)},
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("10m"),
										corev1.ResourceMemory: resource.MustParse("10Mi"),
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      AccountVolumeName,
										MountPath: AccountVolumeMountPath,
									},
									{
										Name:      DataVolumeName,
										MountPath: DataVolumeMountPath,
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:            NetworkContainer,
								Image:           "citacloud/network_p2p:v6.3.0",
								ImagePullPolicy: "Always",
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: NetworkPort,
										Protocol:      "TCP",
										Name:          "network",
									},
									{
										ContainerPort: 50000,
										Protocol:      "TCP",
										Name:          "grpc",
									},
								},
								Command: []string{
									"sh",
									"-c",
									"network run -p 50000",
								},
								WorkingDir: DataVolumeMountPath,
								VolumeMounts: []corev1.VolumeMount{
									// data volume
									{
										Name:      DataVolumeName,
										MountPath: DataVolumeMountPath,
									},
									// node config
									{
										Name:      NodeConfigVolumeName,
										SubPath:   NodeConfigFile,
										MountPath: fmt.Sprintf("%s/%s", NodeConfigVolumeMountPath, NodeConfigFile),
									},
									// log config
									{
										Name:      LogConfigVolumeName,
										SubPath:   NetworkLogConfigFile,
										MountPath: fmt.Sprintf("%s/%s", NodeConfigVolumeMountPath, NetworkLogConfigFile),
									},
								},
							},
							{
								Name:            ConsensusContainer,
								Image:           "citacloud/consensus_raft:v6.3.0",
								ImagePullPolicy: "Always",
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: 50001,
										Protocol:      "TCP",
										Name:          "grpc",
									},
								},
								Command: []string{
									"sh",
									"-c",
									"consensus run --stdout",
								},
								WorkingDir: DataVolumeMountPath,
								VolumeMounts: []corev1.VolumeMount{
									// data volume
									{
										Name:      DataVolumeName,
										MountPath: DataVolumeMountPath,
									},
									// node config
									{
										Name:      NodeConfigVolumeName,
										SubPath:   NodeConfigFile,
										MountPath: fmt.Sprintf("%s/%s", NodeConfigVolumeMountPath, NodeConfigFile),
									},
								},
							},
							{
								Name:            ExecutorContainer,
								Image:           "citacloud/executor_evm:v6.3.0",
								ImagePullPolicy: "Always",
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: 50002,
										Protocol:      "TCP",
										Name:          "grpc",
									},
								},
								Command: []string{
									"sh",
									"-c",
									"executor run -p 50002",
								},
								WorkingDir: DataVolumeMountPath,
								VolumeMounts: []corev1.VolumeMount{
									// data volume
									{
										Name:      DataVolumeName,
										MountPath: DataVolumeMountPath,
									},
									// node config
									{
										Name:      NodeConfigVolumeName,
										SubPath:   NodeConfigFile,
										MountPath: fmt.Sprintf("%s/%s", NodeConfigVolumeMountPath, NodeConfigFile),
									},
									// log config
									{
										Name:      LogConfigVolumeName,
										SubPath:   ExecutorLogConfigFile,
										MountPath: fmt.Sprintf("%s/%s", NodeConfigVolumeMountPath, ExecutorLogConfigFile),
									},
								},
							},
							{
								Name:            StorageContainer,
								Image:           "citacloud/storage_rocksdb:v6.3.0",
								ImagePullPolicy: "Always",
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: 50003,
										Protocol:      "TCP",
										Name:          "grpc",
									},
								},
								Command: []string{
									"sh",
									"-c",
									"storage run -p 50003",
								},
								WorkingDir: DataVolumeMountPath,
								VolumeMounts: []corev1.VolumeMount{
									// data volume
									{
										Name:      DataVolumeName,
										MountPath: DataVolumeMountPath,
									},
									// node config
									{
										Name:      NodeConfigVolumeName,
										SubPath:   NodeConfigFile,
										MountPath: fmt.Sprintf("%s/%s", NodeConfigVolumeMountPath, NodeConfigFile),
									},
									// log config
									{
										Name:      LogConfigVolumeName,
										SubPath:   StorageLogConfigFile,
										MountPath: fmt.Sprintf("%s/%s", NodeConfigVolumeMountPath, StorageLogConfigFile),
									},
								},
							},
							{
								Name:            ControllerContainer,
								Image:           "citacloud/controller:v6.3.0",
								ImagePullPolicy: "Always",
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: 50004,
										Protocol:      "TCP",
										Name:          "grpc",
									},
								},
								Command: []string{
									"sh",
									"-c",
									"controller run -p 50004",
								},
								WorkingDir: DataVolumeMountPath,
								VolumeMounts: []corev1.VolumeMount{
									// data volume
									{
										Name:      DataVolumeName,
										MountPath: DataVolumeMountPath,
									},
									// node config
									{
										Name:      NodeConfigVolumeName,
										SubPath:   NodeConfigFile,
										MountPath: fmt.Sprintf("%s/%s", NodeConfigVolumeMountPath, NodeConfigFile),
									},
									// log config
									{
										Name:      LogConfigVolumeName,
										SubPath:   ControllerLogConfigFile,
										MountPath: fmt.Sprintf("%s/%s", NodeConfigVolumeMountPath, ControllerLogConfigFile),
									},
								},
							},
							{
								Name:            KmsContainer,
								Image:           "citacloud/kms_sm:v6.3.0",
								ImagePullPolicy: "Always",
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: 50005,
										Protocol:      "TCP",
										Name:          "grpc",
									},
								},
								Command: []string{
									"sh",
									"-c",
									"kms run -p 50005",
								},
								WorkingDir: DataVolumeMountPath,
								VolumeMounts: []corev1.VolumeMount{
									// data volume
									{
										Name:      DataVolumeName,
										MountPath: DataVolumeMountPath,
									},
									// node config
									{
										Name:      NodeConfigVolumeName,
										SubPath:   NodeConfigFile,
										MountPath: fmt.Sprintf("%s/%s", NodeConfigVolumeMountPath, NodeConfigFile),
									},
									// log config
									{
										Name:      LogConfigVolumeName,
										SubPath:   KmsLogConfigFile,
										MountPath: fmt.Sprintf("%s/%s", NodeConfigVolumeMountPath, KmsLogConfigFile),
									},
								},
							},
						},
						Volumes: getVolumes(chainNode),
					},
				},
			},
		}
		// set ownerReference
		_ = ctrl.SetControllerReference(chainNode, sts, r.Scheme)
		//  set pvc reference to crd
		for i := range sts.Spec.VolumeClaimTemplates {
			_ = ctrl.SetControllerReference(chainNode, &sts.Spec.VolumeClaimTemplates[i], r.Scheme)
		}
		// create chain secret
		err = r.Create(ctx, sts)
		if err != nil {
			logger.Error(err, "failed to create chain node statefulset")
			return ctrl.Result{}, err
		}
		// requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "failed to get chain node stateful")
		return ctrl.Result{}, err
	}

	// set running status
	chainNode.Status.Status = citacloudv1.NodeRunning
	err = r.Status().Update(ctx, chainNode)
	if err != nil {
		logger.Error(err, fmt.Sprintf("failed to update chain node status[%s]", citacloudv1.NodeRunning))
		return ctrl.Result{}, err
	}
	logger.Info(fmt.Sprintf("update chain node status[%s] successful", citacloudv1.NodeRunning))

	return ctrl.Result{}, nil
}

func generatePVC(chainNode *citacloudv1.ChainNode) []corev1.PersistentVolumeClaim {
	return []corev1.PersistentVolumeClaim{
		// data pvc
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DataVolumeName,
				Namespace: chainNode.Namespace,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: *resource.NewQuantity(*chainNode.Spec.StorageSize, resource.BinarySI),
					},
				},
				StorageClassName: chainNode.Spec.StorageClassName,
			},
		},
	}
}

func getVolumes(chainNode *citacloudv1.ChainNode) []corev1.Volume {
	return []corev1.Volume{
		// account configmap as a volume
		{
			Name: AccountVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-%s-account", chainNode.Spec.ChainName, chainNode.Name),
					},
				},
			},
		},
		// node config configmap as a volume
		{
			Name: NodeConfigVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-%s-config", chainNode.Spec.ChainName, chainNode.Name),
					},
				},
			},
		},
		// log config configmap as a volume
		{
			Name: LogConfigVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-%s-log-config", chainNode.Spec.ChainName, chainNode.Name),
					},
				},
			},
		},
	}
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
