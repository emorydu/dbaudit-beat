// Copyright 2024 Emory.Du <orangeduxiaocheng@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/emorydu/dbaudit/internal/common"
	"github.com/emorydu/dbaudit/internal/common/genproto/auditbeat"
	"github.com/emorydu/dbaudit/internal/common/gops"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

func (s service) UsageStatus() {
	req := &auditbeat.UsageStatusRequest{
		Ip: os.Getenv("LOCAL_IP"),
	}
	// TODO
	info := gops.ProcessByNameUsed("fluent-bit")
	pid, err := RunShellReturnPid(fluentBit)
	if err != nil || pid == "" {
		req.Status = common.BitStatusClosed
	} else {
		req.Status = common.BitStatusStartup
	}
	if info.MemoryUsage != 0 || info.CpuUsage != 0 {
		req.MemUsage = float64(info.MemoryUsage) / 1024 / 1024
		v := fmt.Sprintf("%.2f", info.CpuUsage)
		req.CpuUsage, _ = strconv.ParseFloat(v, 10)
	}
	_, err = s.cli.UsageStatus(s.ctx, req)
	if err != nil {
		logrus.Errorf("upload usage and status error: %v", err)
	}
}

func (s service) Usages() string {
	return "UsageStatus"
}
