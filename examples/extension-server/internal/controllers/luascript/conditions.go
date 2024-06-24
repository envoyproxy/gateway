package luascript_controller

import (
	"time"

	v1 "github.com/exampleorg/envoygateway-extension/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getNewLuaScriptStatus(oldStatus v1.GlobalLuaScriptStatus, newStatus v1.GlobalLuaScriptStatus, observedGeneration int64) (status v1.GlobalLuaScriptStatus, needsUpdate bool) {
	ret := v1.GlobalLuaScriptStatus{}

	oldConditionsByType := make(map[string]metav1.Condition)
	for _, condition := range oldStatus.Conditions {
		oldConditionsByType[condition.Type] = condition
	}

	newConditionsByType := make(map[string]metav1.Condition)
	newConditions := make([]metav1.Condition, 0)

	for _, condition := range newStatus.Conditions {
		newCondition, needsUpdate := updateConditionIfDifferent(oldConditionsByType, condition)
		if !needsUpdate {
			return ret, false
		}
		newCondition.ObservedGeneration = observedGeneration

		if _, ok := newConditionsByType[newCondition.Type]; ok {
			continue
		}

		newConditionsByType[newCondition.Type] = newCondition
		newConditions = append(newConditions, newCondition)
	}

	ret.Conditions = newConditions
	return ret, true
}

func updateConditionIfDifferent(oldConditionsByType map[string]metav1.Condition, newCondition metav1.Condition) (condition metav1.Condition, needsUpdate bool) {
	if newCondition.LastTransitionTime.Time.IsZero() {
		newCondition.LastTransitionTime = metav1.NewTime(time.Now().UTC())
	}

	oldCondition, ok := oldConditionsByType[newCondition.Type]
	if !ok || newCondition.Message != oldCondition.Message ||
		newCondition.Reason != oldCondition.Reason ||
		newCondition.Status != oldCondition.Status {
		return newCondition, true
	}

	return oldCondition, false
}
