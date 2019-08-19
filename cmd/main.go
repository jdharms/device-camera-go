// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 Dell Technologies
//
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/edgexfoundry/device-sdk-go/pkg/startup"

	"device-camera-go/internal/driver"
)

const (
	version     string = "0.1.0"
	serviceName string = "device-camera-go"
)

func main() {
	sd := driver.NewProtocolDriver()
	startup.Bootstrap(serviceName, version, sd)
}
