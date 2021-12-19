package controllers

import (
	"context"

	citacloudv1 "github.com/cita-cloud/cita-cloud-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ChainNodeReconciler) ReconcileService(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) error {
	logger := log.FromContext(ctx)
	old := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: GetClusterIPName(chainNode.Name, chainNode.Spec.ChainName), Namespace: chainNode.Namespace}, old)
	if errors.IsNotFound(err) {
		newObj := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GetClusterIPName(chainNode.Name, chainNode.Spec.ChainName),
				Namespace: chainNode.Namespace,
			},
		}
		if err = r.updateService(ctx, chainConfig, chainNode, newObj); err != nil {
			return err
		}
		logger.Info("create node service....")
		return r.Create(ctx, newObj)
	} else if err != nil {
		return err
	}

	logger.Info("service update is currently not supported, go pass")
	return nil
	//cur := old.DeepCopy()
	//if err := r.updateService(ctx, chainConfig, chainNode, cur); err != nil {
	//	return err
	//}
	//if IsEqual(old, cur) {
	//	return nil
	//}
	//
	//logger.Info("update node service...")
	//return r.Update(ctx, cur)
}

func (r *ChainNodeReconciler) updateService(ctx context.Context, chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode, service *corev1.Service) error {
	labels := MergeLabels(service.Labels, LabelsForNode( chainNode.Spec.ChainName, chainNode.Name))
	logger := log.FromContext(ctx)
	service.Labels = labels
	if err := ctrl.SetControllerReference(chainNode, service, r.Scheme); err != nil {
		logger.Error(err, "node service SetControllerReference error")
		return err
	}

	service.Spec = corev1.ServiceSpec{
		Selector: labels,
		Ports: []corev1.ServicePort{
			{
				Name:       "network",
				Port:       40000,
				TargetPort: intstr.FromInt(NetworkPort),
			},
		},
		Type: corev1.ServiceTypeClusterIP,
	}

	return nil
}
