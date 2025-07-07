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
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"

	amtdv1beta1 "github.com/r6security/phoenix/api/v1beta1"
)

func setupIntervalTimerTest() (client.Client, *runtime.Scheme, *IntervalTimerController) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = amtdv1beta1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	controller := NewIntervalTimerController(fakeClient, scheme)

	return fakeClient, scheme, controller
}

func TestParseIntervalConfig(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantMin   time.Duration
		wantMax   time.Duration
		wantError bool
	}{
		{
			name:    "valid minutes",
			input:   "30m-45m",
			wantMin: 30 * time.Minute,
			wantMax: 45 * time.Minute,
		},
		{
			name:    "valid seconds",
			input:   "1800s-2700s",
			wantMin: 1800 * time.Second,
			wantMax: 2700 * time.Second,
		},
		{
			name:    "valid hours",
			input:   "1h-2h",
			wantMin: 1 * time.Hour,
			wantMax: 2 * time.Hour,
		},
		{
			name:    "valid mixed",
			input:   "30m-1h30m",
			wantMin: 30 * time.Minute,
			wantMax: 90 * time.Minute,
		},
		{
			name:      "invalid format - no dash",
			input:     "30m",
			wantError: true,
		},
		{
			name:      "invalid format - multiple dashes",
			input:     "30m-45m-1h",
			wantError: true,
		},
		{
			name:      "invalid min duration",
			input:     "invalid-45m",
			wantError: true,
		},
		{
			name:      "invalid max duration",
			input:     "30m-invalid",
			wantError: true,
		},
		{
			name:      "min >= max",
			input:     "45m-30m",
			wantError: true,
		},
		{
			name:      "min < 1 minute",
			input:     "30s-45s",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parseIntervalConfig(tt.input)

			if tt.wantError {
				if err == nil {
					t.Errorf("parseIntervalConfig() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("parseIntervalConfig() unexpected error: %v", err)
				return
			}

			if config.MinInterval != tt.wantMin {
				t.Errorf("parseIntervalConfig() MinInterval = %v, want %v", config.MinInterval, tt.wantMin)
			}

			if config.MaxInterval != tt.wantMax {
				t.Errorf("parseIntervalConfig() MaxInterval = %v, want %v", config.MaxInterval, tt.wantMax)
			}
		})
	}
}

func TestCalculateNextTriggerTime(t *testing.T) {
	config := IntervalConfig{
		MinInterval: 30 * time.Minute,
		MaxInterval: 45 * time.Minute,
	}

	now := time.Now()
	
	// Test multiple calculations to ensure randomness
	times := make([]time.Time, 100)
	for i := 0; i < 100; i++ {
		times[i] = calculateNextTriggerTime(config)
	}

	// Check that all times are within the expected range
	for i, triggerTime := range times {
		duration := triggerTime.Sub(now)
		
		if duration < config.MinInterval {
			t.Errorf("Trigger time %d is too early: %v < %v", i, duration, config.MinInterval)
		}
		
		if duration > config.MaxInterval {
			t.Errorf("Trigger time %d is too late: %v > %v", i, duration, config.MaxInterval)
		}
	}

	// Check that we get some variation (not all the same)
	allSame := true
	for i := 1; i < len(times); i++ {
		if !times[i].Equal(times[0]) {
			allSame = false
			break
		}
	}
	
	if allSame {
		t.Error("All trigger times are identical - randomness not working")
	}
}

func TestIntervalTimerController_ReconcilePodTimer(t *testing.T) {
	fakeClient, _, controller := setupIntervalTimerTest()
	ctx := context.Background()

	// Test pod with interval timer enabled
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
			Annotations: map[string]string{
				INTERVAL_TIMER_ENABLED:    "true",
				INTERVAL_TIMER_ANNOTATION: "2m-5m",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "test-container", Image: "test-image"},
			},
		},
	}

	err := fakeClient.Create(ctx, pod)
	if err != nil {
		t.Fatalf("Failed to create test pod: %v", err)
	}

	// Create a logger for testing
	logger := log.Log.WithName("test")

	// Test reconciling the pod timer
	err = controller.reconcilePodTimer(ctx, pod, logger)
	if err != nil {
		t.Errorf("reconcilePodTimer() failed: %v", err)
	}

	// Check that timer state was created
	podKey := "test-namespace/test-pod"
	controller.timerMutex.RLock()
	state, exists := controller.timerStates[podKey]
	controller.timerMutex.RUnlock()

	if !exists {
		t.Error("Timer state was not created")
		return
	}

	if state.PodName != "test-pod" {
		t.Errorf("Timer state PodName = %v, want %v", state.PodName, "test-pod")
	}

	if state.Namespace != "test-namespace" {
		t.Errorf("Timer state Namespace = %v, want %v", state.Namespace, "test-namespace")
	}

	expectedMin := 2 * time.Minute
	expectedMax := 5 * time.Minute
	if state.IntervalConfig.MinInterval != expectedMin {
		t.Errorf("Timer state MinInterval = %v, want %v", state.IntervalConfig.MinInterval, expectedMin)
	}

	if state.IntervalConfig.MaxInterval != expectedMax {
		t.Errorf("Timer state MaxInterval = %v, want %v", state.IntervalConfig.MaxInterval, expectedMax)
	}
}

