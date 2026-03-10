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

const (
	// Trigger type/source that the NIM controller reacts to (time-based restarts)
	TriggerTypeTimed              string = "timed"
	TriggerSourceTimeBasedTrigger string = "TimeBasedTrigger"

	// NIM (Node Infrastructure Module) annotation constants
	NIM_POD_STARTUP_STATE      string = "nim.r6security.com/pod-startup-state"
	NIM_STARTUP_STATE_PENDING  string = "pending"
	NIM_STARTUP_STATE_STARTING string = "starting"
	NIM_STARTUP_STATE_RUNNING  string = "running"
	NIM_STARTUP_STATE_FAILED   string = "failed"
	NIM_RESCHEDULE_COUNT       string = "nim.r6security.com/reschedule-count"
	NIM_TRIGGERS_STOPPED       string = "nim.r6security.com/triggers-stopped"
	NIM_ACTION_ACTIVE          string = "nim.r6security.com/action-active"
	NIM_LAST_ACTION_TIMESTAMP  string = "nim.r6security.com/last-action-timestamp"
)
