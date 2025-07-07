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
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	amtdv1beta1 "github.com/r6security/phoenix/api/v1beta1"
)

// IntervalConfig represents the timing configuration for interval-based triggers
type IntervalConfig struct {
	MinInterval time.Duration `json:"minInterval"`
	MaxInterval time.Duration `json:"maxInterval"`
}

// PodTimerState tracks the timer state for each pod
type PodTimerState struct {
	PodName          string            `json:"podName"`
	Namespace        string            `json:"namespace"`
	IntervalConfig   IntervalConfig    `json:"intervalConfig"`
	NextTriggerTime  time.Time        `json:"nextTriggerTime"`
	ActionInProgress bool             `json:"actionInProgress"`
	LastActionStart  time.Time        `json:"lastActionStart"`
	LastActionEnd    time.Time        `json:"lastActionEnd"`
	TimerID          string           `json:"timerID"`
}

// IntervalTimerController manages interval-based timer triggers
type IntervalTimerController struct {
	client.Client
	Scheme     *runtime.Scheme
	timerStates map[string]*PodTimerState
	timerMutex  sync.RWMutex
	stopChans   map[string]chan struct{}
}

// ActionTracker tracks ongoing actions to prevent overlap
type ActionTracker struct {
	mu      sync.RWMutex
	actions map[string]time.Time // podKey -> action start time
}

var globalActionTracker = &ActionTracker{
	actions: make(map[string]time.Time),
}

const (
	INTERVAL_TIMER_ANNOTATION = "interval-timer.amtd.r6security.com/config"
	INTERVAL_TIMER_ENABLED    = "interval-timer.amtd.r6security.com/enabled"
	ACTION_IN_PROGRESS        = "interval-timer.amtd.r6security.com/action-in-progress"
)

// NewIntervalTimerController creates a new interval timer controller
func NewIntervalTimerController(client client.Client, scheme *runtime.Scheme) *IntervalTimerController {
	return &IntervalTimerController{
		Client:      client,
		Scheme:      scheme,
		timerStates: make(map[string]*PodTimerState),
		stopChans:   make(map[string]chan struct{}),
	}
}

// Reconcile handles interval timer reconciliation
func (r *IntervalTimerController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Get all pods to check for interval timer annotations
	podList := &corev1.PodList{}
	if err := r.List(ctx, podList); err != nil {
		log.Error(err, "Failed to list pods")
		return ctrl.Result{}, err
	}

	for _, pod := range podList.Items {
		if err := r.reconcilePodTimer(ctx, &pod, log); err != nil {
			log.Error(err, fmt.Sprintf("Failed to reconcile timer for pod %s/%s", pod.Namespace, pod.Name))
			continue
		}
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *IntervalTimerController) reconcilePodTimer(ctx context.Context, pod *corev1.Pod, log logr.Logger) error {
	podKey := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

	// Check if interval timer is enabled for this pod
	enabled, exists := pod.Annotations[INTERVAL_TIMER_ENABLED]
	if !exists || enabled != "true" {
		r.stopPodTimer(podKey, log)
		return nil
	}

	// Parse interval configuration
	intervalConfigStr, exists := pod.Annotations[INTERVAL_TIMER_ANNOTATION]
	if !exists {
		log.Info(fmt.Sprintf("Pod %s has interval timer enabled but no config", podKey))
		return nil
	}

	intervalConfig, err := parseIntervalConfig(intervalConfigStr)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to parse interval config for pod %s: %s", podKey, intervalConfigStr))
		return err
	}

	// Start or update timer for this pod
	return r.ensurePodTimer(ctx, pod, intervalConfig, log)
}

func (r *IntervalTimerController) ensurePodTimer(ctx context.Context, pod *corev1.Pod, config IntervalConfig, log logr.Logger) error {
	podKey := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

	r.timerMutex.Lock()
	defer r.timerMutex.Unlock()

	// Check if timer already exists and config hasn't changed
	if state, exists := r.timerStates[podKey]; exists {
		if state.IntervalConfig.MinInterval == config.MinInterval && 
		   state.IntervalConfig.MaxInterval == config.MaxInterval {
			return nil // Timer already running with correct config
		}
		// Config changed, stop old timer
		if stopChan, exists := r.stopChans[podKey]; exists {
			close(stopChan)
			delete(r.stopChans, podKey)
		}
	}

	// Create new timer state
	timerState := &PodTimerState{
		PodName:         pod.Name,
		Namespace:       pod.Namespace,
		IntervalConfig:  config,
		NextTriggerTime: calculateNextTriggerTime(config),
		TimerID:         fmt.Sprintf("%s-%d", podKey, time.Now().Unix()),
	}

	r.timerStates[podKey] = timerState

	// Start timer goroutine
	stopChan := make(chan struct{})
	r.stopChans[podKey] = stopChan

	go r.runPodTimer(ctx, timerState, stopChan, log)

	log.Info(fmt.Sprintf("Started interval timer for pod %s with config %+v, next trigger: %s",
		podKey, config, timerState.NextTriggerTime.Format(time.RFC3339)))

	return nil
}

