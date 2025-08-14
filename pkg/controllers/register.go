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

package controllers

import (
    ctrl "sigs.k8s.io/controller-runtime"

    internalcontroller "github.com/r6security/phoenix/internal/controller"
)

// RegisterCoreControllers registers all core Phoenix controllers with the manager.
// This wrapper keeps controller implementations internal while exposing a public entrypoint.
func RegisterCoreControllers(mgr ctrl.Manager) error {
    if err := (&internalcontroller.AdaptiveMovingTargetDefenseReconciler{
        Client: mgr.GetClient(),
        Scheme: mgr.GetScheme(),
    }).SetupWithManager(mgr); err != nil {
        return err
    }

    if err := (&internalcontroller.PodReconciler{
        Client: mgr.GetClient(),
        Scheme: mgr.GetScheme(),
    }).SetupWithManager(mgr); err != nil {
        return err
    }

    if err := (&internalcontroller.SecurityEventReconciler{
        Client: mgr.GetClient(),
        Scheme: mgr.GetScheme(),
    }).SetupWithManager(mgr); err != nil {
        return err
    }

    return nil
}

// RegisterAMTDAndPodControllers registers only AMTD and Pod controllers.
func RegisterAMTDAndPodControllers(mgr ctrl.Manager) error {
    if err := (&internalcontroller.AdaptiveMovingTargetDefenseReconciler{
        Client: mgr.GetClient(),
        Scheme: mgr.GetScheme(),
    }).SetupWithManager(mgr); err != nil {
        return err
    }

    if err := (&internalcontroller.PodReconciler{
        Client: mgr.GetClient(),
        Scheme: mgr.GetScheme(),
    }).SetupWithManager(mgr); err != nil {
        return err
    }
    return nil
}