func TestIntervalTimerController_DisableTimer(t *testing.T) {
	fakeClient, _, controller := setupIntervalTimerTest()
	ctx := context.Background()

	// Create pod with timer enabled first
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
			Annotations: map[string]string{
				INTERVAL_TIMER_ENABLED:    "true",
				INTERVAL_TIMER_ANNOTATION: "2m-5m",
			},
		},
	}

	err := fakeClient.Create(ctx, pod)
	if err != nil {
		t.Fatalf("Failed to create test pod: %v", err)
	}

	logger := log.Log.WithName("test")

	// Enable timer first
	err = controller.reconcilePodTimer(ctx, pod, logger)
	if err != nil {
		t.Errorf("reconcilePodTimer() failed: %v", err)
	}

	podKey := "test-namespace/test-pod"
	
	// Verify timer was created
	controller.timerMutex.RLock()
	_, exists := controller.timerStates[podKey]
	controller.timerMutex.RUnlock()
	if !exists {
		t.Error("Timer state was not created")
	}

	// Now disable the timer
	pod.Annotations[INTERVAL_TIMER_ENABLED] = "false"
	err = fakeClient.Update(ctx, pod)
	if err != nil {
		t.Fatalf("Failed to update test pod: %v", err)
	}

	// Reconcile again
	err = controller.reconcilePodTimer(ctx, pod, logger)
	if err != nil {
		t.Errorf("reconcilePodTimer() failed: %v", err)
	}

	// Verify timer was removed
	controller.timerMutex.RLock()
	_, exists = controller.timerStates[podKey]
	controller.timerMutex.RUnlock()
	if exists {
		t.Error("Timer state was not removed when disabled")
	}
}

func TestActionOverlapPrevention(t *testing.T) {
	_, _, controller := setupIntervalTimerTest()

	state := &PodTimerState{
		PodName:   "test-pod",
		Namespace: "test-namespace",
		IntervalConfig: IntervalConfig{
			MinInterval: 5 * time.Minute,
			MaxInterval: 10 * time.Minute,
		},
	}

	logger := log.Log.WithName("test")

	// Initially should be able to trigger
	if !controller.shouldTriggerAction(state, logger) {
		t.Error("Should be able to trigger action initially")
	}

	// Mark action in progress
	controller.markActionInProgress(state)

	// Should not be able to trigger while action in progress
	if controller.shouldTriggerAction(state, logger) {
		t.Error("Should not be able to trigger action while in progress")
	}

	// Clear action
	podKey := "test-namespace/test-pod"
	controller.clearActionInProgress(podKey)

	// Should be able to trigger again
	if !controller.shouldTriggerAction(state, logger) {
		t.Error("Should be able to trigger action after clearing")
	}
}

func TestActionTimeout(t *testing.T) {
	_, _, controller := setupIntervalTimerTest()

	state := &PodTimerState{
		PodName:   "test-pod",
		Namespace: "test-namespace",
		IntervalConfig: IntervalConfig{
			MinInterval: 5 * time.Minute,
			MaxInterval: 10 * time.Minute,
		},
	}

	podKey := "test-namespace/test-pod"
	logger := log.Log.WithName("test")

	// Mark action in progress with old timestamp
	globalActionTracker.mu.Lock()
	globalActionTracker.actions[podKey] = time.Now().Add(-6 * time.Minute) // 6 minutes ago
	globalActionTracker.mu.Unlock()

	// Should be able to trigger due to timeout
	if !controller.shouldTriggerAction(state, logger) {
		t.Error("Should be able to trigger action after timeout")
	}

	// Action should be cleared after timeout check
	globalActionTracker.mu.RLock()
	_, exists := globalActionTracker.actions[podKey]
	globalActionTracker.mu.RUnlock()

	if exists {
		t.Error("Action should be cleared after timeout")
	}
}

func TestMinimumCooldownPeriod(t *testing.T) {
	_, _, controller := setupIntervalTimerTest()

	state := &PodTimerState{
		PodName:   "test-pod",
		Namespace: "test-namespace",
		IntervalConfig: IntervalConfig{
			MinInterval: 20 * time.Minute, // 20 minute minimum
			MaxInterval: 30 * time.Minute,
		},
		LastActionEnd: time.Now().Add(-2 * time.Minute), // 2 minutes ago
	}

	logger := log.Log.WithName("test")

	// Should not be able to trigger due to cooldown (need 5 minutes = 25% of 20 minutes)
	if controller.shouldTriggerAction(state, logger) {
		t.Error("Should not be able to trigger during cooldown period")
	}

	// Set last action to 6 minutes ago (more than 5 minute cooldown)
	state.LastActionEnd = time.Now().Add(-6 * time.Minute)

	// Should be able to trigger now
	if !controller.shouldTriggerAction(state, logger) {
		t.Error("Should be able to trigger after cooldown period")
	}
} 