package controllers

import (
	"context"
	"fmt"

	citacloudv1 "github.com/cita-cloud/cita-cloud-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ReconcileStatefulSet if statefulset update, then should return true for sync status
func (r *ChainNodeReconciler) ReconcileStatefulSet(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) (bool, error) {
	logger := log.FromContext(ctx)
	old := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: chainNode.Name, Namespace: chainNode.Namespace}, old)
	if errors.IsNotFound(err) {
		newObj := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      chainNode.Name,
				Namespace: chainNode.Namespace,
			},
		}
		if err = r.generateStatefulSet(ctx, chainConfig, chainNode, newObj); err != nil {
			return false, err
		}
		logger.Info("create node statefulset....")
		return false, r.Create(ctx, newObj)
	} else if err != nil {
		return false, err
	}

	cur := old.DeepCopy()
	if err := r.generateStatefulSet(ctx, chainConfig, chainNode, cur); err != nil {
		return false, err
	}

	t1Copy := old.Spec.Template.Spec.Containers
	t2Copy := cur.Spec.Template.Spec.Containers
	if IsEqual(t1Copy, t2Copy) && *old.Spec.Replicas == *cur.Spec.Replicas {
		logger.Info("the statefulset part has not changed, go pass")
		return false, nil
	}
	// currently only update the changes under Containers
	old.Spec.Template.Spec.Containers = cur.Spec.Template.Spec.Containers
	old.Spec.Replicas = cur.Spec.Replicas
	logger.Info("update node statefulset...")
	return true, r.Update(ctx, old)
}

func (r *ChainNodeReconciler) generateStatefulSet(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode, set *appsv1.StatefulSet) error {
	replica := int32(1)
	if chainNode.Spec.Action == citacloudv1.NodeStop {
		replica = 0
	}

	labels := MergeLabels(set.Labels, LabelsForNode(chainNode.Spec.ChainName, chainNode.Name))
	logger := log.FromContext(ctx)
	set.Labels = labels
	if err := ctrl.SetControllerReference(chainNode, set, r.Scheme); err != nil {
		logger.Error(err, "node statefulset SetControllerReference error")
		return err
	}
	var networkCmdStr string
	if chainConfig.Spec.EnableTLS {
		networkCmdStr = "network run --stdout"
	} else {
		networkCmdStr = "network run -p 50000"
	}

	set.Spec = appsv1.StatefulSetSpec{
		Replicas: pointer.Int32(replica),
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		},
		PodManagementPolicy: appsv1.ParallelPodManagement,
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		VolumeClaimTemplates: GeneratePVC(chainNode),
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				ShareProcessNamespace: pointer.Bool(true),
				InitContainers: []corev1.Container{
					{
						Name:            "init",
						Image:           "busybox:stable",
						ImagePullPolicy: chainNode.Spec.PullPolicy,
						Command:         []string{"/bin/sh"},
						Args:            []string{"-c", fmt.Sprintf("if [ ! -f \"%s/kms.db\" ]; then cp %s/kms.db %s;fi;", DataVolumeMountPath, AccountVolumeMountPath, DataVolumeMountPath)},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("10m"),
								corev1.ResourceMemory: resource.MustParse("10Mi"),
							},
						},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: corev1.TerminationMessageReadFile,
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
						Image:           chainNode.Spec.NetworkImage,
						ImagePullPolicy: chainNode.Spec.PullPolicy,
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: NetworkPort,
								Protocol:      corev1.ProtocolTCP,
								Name:          "network",
							},
							{
								ContainerPort: 50000,
								Protocol:      corev1.ProtocolTCP,
								Name:          "grpc",
							},
						},
						Command: []string{
							"sh",
							"-c",
							networkCmdStr,
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
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					},
					{
						Name:            ConsensusContainer,
						Image:           chainNode.Spec.ConsensusImage,
						ImagePullPolicy: chainNode.Spec.PullPolicy,
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 50001,
								Protocol:      corev1.ProtocolTCP,
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
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					},
					{
						Name:            ExecutorContainer,
						Image:           chainNode.Spec.ExecutorImage,
						ImagePullPolicy: chainNode.Spec.PullPolicy,
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 50002,
								Protocol:      corev1.ProtocolTCP,
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
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					},
					{
						Name:            StorageContainer,
						Image:           chainNode.Spec.StorageImage,
						ImagePullPolicy: chainNode.Spec.PullPolicy,
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 50003,
								Protocol:      corev1.ProtocolTCP,
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
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					},
					{
						Name:            ControllerContainer,
						Image:           chainNode.Spec.ControllerImage,
						ImagePullPolicy: chainNode.Spec.PullPolicy,
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 50004,
								Protocol:      corev1.ProtocolTCP,
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
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					},
					{
						Name:            KmsContainer,
						Image:           chainNode.Spec.KmsImage,
						ImagePullPolicy: chainNode.Spec.PullPolicy,
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 50005,
								Protocol:      corev1.ProtocolTCP,
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
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					},
				},
				Volumes: GetVolumes(chainNode),
			},
		},
	}

	//  set pvc reference to crd
	for i := range set.Spec.VolumeClaimTemplates {
		if err := ctrl.SetControllerReference(chainNode, &set.Spec.VolumeClaimTemplates[i], r.Scheme); err != nil {
			logger.Error(err, "node statefulset pvc SetControllerReference error")
			return err
		}
	}

	return nil
}

func GeneratePVC(chainNode *citacloudv1.ChainNode) []corev1.PersistentVolumeClaim {
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

func GetVolumes(chainNode *citacloudv1.ChainNode) []corev1.Volume {
	return []corev1.Volume{
		// account configmap as a volume
		{
			Name: AccountVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: GetAccountConfigmap(chainNode.Spec.ChainName, chainNode.Spec.Account),
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
						Name: GetNodeConfigName(chainNode.Name),
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
						Name: GetLogConfigName(chainNode.Name),
					},
				},
			},
		},
	}
}
