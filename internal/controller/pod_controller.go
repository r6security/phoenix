/*
 * Copyright (C) 2023 R6 Security, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the Server Side Public License, version 1,
 * as published by MongoDB, Inc.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * Server Side Public License for more details.
 *
 * You should have received a copy of the Server Side Public License
 * along with this program. If not, see
 * <http://www.mongodb.com/licensing/server-side-public-license>.
 */

package controller

import (
	"context"
	"reflect"
	"strings"

	//corev1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//log := log.FromContext(ctx)
	/*
		// TODO(user): your logic here
		// Start by declaring the custom resource to be type "Pod"
		pod := &corev1.Pod{}

		// Then retrieve from the cluster the resource that triggered this reconciliation.
		// Store these contents into an object used throughout reconciliation.
		err := r.Client.Get(context.Background(), req.NamespacedName, pod)
		if err != nil {
			if errors.IsNotFound(err) {
				// If the resource is not found, that is OK. It just means the desired state is to
				// not have any resources for this Pod but no delete is required.
				log.Info(fmt.Sprintf(`Pod was deleted for Pod "%s" does not exist, but that's OK, no action is required`, req.Name))
				return ctrl.Result{}, nil
			} else {
				// some other error happend
				log.Error(err, fmt.Sprintf(`Failed to retrieve pod resource "%s": %s`, req.Name, err.Error()))
				return ctrl.Result{}, err
			}
		}

		if pod.ObjectMeta.Annotations == nil {
			// here no scheduled return is needed, since anything modifies a pod (with annotations) Reconcile will catch it
			return ctrl.Result{}, nil
		}

		// Does the pod have SecurityEvent on it?
		if securityEvent, found := pod.ObjectMeta.Annotations[AMTD_SECURITY_EVENT]; found {

			// Filter annotations to find to best action for the security event based on strategies
			action := getAction(pod.ObjectMeta.Annotations, securityEvent)

			switch action {
			case "destroy":
				// Try to delete pod
				podErr := r.Client.Delete(ctx, pod)

				// Success for this delete is either:
				// 1. the delete is successful without error
				// 2. the resource already doesn't exist so delete can't take action
				if err != nil && !errors.IsNotFound(err) {
					// If any other error occurs, log it
					log.Error(podErr, fmt.Sprintf(`Failed to delete pod "%s"`, req.Name))
				}
				log.Info(fmt.Sprintf(`Pod: "%s" was sucessfully deleted with ACTION: delete`, pod.Name))
			default:
				log.Info(fmt.Sprintf(`ACTION: %s -> POD: %s - NOT IMPLEMENTED YET`, action, pod.Name))
			}
		}

		// here no scheduled return is needed, since anything modifies a pod (with annotations) Reconcile will catch it
		// return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		// TODO: shall we schedule it again if something went wrong? (Should it be handled above?)
	*/
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}

func getAction(podAnnotations map[string]string, securityEvent string) string {
	// Check for the proper strategy to match and action to execute
	securityEventMap := stringToMap(securityEvent)

	action := ""
	for key, value := range podAnnotations {
		// strategy annotation is found
		if strings.HasPrefix(key, AMTD_STRATEGY_BASE) {

			// prepare strategy and action for comparison with SecurityEvent
			strategyMap := stringToMap(value)
			actionCandidate := strategyMap["action"]
			// remove action to be easily comparable with SecurityEvent (it has no "action" key)
			delete(strategyMap, "action")

			// check for a custom strategy (e.g. not default type and action)
			if reflect.DeepEqual(securityEventMap, strategyMap) {
				action = actionCandidate
				return action
			}

			// check for a default strategy (type=...,action=...)
			if strategyMap["type"] == "default" {
				action = actionCandidate
			}
		}
	}

	return action
}

func isSecurityEventMatchStrategy(securityEvent string, strategy string) (bool, string) {
	securityEventMap := stringToMap(securityEvent)
	strategyMap := stringToMap(strategy)

	if reflect.DeepEqual(securityEventMap, strategyMap) {
		return true, strategyMap["action"]
	}

	return false, ""
}

func stringToMap(text string) map[string]string {
	entries := strings.Split(text, ",")

	m := make(map[string]string)
	for _, e := range entries {
		parts := strings.Split(e, "=")
		m[parts[0]] = parts[1]
	}

	return m
}
