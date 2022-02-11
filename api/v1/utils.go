package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func GetNewTime(time time.Time) *metav1.Time {
	t := metav1.NewTime(time)
	return &t
}
