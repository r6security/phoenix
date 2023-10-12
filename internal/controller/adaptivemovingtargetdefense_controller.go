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
	"strconv"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	amtdv1beta1 "github.com/r6security/phoenix/api/v1beta1"
)

// AdaptiveMovingTargetDefenseReconciler reconciles a AdaptiveMovingTargetDefense object
type AdaptiveMovingTargetDefenseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=amtd.r6security.com,resources=adaptivemovingtargetdefenses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=amtd.r6security.com,resources=adaptivemovingtargetdefenses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=amtd.r6security.com,resources=adaptivemovingtargetdefenses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AdaptiveMovingTargetDefense object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *AdaptiveMovingTargetDefenseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Start by declaring the custom resource to be type "AdaptiveMovingTargetDefense"
	AMTD := &amtdv1beta1.AdaptiveMovingTargetDefense{}

	// Then retrieve from the cluster the resource that triggered this reconciliation.
	// Store these contents into an object used throughout reconciliation.
	err := r.Client.Get(context.Background(), req.NamespacedName, AMTD)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf(`Custom resource for AdaptiveMovingTargetDefense "%s" does not exist, remove annotations from pods`, req.Namespace+"/"+req.Name))
			// Get pods based on PodSelector
			podList := &corev1.PodList{}
			listOptions := &client.ListOptions{Namespace: req.Namespace}
			if err = r.List(context.TODO(), podList, listOptions); err != nil {
				log.Error(err, fmt.Sprintf(`Failed to retrieve pods: "%s"`, err.Error()))
				return ctrl.Result{}, err
			}

			for _, pod := range podList.Items {
				// Check whether already AMTD "member", if so AMTD_MANAGED_TIME is not changed
				if _, found := pod.ObjectMeta.Annotations[AMTD_MANAGED_BY]; found {

					removeAMTDAnnotationFromPod(pod, log, req.Namespace, req.Name)

					// Try to apply this patch, if it fails, return the failure
					err = r.Client.Update(ctx, &pod)
					if err != nil {
						log.Error(err, fmt.Sprintf(`Failed to update pod: "%s": %s`, pod.Name, err.Error()))
						// this makes the controller to log the error and in the future ignore the this AMTD resource (at least until it changes)
						return ctrl.Result{}, err
					}

					log.Info(fmt.Sprintf(`Pod: "%s" was sucessfully updated with annotations`, pod.Name))
				}
			}
			return ctrl.Result{}, nil
		} else {
			// some other error happend
			log.Error(err, fmt.Sprintf(`Failed to retrieve custom resource "%s": %s`, req.Name, err.Error()))
			return ctrl.Result{}, err
		}
	}

	// Get pods based on PodSelector
	podSelectorSet := labels.SelectorFromSet(AMTD.Spec.PodSelector)
	podList := &corev1.PodList{}
	listOptions := &client.ListOptions{Namespace: AMTD.Namespace, LabelSelector: podSelectorSet}
	if err = r.List(context.TODO(), podList, listOptions); err != nil {
		log.Error(err, fmt.Sprintf(`Failed to retrieve pods: "%s"`, err.Error()))
		return ctrl.Result{}, err
	}

	// Annotate pods
	for _, pod := range podList.Items {

		var amtdManageInfoList []AMTDManageInfo

		// Check whether already AMTD "member", if so AMTD_MANAGED_TIME is not changed
		if _, found := pod.ObjectMeta.Annotations[AMTD_MANAGED_BY]; !found {
			// if no annotations set so far initialize Annotations first
			if pod.ObjectMeta.Annotations == nil {
				pod.ObjectMeta.Annotations = map[string]string{}
			}
		} else {
			// Already AMTD member
			json.Unmarshal([]byte(pod.ObjectMeta.Annotations[AMTD_MANAGED_BY]), &amtdManageInfoList)
		}

		// Check whether there is collision with other AMTD resources?
		for _, AMTDManageInfo := range amtdManageInfoList {
			// Do not want to test collision with self
			if !(AMTD.Namespace == AMTDManageInfo.AMTDNamespace && AMTD.Name == AMTDManageInfo.AMTDName) {
				AMTDOther := &amtdv1beta1.AdaptiveMovingTargetDefense{}
				err := r.Client.Get(context.Background(), types.NamespacedName{Namespace: AMTDManageInfo.AMTDNamespace, Name: AMTDManageInfo.AMTDName}, AMTDOther)
				if err != nil {
					if errors.IsNotFound(err) {
						// If the resource is not found, that is OK. It just means the desired state is to
						// not have any resources for this AdaptiveMovingTargetDefense but no delete is required.
						// TODO: remove the annotations
						log.Info(fmt.Sprintf(`Custom resource for AdaptiveMovingTargetDefense "%s" does not exist, remove annotations from pods`, req.Namespace+"/"+req.Name))
					} else {
						// some other error happend
						log.Error(err, fmt.Sprintf(`Failed to retrieve custom resource "%s": %s`, AMTDManageInfo.AMTDName, err.Error()))
						return ctrl.Result{}, err
					}
				}

				for _, strategyOther := range AMTDOther.Spec.Strategy {
					for _, strategy := range AMTD.Spec.Strategy {
						if reflect.DeepEqual(strategyOther.Rule, strategy.Rule) {
							log.Error(err, fmt.Sprintf(`AMTD RuleIDs collision "%s" <> "%s"`, AMTD.Name, AMTDOther.Name))
							return ctrl.Result{}, err
						}
					}
				}
			}
		}

		// Add AMTD manage info if necessary
		amtdManagedInfoExist := false
		for _, amtdManageInfo := range amtdManageInfoList {
			if amtdManageInfo.AMTDNamespace == AMTD.Namespace && amtdManageInfo.AMTDName == AMTD.Name {
				amtdManagedInfoExist = true
				break
			}
		}
		if !amtdManagedInfoExist {
			initalTime := strconv.FormatInt(time.Now().Unix(), 10)
			amtdManageInfoList = append(amtdManageInfoList, AMTDManageInfo{initalTime, AMTD.Namespace, AMTD.Name})
		}
		amtdManagedInfoListEncoded, err := json.Marshal(amtdManageInfoList)
		if err != nil {
			log.Error(err, fmt.Sprintf(`amtdManagedInfoList json encoding does not work: %s`, err.Error()))
		}
		pod.ObjectMeta.Annotations[AMTD_MANAGED_BY] = string(amtdManagedInfoListEncoded)

		// Try to apply this patch, if it fails, return the failure
		// TODO: before update we could check whether any changes would happen
		err = r.Client.Update(ctx, &pod)
		if err != nil {
			log.Error(err, fmt.Sprintf(`Failed to update pod: "%s": %s`, pod.Name, err.Error()))
			// this makes the controller to log the error and in the future ignore the this AMTD resource (at least until it changes)
			return ctrl.Result{}, err
		}

		log.Info(fmt.Sprintf(`Pod: "%s" was sucessfully updated with annotations`, pod.Name))
	}

	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func removeAMTDAnnotationFromPod(pod corev1.Pod, log logr.Logger, reqNamespace string, reqName string) {
	// Already AMTD member
	var amtdManageInfoList []AMTDManageInfo
	json.Unmarshal([]byte(pod.ObjectMeta.Annotations[AMTD_MANAGED_BY]), &amtdManageInfoList)
	for index, amtdManagedInfo := range amtdManageInfoList {
		if amtdManagedInfo.AMTDNamespace == reqNamespace && amtdManagedInfo.AMTDName == reqName {
			// Remove the element at index i from a
			amtdManageInfoList[index] = amtdManageInfoList[len(amtdManageInfoList)-1]  // Copy last element to index i.
			amtdManageInfoList[len(amtdManageInfoList)-1] = AMTDManageInfo{"", "", ""} // Erase last element (write zero value).
			amtdManageInfoList = amtdManageInfoList[:len(amtdManageInfoList)-1]        // Truncate slice.
			break
		}
	}

	if len(amtdManageInfoList) == 0 {
		delete(pod.ObjectMeta.Annotations, AMTD_MANAGED_BY)
	} else {
		amtdManagedInfoListEncoded, err := json.Marshal(amtdManageInfoList)
		if err != nil {
			log.Error(err, fmt.Sprintf(`amtdManagedInfoList json encoding does not work: %s`, err.Error()))
		}
		pod.ObjectMeta.Annotations[AMTD_MANAGED_BY] = string(amtdManagedInfoListEncoded)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *AdaptiveMovingTargetDefenseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&amtdv1beta1.AdaptiveMovingTargetDefense{}).
		Complete(r)
}
