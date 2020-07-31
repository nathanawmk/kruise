/*
Copyright 2020 The Kruise Authors.

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

package core

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/appscode/jsonpatch"
	appsv1alpha1 "github.com/openkruise/kruise/apis/apps/v1alpha1"
	clonesetutils "github.com/openkruise/kruise/pkg/controller/cloneset/utils"
	"github.com/openkruise/kruise/pkg/util/inplaceupdate"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubecontroller "k8s.io/kubernetes/pkg/controller"
)

var (
	inPlaceUpdateTemplateSpecPatchRexp = regexp.MustCompile("^/containers/([0-9]+)/image$")
)

type commonControl struct {
	*appsv1alpha1.CloneSet
}

var _ Control = &commonControl{}

func (c *commonControl) IsInitializing() bool {
	return false
}

func (c *commonControl) SetRevisionTemplate(revisionSpec map[string]interface{}, template map[string]interface{}) {
	revisionSpec["template"] = template
	template["$patch"] = "replace"
}

func (c *commonControl) ApplyRevisionPatch(patched []byte) (*appsv1alpha1.CloneSet, error) {
	restoredSet := &appsv1alpha1.CloneSet{}
	if err := json.Unmarshal(patched, restoredSet); err != nil {
		return nil, err
	}
	return restoredSet, nil
}

func (c *commonControl) IsReadyToScale() bool {
	return true
}

func (c *commonControl) NewVersionedPods(currentCS, updateCS *appsv1alpha1.CloneSet,
	currentRevision, updateRevision string,
	expectedCreations, expectedCurrentCreations int,
	availableIDs []string,
) ([]*v1.Pod, error) {
	var newPods []*v1.Pod
	if expectedCreations <= expectedCurrentCreations {
		newPods = c.newVersionedPods(currentCS, currentRevision, expectedCreations, &availableIDs)
	} else {
		newPods = c.newVersionedPods(currentCS, currentRevision, expectedCurrentCreations, &availableIDs)
		newPods = append(newPods, c.newVersionedPods(updateCS, updateRevision, expectedCreations-expectedCurrentCreations, &availableIDs)...)
	}
	return newPods, nil
}

func (c *commonControl) newVersionedPods(cs *appsv1alpha1.CloneSet, revision string, replicas int, availableIDs *[]string) []*v1.Pod {
	var newPods []*v1.Pod
	for i := 0; i < replicas; i++ {
		if len(*availableIDs) == 0 {
			return newPods
		}
		id := (*availableIDs)[0]
		*availableIDs = (*availableIDs)[1:]

		pod, _ := kubecontroller.GetPodFromTemplate(&cs.Spec.Template, cs, metav1.NewControllerRef(cs, clonesetutils.ControllerKind))
		if pod.Labels == nil {
			pod.Labels = make(map[string]string)
		}
		pod.Labels[apps.ControllerRevisionHashLabelKey] = revision

		pod.Name = fmt.Sprintf("%s-%s", cs.Name, id)
		pod.Namespace = cs.Namespace
		pod.Labels[appsv1alpha1.CloneSetInstanceID] = id

		inplaceupdate.InjectReadinessGate(pod)
		clonesetutils.UpdateStorage(cs, pod)

		newPods = append(newPods, pod)
	}
	return newPods
}

func (c *commonControl) IsPodUpdatePaused(pod *v1.Pod) bool {
	return false
}

func (c *commonControl) IsPodUpdateReady(pod *v1.Pod, minReadySeconds int32) bool {
	if !clonesetutils.IsRunningAndAvailable(pod, minReadySeconds) {
		return false
	}
	condition := inplaceupdate.GetCondition(pod)
	if condition != nil && condition.Status != v1.ConditionTrue {
		return false
	}
	return true
}

func (c *commonControl) GetPodsSortFunc(pods []*v1.Pod, waitUpdateIndexes []int) func(i, j int) bool {
	// not-ready < ready, unscheduled < scheduled, and pending < running
	return func(i, j int) bool {
		return kubecontroller.ActivePods(pods).Less(waitUpdateIndexes[i], waitUpdateIndexes[j])
	}
}

func (c *commonControl) GetUpdateOptions() *inplaceupdate.UpdateOptions {
	opts := &inplaceupdate.UpdateOptions{}
	if c.Spec.UpdateStrategy.InPlaceUpdateStrategy != nil {
		opts.GracePeriodSeconds = c.Spec.UpdateStrategy.InPlaceUpdateStrategy.GracePeriodSeconds
	}
	return opts
}

func (c *commonControl) ValidateCloneSetUpdate(oldCS, newCS *appsv1alpha1.CloneSet) error {
	if newCS.Spec.UpdateStrategy.Type != appsv1alpha1.InPlaceOnlyCloneSetUpdateStrategyType {
		return nil
	}

	oldTempJSON, _ := json.Marshal(oldCS.Spec.Template.Spec)
	newTempJSON, _ := json.Marshal(newCS.Spec.Template.Spec)
	patches, err := jsonpatch.CreatePatch(oldTempJSON, newTempJSON)
	if err != nil {
		return fmt.Errorf("failed calculate patches between old/new template spec")
	}

	for _, p := range patches {
		if p.Operation != "replace" || !inPlaceUpdateTemplateSpecPatchRexp.MatchString(p.Path) {
			return fmt.Errorf("only allowed to update images in spec for %s, but found %s %s",
				appsv1alpha1.InPlaceOnlyCloneSetUpdateStrategyType, p.Operation, p.Path)
		}
	}
	return nil
}