func (r *IntervalTimerController) runPodTimer(ctx context.Context, state *PodTimerState, stopChan <-chan struct{}, log logr.Logger) {
	podKey := fmt.Sprintf("%s/%s", state.Namespace, state.PodName)
	
	for {
		// Calculate time until next trigger
		now := time.Now()
		timeUntilTrigger := time.Until(state.NextTriggerTime)

		if timeUntilTrigger <= 0 {
			// Time to trigger, but check for action overlap
			if r.shouldTriggerAction(state, log) {
				r.triggerSecurityEvent(ctx, state, log)
				// Calculate next trigger time after successful action
				state.NextTriggerTime = calculateNextTriggerTime(state.IntervalConfig)
				log.Info(fmt.Sprintf("Triggered interval-based SecurityEvent for pod %s, next trigger: %s",
					podKey, state.NextTriggerTime.Format(time.RFC3339)))
			} else {
				// Action in progress or too soon since last action, delay by 1 minute
				state.NextTriggerTime = now.Add(time.Minute)
				log.Info(fmt.Sprintf("Delaying trigger for pod %s due to action in progress, next check: %s",
					podKey, state.NextTriggerTime.Format(time.RFC3339)))
			}
			continue
		}

		// Wait for either the trigger time or stop signal
		select {
		case <-time.After(timeUntilTrigger):
			// Continue to trigger logic
		case <-stopChan:
			log.Info(fmt.Sprintf("Stopping interval timer for pod %s", podKey))
			return
		case <-ctx.Done():
			log.Info(fmt.Sprintf("Context cancelled, stopping interval timer for pod %s", podKey))
			return
		}
	}
}

func (r *IntervalTimerController) shouldTriggerAction(state *PodTimerState, log logr.Logger) bool {
	podKey := fmt.Sprintf("%s/%s", state.Namespace, state.PodName)

	// Check global action tracker
	globalActionTracker.mu.RLock()
	actionStart, actionInProgress := globalActionTracker.actions[podKey]
	globalActionTracker.mu.RUnlock()

	if actionInProgress {
		// Check if action has been running too long (timeout after 5 minutes)
		if time.Since(actionStart) > 5*time.Minute {
			log.Info(fmt.Sprintf("Action timeout detected for pod %s, allowing new trigger", podKey))
			r.clearActionInProgress(podKey)
			return true
		}
		log.Info(fmt.Sprintf("Action still in progress for pod %s (started %s ago)", 
			podKey, time.Since(actionStart).String()))
		return false
	}

	// Check minimum interval since last action
	if !state.LastActionEnd.IsZero() {
		timeSinceLastAction := time.Since(state.LastActionEnd)
		minCooldown := state.IntervalConfig.MinInterval / 4 // 25% of min interval as cooldown
		if timeSinceLastAction < minCooldown {
			log.Info(fmt.Sprintf("Too soon since last action for pod %s (only %s ago, need %s)",
				podKey, timeSinceLastAction.String(), minCooldown.String()))
			return false
		}
	}

	return true
}

