package elasticquota

import (
	"github.com/nebuly-ai/nebulnetes/pkg/constant"
	"github.com/nebuly-ai/nebulnetes/pkg/gpu"
	"github.com/nebuly-ai/nebulnetes/pkg/test/factory"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestElasticQuotaPodsReconciler_sortPodListForFindingOverQuotaPods(t *testing.T) {
	tests := []struct {
		name                   string
		podList                v1.PodList
		expectedSortedPodNames []string
	}{
		{
			name:                   "Empty list",
			podList:                v1.PodList{},
			expectedSortedPodNames: []string{},
		},
		{
			name: "Sorted by ascending creation timestamp",
			podList: v1.PodList{Items: []v1.Pod{
				factory.BuildPod("ns-1", "pd-1").
					WithCreationTimestamp(metav1.NewTime(time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC))).
					WithPriority(int32(100)).
					Get(),
				factory.BuildPod("ns-1", "pd-2").
					WithCreationTimestamp(metav1.NewTime(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))).
					WithPriority(int32(10)).
					Get(),
				factory.BuildPod("ns-1", "pd-3").
					WithCreationTimestamp(metav1.NewTime(time.Date(1996, 1, 1, 0, 0, 0, 0, time.UTC))).
					WithPriority(int32(50)).
					Get(),
			}},
			expectedSortedPodNames: []string{"pd-1", "pd-3", "pd-2"},
		},
		{
			name: "Sorted by priority if creation timestamp is equal",
			podList: v1.PodList{Items: []v1.Pod{
				factory.BuildPod("ns-1", "pd-1").
					WithCreationTimestamp(metav1.NewTime(time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC))).
					WithPriority(int32(100)).
					Get(),
				factory.BuildPod("ns-1", "pd-2").
					WithCreationTimestamp(metav1.NewTime(time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC))).
					WithPriority(int32(10)).
					Get(),
				factory.BuildPod("ns-1", "pd-3").
					WithCreationTimestamp(metav1.NewTime(time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC))).
					WithPriority(int32(50)).
					Get(),
			}},
			expectedSortedPodNames: []string{"pd-2", "pd-3", "pd-1"},
		},
		{
			name: "Sorted by request resources if same priority",
			podList: v1.PodList{Items: []v1.Pod{
				factory.BuildPod("ns-1", "pd-1").
					WithCreationTimestamp(metav1.NewTime(time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC))).
					WithPriority(int32(10)).
					WithContainer(factory.BuildContainer("c1", "test").WithNvidiaGPURequest(10).Get()).
					Get(),
				factory.BuildPod("ns-1", "pd-2").
					WithCreationTimestamp(metav1.NewTime(time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC))).
					WithPriority(int32(10)).
					WithContainer(factory.BuildContainer("c1", "test").WithNvidiaGPURequest(1).Get()).
					Get(),
				factory.BuildPod("ns-1", "pd-3").
					WithCreationTimestamp(metav1.NewTime(time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC))).
					WithPriority(int32(10)).
					WithContainer(factory.BuildContainer("c1", "test").WithNvidiaGPURequest(5).Get()).
					Get(),
			}},
			expectedSortedPodNames: []string{"pd-2", "pd-3", "pd-1"},
		},
		{
			name: "Sorted alphabetically as last resort",
			podList: v1.PodList{Items: []v1.Pod{
				factory.BuildPod("ns-1", "pd-1").
					WithCreationTimestamp(metav1.NewTime(time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC))).
					WithPriority(int32(10)).
					Get(),
				factory.BuildPod("ns-1", "pd-3").
					WithCreationTimestamp(metav1.NewTime(time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC))).
					WithPriority(int32(10)).
					Get(),
				factory.BuildPod("ns-1", "pd-2").
					WithCreationTimestamp(metav1.NewTime(time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC))).
					WithPriority(int32(10)).
					Get(),
			}},
			expectedSortedPodNames: []string{"pd-1", "pd-2", "pd-3"},
		},
	}

	r := elasticQuotaPodsReconciler{
		resourceCalculator: &gpu.Calculator{
			NvidiaGPUDeviceMemoryGB: constant.DefaultNvidiaGPUResourceMemory,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.sortPodListForFindingOverQuotaPods(tt.podList.Items)
			podNames := make([]string, len(tt.podList.Items))
			for i, pod := range tt.podList.Items {
				podNames[i] = pod.Name
			}
			assert.Equal(t, tt.expectedSortedPodNames, podNames)
		})
	}
}
