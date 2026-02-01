package collector

import (
	"context"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type MetricSample struct {
	Timestamp       time.Time `json:"timestamp"`
	HostID          string    `json:"host_id"`
	CPUPercent      float64   `json:"cpu_percent"`
	MemUsedPercent  float64   `json:"mem_used_percent"`
	DiskUsedPercent float64   `json:"disk_used_percent"`
	DiskReadBytes   uint64    `json:"disk_read_bytes"`
	DiskWriteBytes  uint64    `json:"disk_write_bytes"`
	NetRxBytes      uint64    `json:"net_rx_bytes"`
	NetTxBytes      uint64    `json:"net_tx_bytes"`
}

type Sampler struct {
	hostID string
}

func NewSampler(hostID string) *Sampler {
	return &Sampler{hostID: hostID}
}

func (s *Sampler) Sample(ctx context.Context) (MetricSample, error) {
	cpuPercents, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return MetricSample{}, err
	}
	cpuPercent := 0.0
	if len(cpuPercents) > 0 {
		cpuPercent = cpuPercents[0]
	}

	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return MetricSample{}, err
	}

	diskPath := "/"
	if runtime.GOOS == "windows" {
		diskPath = "C:\\"
	}
	usage, err := disk.UsageWithContext(ctx, diskPath)
	if err != nil {
		return MetricSample{}, err
	}

	ioCounters, err := disk.IOCountersWithContext(ctx)
	if err != nil {
		return MetricSample{}, err
	}
	var readBytes uint64
	var writeBytes uint64
	for _, stat := range ioCounters {
		readBytes += stat.ReadBytes
		writeBytes += stat.WriteBytes
	}

	netCounters, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		return MetricSample{}, err
	}
	var rxBytes uint64
	var txBytes uint64
	if len(netCounters) > 0 {
		rxBytes = netCounters[0].BytesRecv
		txBytes = netCounters[0].BytesSent
	}

	return MetricSample{
		Timestamp:       time.Now().UTC(),
		HostID:          s.hostID,
		CPUPercent:      cpuPercent,
		MemUsedPercent:  vm.UsedPercent,
		DiskUsedPercent: usage.UsedPercent,
		DiskReadBytes:   readBytes,
		DiskWriteBytes:  writeBytes,
		NetRxBytes:      rxBytes,
		NetTxBytes:      txBytes,
	}, nil
}