func (r *IntervalTimerController) triggerSecurityEvent(ctx context.Context, state *PodTimerState, log logr.Logger) {
	podKey := fmt.Sprintf("%s/%s", state.Namespace, state.PodName)

	// Mark action as in progress
	r.markActionInProgress(state)

	// Create SecurityEvent
	securityEvent := &amtdv1beta1.SecurityEvent{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("interval-timer-%s-%d", 
				strings.ReplaceAll(podKey, "/", "-"), 
				time.Now().Unix()),
			Labels: map[string]string{
				"interval-timer.amtd.r6security.com/pod-name":      state.PodName,
				"interval-timer.amtd.r6security.com/pod-namespace": state.Namespace,
				"interval-timer.amtd.r6security.com/trigger-type":  "interval-timer",
			},
		},
		Spec: amtdv1beta1.SecurityEventSpec{
			Targets: []string{podKey},
			Rule: amtdv1beta1.Rule{
				Source:      "IntervalTimerTrigger",
				ThreatLevel: "info",
				Type:        "interval-timer",
			},
			Description: fmt.Sprintf("Interval-based timer trigger (interval: %s-%s) for pod %s",
				state.IntervalConfig.MinInterval.String(),
				state.IntervalConfig.MaxInterval.String(),
				podKey),
		},
	}

	if err := r.Create(ctx, securityEvent); err != nil {
		log.Error(err, fmt.Sprintf("Failed to create interval-based SecurityEvent for pod %s", podKey))
		r.clearActionInProgress(podKey)
		return
	}

	log.Info(fmt.Sprintf("Created interval-based SecurityEvent %s for pod %s (interval: %s-%s)",
		securityEvent.Name, podKey,
		state.IntervalConfig.MinInterval.String(),
		state.IntervalConfig.MaxInterval.String()))

	// Schedule action completion tracking
	go r.trackActionCompletion(podKey, state, log)
}

func (r *IntervalTimerController) trackActionCompletion(podKey string, state *PodTimerState, log logr.Logger) {
	// Wait for action to complete (simulate with timeout)
	// In a real implementation, this would watch for the actual action completion
	time.Sleep(10 * time.Second) // Simulate action duration

	r.clearActionInProgress(podKey)
	state.LastActionEnd = time.Now()

	log.Info(fmt.Sprintf("Action completed for pod %s", podKey))
}

func (r *IntervalTimerController) markActionInProgress(state *PodTimerState) {
	podKey := fmt.Sprintf("%s/%s", state.Namespace, state.PodName)
	
	globalActionTracker.mu.Lock()
	globalActionTracker.actions[podKey] = time.Now()
	globalActionTracker.mu.Unlock()

	state.ActionInProgress = true
	state.LastActionStart = time.Now()
}

func (r *IntervalTimerController) clearActionInProgress(podKey string) {
	globalActionTracker.mu.Lock()
	delete(globalActionTracker.actions, podKey)
	globalActionTracker.mu.Unlock()

	// Also update state if it exists
	r.timerMutex.RLock()
	if state, exists := r.timerStates[podKey]; exists {
		state.ActionInProgress = false
	}
	r.timerMutex.RUnlock()
}

func (r *IntervalTimerController) stopPodTimer(podKey string, log logr.Logger) {
	r.timerMutex.Lock()
	defer r.timerMutex.Unlock()

	if stopChan, exists := r.stopChans[podKey]; exists {
		close(stopChan)
		delete(r.stopChans, podKey)
		log.Info(fmt.Sprintf("Stopped interval timer for pod %s", podKey))
	}

	delete(r.timerStates, podKey)
	r.clearActionInProgress(podKey)
}

// parseIntervalConfig parses interval configuration from annotation
// Format: "30m-45m" or "1800s-2700s" or "30-45m"
func parseIntervalConfig(configStr string) (IntervalConfig, error) {
	// Remove whitespace
	configStr = strings.TrimSpace(configStr)

	// Split on dash
	parts := strings.Split(configStr, "-")
	if len(parts) != 2 {
		return IntervalConfig{}, fmt.Errorf("invalid interval format, expected 'min-max' (e.g., '30m-45m')")
	}

	minDuration, err := time.ParseDuration(parts[0])
	if err != nil {
		return IntervalConfig{}, fmt.Errorf("invalid minimum duration: %v", err)
	}

	maxDuration, err := time.ParseDuration(parts[1])
	if err != nil {
		return IntervalConfig{}, fmt.Errorf("invalid maximum duration: %v", err)
	}

	if minDuration >= maxDuration {
		return IntervalConfig{}, fmt.Errorf("minimum duration must be less than maximum duration")
	}

	if minDuration < time.Minute {
		return IntervalConfig{}, fmt.Errorf("minimum duration must be at least 1 minute")
	}

	return IntervalConfig{
		MinInterval: minDuration,
		MaxInterval: maxDuration,
	}, nil
}

// calculateNextTriggerTime calculates a random time within the interval
func calculateNextTriggerTime(config IntervalConfig) time.Time {
	now := time.Now()
	
	// Calculate random duration within the interval
	intervalRange := config.MaxInterval - config.MinInterval
	randomOffset := time.Duration(rand.Int63n(int64(intervalRange)))
	randomDuration := config.MinInterval + randomOffset

	return now.Add(randomDuration)
}

// SetupWithManager sets up the controller with the Manager
func (r *IntervalTimerController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
} 