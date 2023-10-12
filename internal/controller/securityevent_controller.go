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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	amtdv1beta1 "github.com/r6security/phoenix/api/v1beta1"
)

// SecurityEventReconciler reconciles a SecurityEvent object
type SecurityEventReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=amtd.r6security.com,resources=securityevents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=amtd.r6security.com,resources=securityevents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=amtd.r6security.com,resources=securityevents/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecurityEvent object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *SecurityEventReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	securityEvent := &amtdv1beta1.SecurityEvent{}

	// Retrieve SecurityEvent that triggered reconciliation from the cluster
	err := r.Client.Get(context.Background(), req.NamespacedName, securityEvent)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf(`SecurityEvent "%s" removed, no action is needed`, req.Name))
			return ctrl.Result{}, nil
		} else {
			log.Error(err, fmt.Sprintf(`Failed to retrieve Security Event "%s": %s`, req.Name, err.Error()))
			return ctrl.Result{}, err
		}
	}

	log.Info(fmt.Sprintf(`SecurityEvent found: "%s, targets: %s"`, securityEvent.Name, securityEvent.Spec.Targets))

	// ---------------------------------------------------
	// Process pods in the target list of the SecurityEvent
	// ---------------------------------------------------
	var AMTD *amtdv1beta1.AdaptiveMovingTargetDefense
	for _, target := range securityEvent.Spec.Targets {
		namespace := strings.Split(target, "/")[0]
		name := strings.Split(target, "/")[1]

		// ---------------------------------------------------
		// Check that the resource exists and whether to deal with it
		// ---------------------------------------------------
		pod := &corev1.Pod{}
		namespacedName := types.NamespacedName{Namespace: namespace, Name: name}
		err = r.Client.Get(context.Background(), namespacedName, pod)

		if err != nil {
			if errors.IsNotFound(err) {
				log.Info(fmt.Sprintf(`Pod "%s" does not exist`, pod.Name))
				return ctrl.Result{}, nil
			} else {
				// some other error happend
				log.Error(err, fmt.Sprintf(`Failed to retrieve pod "%s"`, pod.Name))
				return ctrl.Result{}, err
			}
		}

		// Does the pod have annotation AMTD_MANAGED_BY?
		if pod.ObjectMeta.Annotations == nil {
			log.Error(err, fmt.Sprintf(`Pod "%s" is a SecurityEvent target but not AMTD managed`, pod.Name))
			return ctrl.Result{}, err
		} else if _, found := pod.ObjectMeta.Annotations[AMTD_MANAGED_BY]; !found {
			log.Error(err, fmt.Sprintf(`Pod "%s" is a SecurityEvent target but not AMTD managed`, pod.Name))
			return ctrl.Result{}, err
		}

		// ---------------------------------------------------
		// Look for the proper action for the SecurityEvent in AMTDs that manage the pod
		// ---------------------------------------------------
		action := ""
		var AMTDManageInfoList []AMTDManageInfo
		json.Unmarshal([]byte(pod.ObjectMeta.Annotations[AMTD_MANAGED_BY]), &AMTDManageInfoList)
		for _, AMTDManageInfo := range AMTDManageInfoList {

			// Get AMTD resource that is in AMTDManageInfo so it belongs to the pod
			AMTD = &amtdv1beta1.AdaptiveMovingTargetDefense{}
			err := r.Client.Get(context.Background(), types.NamespacedName{Namespace: AMTDManageInfo.AMTDNamespace, Name: AMTDManageInfo.AMTDName}, AMTD)
			if err != nil {
				if errors.IsNotFound(err) {
					log.Info(fmt.Sprintf(`AdaptiveMovingTargetDefense "%s" does not exist but found in pod annotation`, AMTDManageInfo.AMTDNamespace+"/"+AMTDManageInfo.AMTDName))
					break
				} else {
					log.Error(err, fmt.Sprintf(`Failed to retrieve AdaptiveMovingTargetDefense "%s": %s`, AMTDManageInfo.AMTDName, err.Error()))
					return ctrl.Result{}, err
				}
			}

			// Check whether there is a specific action to the security event
			for _, strategy := range AMTD.Spec.Strategy {
				if reflect.DeepEqual(strategy.Rule, securityEvent.Spec.Rule) {
					// we found the matching strategy no need to look further
					action = strategy.Action
					break
				}
			}
			// we found the matching strategy no need to look further
			if action != "" {
				break
			}
		}

		// ---------------------------------------------------
		// Add SecurityEvent spec to the annotation of the pod
		// ---------------------------------------------------
		var appliedSecurityEvents []amtdv1beta1.SecurityEvent
		if _, found := pod.ObjectMeta.Annotations[AMTD_APPLIED_SECURITY_EVENTS]; !found {
			appliedSecurityEvents = append(appliedSecurityEvents, *securityEvent)
		} else {
			// Already AMTD member
			json.Unmarshal([]byte(pod.ObjectMeta.Annotations[AMTD_APPLIED_SECURITY_EVENTS]), &appliedSecurityEvents)

			securityEventExist := false
			for _, appliedSecurityEvent := range appliedSecurityEvents {
				if appliedSecurityEvent.Name == securityEvent.Name {
					securityEventExist = true
					break
				}
			}

			if !securityEventExist {
				appliedSecurityEvents = append(appliedSecurityEvents, *securityEvent)
			} else {
				log.Info(fmt.Sprintf(`This SecurityEvent ("%s") was already processed - ignore it`, securityEvent.Name))
				//return ctrl.Result{}, nil
			}
		}

		if appliedSecurityEvents != nil {
			appliedSecurityEventsEncoded, err := json.Marshal(appliedSecurityEvents)
			if err != nil {
				log.Error(err, fmt.Sprintf(`appliedSecurityEvents json encoding does not work: %s`, err.Error()))
			}
			pod.ObjectMeta.Annotations[AMTD_APPLIED_SECURITY_EVENTS] = string(appliedSecurityEventsEncoded)

			err = r.Client.Update(ctx, pod)
			if err != nil {
				log.Error(err, fmt.Sprintf(`Failed to update pod: "%s": %s`, pod.Name, err.Error()))
				return ctrl.Result{}, err
			}

			log.Info(fmt.Sprintf(`SecurityEvent was sucessfully applied to the pod`))
		}

		// ---------------------------------------------------
		// Execute the proper action
		// ---------------------------------------------------
		switch action {
		case "delete":
			podErr := r.Client.Delete(ctx, pod)

			if err != nil && !errors.IsNotFound(err) {
				log.Error(podErr, fmt.Sprintf(`Failed to delete pod "%s"`, req.Name))
			}
			log.Info(fmt.Sprintf(`Pod: "%s" was sucessfully deleted with ACTION: delete`, pod.Name))
		case "quarantine":
			networkPolicyName := fmt.Sprintf("%s-%s-%s", pod.Namespace, pod.Name, "policy")

			networkPolicy := &v1.NetworkPolicy{}
			err = r.Client.Get(context.Background(), types.NamespacedName{
				Namespace: pod.Namespace,
				Name:      networkPolicyName,
			}, networkPolicy)

			if err != nil && errors.IsNotFound(err) {
				networkPolicy := &v1.NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      networkPolicyName,
						Namespace: pod.Namespace,
					},
					Spec: v1.NetworkPolicySpec{
						PodSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{AMTD_NETWORK_POLICY: networkPolicyName},
						},
						Ingress: []v1.NetworkPolicyIngressRule{},
						Egress:  []v1.NetworkPolicyEgressRule{},
						PolicyTypes: []v1.PolicyType{
							v1.PolicyTypeIngress,
							v1.PolicyTypeEgress,
						},
					},
				}

				// Set AMTD instance as the owner and controller for the NetworkPolicy
				err := ctrl.SetControllerReference(AMTD, networkPolicy, r.Scheme)
				if err != nil {
					log.Error(err, "Failed to set AMTD as owner and controller reference on NetworkPolicy",
						"AMTD", AMTD.ObjectMeta.Name,
						"NetworkPolicy", networkPolicy.Name,
						"Namespace", networkPolicy.Namespace,
					)
				}

				err = r.Create(ctx, networkPolicy)
				if err != nil {
					log.Error(err, "Failed to create Networkpolicy in the cluster",
						"NetworkPolicy", networkPolicy.Name,
						"Namespace", networkPolicy.Namespace)
				}

				// Relabel pod:
				// i) move labels under annotations to preserve them - except those that belong to AMTD management,
				for key, value := range pod.ObjectMeta.Labels {
					if !isSetContain(AMTD.Spec.PodSelector, map[string]string{key: value}) {
						pod.ObjectMeta.Annotations[key] = value
						delete(pod.ObjectMeta.Labels, key)
					}
				}
				// ii) add new label that match networkPolicy podSelector
				pod.ObjectMeta.Labels[AMTD_NETWORK_POLICY] = networkPolicyName

				err = r.Client.Update(ctx, pod)
				if err != nil {
					log.Error(err, fmt.Sprintf(`Failed to update pod: "%s": %s`, pod.Name, err.Error()))
					return ctrl.Result{}, err
				}

				log.Info(fmt.Sprintf(`Pod %s was put in quarantine`, pod.Name))
			}

			// Set AMTD instance as the owner and controller for the Pod - this
			// step cannot combined into a single update with relabel, because
			// until relabel another owner exists that cannot be updated
			// Actually since it's not immediate that OwnerReference is deleted
			// by ReplicaSet or sg. we need to reschedule and check it later
			err = ctrl.SetControllerReference(AMTD, pod, r.Scheme)
			if err != nil {
				log.Error(err, "Failed to set AMTD as owner and controller reference on Pod - Rescheduling and trying later",
					"AMTD", AMTD.ObjectMeta.Name,
					"Pod", pod.Name,
					"Namespace", networkPolicy.Namespace,
				)
				return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
			}

			err = r.Client.Update(ctx, pod)
			if err != nil {
				log.Error(err, fmt.Sprintf(`Failed to update pod: "%s": %s`, pod.Name, err.Error()))
				return ctrl.Result{}, err
			}
		default:
			log.Info(fmt.Sprintf(`ACTION: %s -> POD: %s - NOT IMPLEMENTED YET`, action, pod.Name))
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecurityEventReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&amtdv1beta1.SecurityEvent{}).
		Complete(r)
}

func isSetContain(set map[string]string, subset map[string]string) bool {
	if len(subset) == 0 {
		return false
	}
	for key, value := range subset {
		if value2, ok := set[key]; ok {
			if value != value2 {
				return false
			}
		} else {
			return false
		}
	}
	return true
}
