package v1

import (
	"sort"
	"testing"
	"time"
)

func TestSort(t *testing.T) {
	v := []ValidatorAccountInfo{
		{
			Address:           "3",
			CreationTimestamp: GetNewTime(time.Now()),
		},
		{
			Address:           "1",
			CreationTimestamp: GetNewTime(time.Now().Add(-time.Second * 10)),
		},
		{
			Address:           "2",
			CreationTimestamp: GetNewTime(time.Now()),
		},
	}
	sort.Sort(ByCreationTimestampForValidatorAccount(v))
	t.Log(v)
}
