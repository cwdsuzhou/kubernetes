/*
Copyright 2015 The Kubernetes Authors.

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

package priorities

import (
	"fmt"
	"sync"
	"sync/atomic"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

// CalculateNodeAffinityPriorityMap prioritizes nodes according to node affinity scheduling preferences
// indicated in PreferredDuringSchedulingIgnoredDuringExecution. Each time a node matches a preferredSchedulingTerm,
// it will get an add of preferredSchedulingTerm.Weight. Thus, the more preferredSchedulingTerms
// the node satisfies and the more the preferredSchedulingTerm that is satisfied weights, the higher
// score the node gets.
func CalculateNodeAffinityPriorityMap(pod *v1.Pod, meta interface{}, nodeInfo *schedulernodeinfo.NodeInfo) (schedulerapi.HostPriority, error) {
	node := nodeInfo.Node()
	if node == nil {
		return schedulerapi.HostPriority{}, fmt.Errorf("node not found")
	}

	// default is the podspec.
	affinity := pod.Spec.Affinity
	if priorityMeta, ok := meta.(*priorityMetadata); ok {
		// We were able to parse metadata, use affinity from there.
		affinity = priorityMeta.affinity
	}

	var count = new(int32)
	var selectError error
	var wg sync.WaitGroup

	// A nil element of PreferredDuringSchedulingIgnoredDuringExecution matches no objects.
	// An element of PreferredDuringSchedulingIgnoredDuringExecution that refers to an
	// empty PreferredSchedulingTerm matches all objects.
	if affinity != nil && affinity.NodeAffinity != nil && affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution != nil {
		// Match PreferredDuringSchedulingIgnoredDuringExecution term by term.
		for i := range affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			preferredSchedulingTerm := &affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[i]
			if preferredSchedulingTerm.Weight == 0 {
				continue
			}

			if selectError != nil {
				return schedulerapi.HostPriority{}, selectError
			}
			wg.Add(1)
			go func() {
				wg.Done()
				if selectError != nil {
					return
				}
				// TODO: Avoid computing it for all nodes if this becomes a performance problem.
				nodeSelector, err := v1helper.NodeSelectorRequirementsAsSelector(preferredSchedulingTerm.Preference.MatchExpressions)
				// Not need to add lock since any error happen this func should return
				if err != nil {
					selectError = err
				}
				if nodeSelector.Matches(labels.Set(node.Labels)) {
					atomic.AddInt32(count, preferredSchedulingTerm.Weight)
				}
				return
			}()
		}
	}

	wg.Wait()

	if selectError != nil {
		return schedulerapi.HostPriority{}, selectError
	}

	return schedulerapi.HostPriority{
		Host:  node.Name,
		Score: int(*count),
	}, nil
}

// CalculateNodeAffinityPriorityReduce is a reduce function for node affinity priority calculation.
var CalculateNodeAffinityPriorityReduce = NormalizeReduce(schedulerapi.MaxPriority, false)
