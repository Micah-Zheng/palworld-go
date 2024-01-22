package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type MemoryCheckTask struct {
	Config     Config
	Ticker     *time.Ticker
	BackupTask *BackupTask
}

func NewMemoryCheckTask(config Config, BackupTask *BackupTask) *MemoryCheckTask {
	return &MemoryCheckTask{
		Config:     config,
		Ticker:     time.NewTicker(time.Duration(config.MemoryCheckInterval) * time.Second),
		BackupTask: BackupTask,
	}
}

func (task *MemoryCheckTask) Schedule() {
	for range task.Ticker.C {
		task.checkMemory()
	}
}

func (task *MemoryCheckTask) checkMemory() {
	var cmd *exec.Cmd
	threshold := task.Config.MemoryUsageThreshold

	if runtime.GOOS == "windows" {
		cmd = exec.Command("wmic", "OS", "get", "FreePhysicalMemory", "/Value")
	} else {
		cmd = exec.Command("sh", "-c", "free | grep Mem | awk '{print $3/$2 * 100.0}'")
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Printf("Failed to execute memory check command: %v", err)
		return
	}

	memoryUsage, err := task.parseMemoryUsage(out.String(), runtime.GOOS)
	if err != nil {
		log.Printf("Failed to parse memory usage: %v", err)
		return
	}

	log.Printf("Now Memory usage is  %v%%.", memoryUsage)

	if memoryUsage > threshold {
		log.Printf("Memory usage is above %v%%. Running clean command.", threshold)
		// 初始化RCON客户端
		rconClient := NewRconClient(task.Config.Address, task.Config.AdminPassword, task.BackupTask)
		HandleMemoryUsage(threshold, rconClient)
		defer rconClient.Close()
	} else {
		log.Printf("Memory usage is below %v%%. No action required.", threshold)
	}
}

func (task *MemoryCheckTask) parseMemoryUsage(output, os string) (float64, error) {
	if os == "windows" {
		lines := strings.Fields(output)
		if len(lines) < 1 {
			return 0, fmt.Errorf("unexpected output format")
		}
		freeMemoryKB, err := strconv.ParseFloat(strings.TrimPrefix(lines[0], "FreePhysicalMemory="), 64)
		if err != nil {
			return 0, err
		}
		log.Printf("now FreePhysicalMemoryKB: %v", freeMemoryKB)
		totalMemoryKB := task.Config.TotalMemoryGB * 1024 * 1024
		return 100.0 * (1 - freeMemoryKB/float64(totalMemoryKB)), nil
	} else {
		return strconv.ParseFloat(strings.TrimSpace(output), 64)
	}
}
