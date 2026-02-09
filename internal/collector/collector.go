package collector

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type ProcessAttribution struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	CPUPercent float64 `json:"cpu_percent"`
	RSSBytes   uint64  `json:"rss_bytes"`
}

type MetricSample struct {
	Timestamp       time.Time           `json:"timestamp"`
	HostID          string              `json:"host_id"`
	CPUPercent      float64             `json:"cpu_percent"`
	MemUsedPercent  float64             `json:"mem_used_percent"`
	DiskUsedPercent float64             `json:"disk_used_percent"`
	DiskReadBytes   uint64              `json:"disk_read_bytes"`
	DiskWriteBytes  uint64              `json:"disk_write_bytes"`
	NetRxBytes      uint64              `json:"net_rx_bytes"`
	NetTxBytes      uint64              `json:"net_tx_bytes"`
	TopCPUProcess   *ProcessAttribution `json:"top_cpu_process,omitempty"`
	TopMemProcess   *ProcessAttribution `json:"top_mem_process,omitempty"`
	MetricFamilies  *MetricFamilies     `json:"metric_families,omitempty"`
}

type MetricFamilies struct {
	CPU  bool `json:"cpu"`
	Mem  bool `json:"mem"`
	Disk bool `json:"disk"`
	Net  bool `json:"net"`
}

func DefaultMetricFamilies() MetricFamilies {
	return MetricFamilies{CPU: true, Mem: true, Disk: true, Net: true}
}

type Sampler struct {
	hostID             string
	processAttribution bool
	metrics            MetricFamilies
}

func NewSampler(hostID string, processAttribution bool, metrics MetricFamilies) *Sampler {
	return &Sampler{hostID: hostID, processAttribution: processAttribution, metrics: metrics}
}

func (s *Sampler) Sample(ctx context.Context) (MetricSample, error) {
	cpuPercent := 0.0
	if s.metrics.CPU {
		cpuPercents, err := cpu.PercentWithContext(ctx, 0, false)
		if err != nil {
			return MetricSample{}, err
		}
		if len(cpuPercents) > 0 {
			cpuPercent = cpuPercents[0]
		}
	}

	memUsedPercent := 0.0
	if s.metrics.Mem {
		vm, err := mem.VirtualMemoryWithContext(ctx)
		if err != nil {
			return MetricSample{}, err
		}
		memUsedPercent = vm.UsedPercent
	}

	diskUsedPercent := 0.0
	var readBytes uint64
	var writeBytes uint64
	if s.metrics.Disk {
		diskPath := "/"
		if runtime.GOOS == "windows" {
			diskPath = "C:\\"
		}
		usage, err := disk.UsageWithContext(ctx, diskPath)
		if err != nil {
			return MetricSample{}, err
		}
		diskUsedPercent = usage.UsedPercent

		ioCounters, err := disk.IOCountersWithContext(ctx)
		if err != nil {
			return MetricSample{}, err
		}
		for _, stat := range ioCounters {
			readBytes += stat.ReadBytes
			writeBytes += stat.WriteBytes
		}
	}

	var rxBytes uint64
	var txBytes uint64
	if s.metrics.Net {
		netCounters, err := net.IOCountersWithContext(ctx, false)
		if err != nil {
			return MetricSample{}, err
		}
		if len(netCounters) > 0 {
			rxBytes = netCounters[0].BytesRecv
			txBytes = netCounters[0].BytesSent
		}
	}

	var topCPUProcess *ProcessAttribution
	var topMemProcess *ProcessAttribution
	if s.processAttribution {
		topCPUProcess, topMemProcess = sampleTopProcesses(ctx)
	}

	return MetricSample{
		Timestamp:       time.Now().UTC(),
		HostID:          s.hostID,
		CPUPercent:      cpuPercent,
		MemUsedPercent:  memUsedPercent,
		DiskUsedPercent: diskUsedPercent,
		DiskReadBytes:   readBytes,
		DiskWriteBytes:  writeBytes,
		NetRxBytes:      rxBytes,
		NetTxBytes:      txBytes,
		TopCPUProcess:   topCPUProcess,
		TopMemProcess:   topMemProcess,
		MetricFamilies:  &MetricFamilies{CPU: s.metrics.CPU, Mem: s.metrics.Mem, Disk: s.metrics.Disk, Net: s.metrics.Net},
	}, nil
}

func sampleTopProcesses(ctx context.Context) (*ProcessAttribution, *ProcessAttribution) {
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, nil
	}

	var topCPU *ProcessAttribution
	var topMem *ProcessAttribution

	for _, p := range processes {
		if ctx.Err() != nil {
			break
		}
		snapshot, ok := readProcessAttribution(ctx, p)
		if !ok {
			continue
		}
		if topCPU == nil || snapshot.CPUPercent > topCPU.CPUPercent {
			candidate := snapshot
			topCPU = &candidate
		}
		if topMem == nil || snapshot.RSSBytes > topMem.RSSBytes {
			candidate := snapshot
			topMem = &candidate
		}
	}

	return topCPU, topMem
}

func readProcessAttribution(ctx context.Context, p *process.Process) (ProcessAttribution, bool) {
	name, err := p.NameWithContext(ctx)
	if err != nil {
		name = fmt.Sprintf("pid-%d", p.Pid)
	}

	cpuPercent, cpuErr := p.CPUPercentWithContext(ctx)
	memoryInfo, memErr := p.MemoryInfoWithContext(ctx)
	if cpuErr != nil && memErr != nil {
		return ProcessAttribution{}, false
	}

	var rssBytes uint64
	if memErr == nil && memoryInfo != nil {
		rssBytes = memoryInfo.RSS
	}

	return ProcessAttribution{
		PID:        p.Pid,
		Name:       name,
		CPUPercent: cpuPercent,
		RSSBytes:   rssBytes,
	}, true
}
