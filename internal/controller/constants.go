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
	AMTD_MANAGED_TIME   string = "amtd.r6security.com/managed-time"
	AMTD_MANAGED_BY     string = "amtd.r6security.com/managed-by"
	AMTD_STRATEGY_BASE  string = "amtd.r6security.com/strategy-"
	AMTD_NETWORK_POLICY string = "amtd.r6security.com/network-policy"

	AMTD_APPLIED_SECURITY_EVENTS string = "amtd.r6security.com/applied-sec-events"
	R6_SECURITY_EVENT_RECEIVED   string = "amtd.r6security.event.received"

	// R6Security label for AMTD-managed pods (GitHub issue #15)
	R6_SECURITY_MANAGED_LABEL    string = "r6security.com/managed-by-amtd"
)
