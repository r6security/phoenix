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

// Package controller implements the NIM reconciler. It watches pods that have the
// applied SecurityEvents annotation set by the Phoenix operator (or Time-based Trigger)
// and, when the applied events include a time-based trigger, performs NIM-enhanced
// restarts: updates NIM annotations on the pod and then deletes the pod so the
// workload is rescheduled.
package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	amtdv1beta1 "github.com/r6security/phoenix/api/v1beta1"
)

// NIMPolicy holds configurable NIM behavior (grace period, requeue interval, trigger type/source).
// Zero values are replaced by defaults in effective().
type NIMPolicy struct {
	GracePeriodSeconds int64
	RequeueAfter       time.Duration
	TriggerType        string
	TriggerSource      string
}

// DefaultNIMPolicy returns the default NIM policy (30s grace, 10s requeue, timed/TimeBasedTrigger).
func DefaultNIMPolicy() NIMPolicy {
	return NIMPolicy{
		GracePeriodSeconds: 30,
		RequeueAfter:       10 * time.Second,
		TriggerType:        TriggerTypeTimed,
		TriggerSource:      TriggerSourceTimeBasedTrigger,
	}
}

func (p NIMPolicy) effective() NIMPolicy {
	def := DefaultNIMPolicy()
	if p.GracePeriodSeconds <= 0 {
		p.GracePeriodSeconds = def.GracePeriodSeconds
	}
	if p.RequeueAfter <= 0 {
		p.RequeueAfter = def.RequeueAfter
	}
	if p.TriggerType == "" {
		p.TriggerType = def.TriggerType
	}
	if p.TriggerSource == "" {
		p.TriggerSource = def.TriggerSource
	}
	return p
}

// NIMReconciler watches pods and applies NIM enhancements when they have time-based SecurityEvent annotations.
type NIMReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Policy NIMPolicy
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get
//+kubebuilder:rbac:groups=amtd.r6security.com,resources=adaptivemovingtargetdefenses,verbs=get;list;watch

func (r *NIMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	pod := &corev1.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("NIM controller processing pod", "pod", pod.Name, "namespace", pod.Namespace)

	if pod.Annotations == nil {
		logger.Info("Pod has no annotations, skipping", "pod", pod.Name)
		return ctrl.Result{}, nil
	}

	appliedEventsJSON, hasAppliedEvents := pod.Annotations[AMTD_APPLIED_SECURITY_EVENTS]
	if !hasAppliedEvents {
		logger.Info("Pod has no SecurityEvent annotations, skipping", "pod", pod.Name)
		return ctrl.Result{}, nil
	}

	logger.Info("Processing pod for NIM restart", "pod", pod.Name)
	logger.Info("Pod has SecurityEvent annotation, processing for NIM", "pod", pod.Name, "appliedEvents", appliedEventsJSON)

	var appliedEvents []amtdv1beta1.SecurityEvent
	if err := json.Unmarshal([]byte(appliedEventsJSON), &appliedEvents); err != nil {
		logger.Error(err, "Failed to parse applied SecurityEvents", "pod", pod.Name)
		return ctrl.Result{}, nil
	}

	policy := r.Policy.effective()
	timedEvent := FindTimedEvent(appliedEvents, policy.TriggerType, policy.TriggerSource)
	if timedEvent == nil {
		return ctrl.Result{}, nil
	}

	if err := r.processPodWithNIM(ctx, pod, timedEvent, logger); err != nil {
		logger.Error(err, "Failed to process pod with NIM", "pod", pod.Name)
		return ctrl.Result{}, err
	}

	logger.Info("NIM scheduling next automatic restart", "pod", pod.Name, "requeueAfter", policy.RequeueAfter)
	return ctrl.Result{RequeueAfter: policy.RequeueAfter}, nil
}

// FindTimedEvent returns the first SecurityEvent in events whose rule type and source match.
func FindTimedEvent(events []amtdv1beta1.SecurityEvent, triggerType, triggerSource string) *amtdv1beta1.SecurityEvent {
	for i := range events {
		if events[i].Spec.Rule.Type == triggerType && events[i].Spec.Rule.Source == triggerSource {
			return &events[i]
		}
	}
	return nil
}

func (r *NIMReconciler) processPodWithNIM(ctx context.Context, pod *corev1.Pod, securityEvent *amtdv1beta1.SecurityEvent, logger logr.Logger) error {
	policy := r.Policy.effective()
	gracePeriodSeconds := policy.GracePeriodSeconds
	startupState := r.determinePodStartupState(*pod)

	logger.Info("NIM proceeding with pod reschedule", "pod", pod.Name, "securityEvent", securityEvent.Name)

	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	pod.Annotations[NIM_POD_STARTUP_STATE] = startupState
	pod.Annotations[NIM_ACTION_ACTIVE] = "true"
	pod.Annotations[NIM_LAST_ACTION_TIMESTAMP] = strconv.FormatInt(time.Now().Unix(), 10)
	pod.Annotations[NIM_TRIGGERS_STOPPED] = "true"

	if err := r.Update(ctx, pod); err != nil {
		return fmt.Errorf("failed to update pod annotations: %w", err)
	}

	deleteOpts := &client.DeleteOptions{}
	gp := gracePeriodSeconds
	deleteOpts.GracePeriodSeconds = &gp
	logger.Info("NIM executing pod delete", "pod", pod.Name, "gracePeriod", gp)
	if err := r.Delete(ctx, pod, deleteOpts); err != nil {
		logger.Error(err, "Failed to delete pod for NIM reschedule", "pod", pod.Name)
		return fmt.Errorf("failed to delete pod for NIM reschedule: %w", err)
	}

	logger.Info("NIM reschedule executed successfully", "pod", pod.Name, "securityEvent", securityEvent.Name)
	return nil
}

// SetupWithManager sets up the NIM controller with the Manager.
func (r *NIMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Named("nim-pod").
		Complete(r)
}

func (r *NIMReconciler) determinePodStartupState(pod corev1.Pod) string {
	switch pod.Status.Phase {
	case corev1.PodPending:
		return NIM_STARTUP_STATE_PENDING
	case corev1.PodRunning:
		return NIM_STARTUP_STATE_RUNNING
	case corev1.PodFailed:
		return NIM_STARTUP_STATE_FAILED
	default:
		return NIM_STARTUP_STATE_STARTING
	}
}
