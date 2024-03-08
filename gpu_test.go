package gputil

import (
	"context"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

// timestamp, index, uuid, utilization.gpu [%], memory.total [MiB], memory.used [MiB], memory.free [MiB], driver_version, name, serial, display_active, display_mode, temperature.gpu
var example = `
0, GPU-fd189414-e0f6-58a0-7031-fefe0ce43b1d, 0, 81920, 2, 81226, 535.104.12, NVIDIA A800-SXM4-80GB, 1321923001828, Disabled, Enabled, 31, 2024/03/08 13:49:22.063
1, GPU-121ebc1f-3e5d-139d-7aac-57311d5bafc7, 0, 81920, 2, 81226, 535.104.12, NVIDIA A800-SXM4-80GB, 1321923003018, Disabled, Enabled, 30, 2024/03/08 13:49:22.064
2, GPU-b74d1aeb-0aab-b3ca-ff55-94e24cbe0cd6, 0, 81920, 2, 81226, 535.104.12, NVIDIA A800-SXM4-80GB, 1321423018183, Disabled, Enabled, 30, 2024/03/08 13:49:22.065
3, GPU-401e53f2-8f44-1fc5-469d-8e36c1d6c9c5, 0, 81920, 2, 81226, 535.104.12, NVIDIA A800-SXM4-80GB, 1321923002415, Disabled, Enabled, 30, 2024/03/08 13:49:22.065
4, GPU-8b63b1f2-98e1-b24e-f59f-d725d51b3a2b, 0, 81920, 2, 81226, 535.104.12, NVIDIA A800-SXM4-80GB, 1321923000426, Disabled, Enabled, 42, 2024/03/08 13:49:22.066
5, GPU-67fc57fc-34ad-4126-2f66-0b8d29144c75, 100, 81920, 74745, 6483, 535.104.12, NVIDIA A800-SXM4-80GB, 1321923001319, Disabled, Enabled, 73, 2024/03/08 13:49:22.067
6, GPU-105ee81f-dddd-aaf0-2e30-ca1593fdbf18, 0, 81920, 2, 81226, 535.104.12, NVIDIA A800-SXM4-80GB, 1321923002833, Disabled, Enabled, 30, 2024/03/08 13:49:22.068
7, GPU-349fa89c-151d-340e-c147-94506daf1357, 0, 81920, 2, 81226, 535.104.12, NVIDIA A800-SXM4-80GB, 1321823066087, Disabled, Enabled, 32, 2024/03/08 13:49:22.069
`

var compute = `
2024/03/08 16:05:13.791, NVIDIA A800-SXM4-80GB, GPU-67fc57fc-34ad-4126-2f66-0b8d29144c75, 44141, /opt/miniconda/bin/python, 74736
`

func TestLoad(t *testing.T) {
	lines, err := parse([]byte(example))
	assert.Nil(t, err)
	assert.Len(t, lines, 8)

	for _, line := range lines {
		assert.Len(t, line, 13)
	}
}

func TestGetGPUs(t *testing.T) {
	patches := gomonkey.ApplyFunc(run, func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(example), nil
	})
	defer patches.Reset()

	ctx := context.Background()
	gpus, err := GetGPUs(ctx)
	assert.Nil(t, err)
	assert.Len(t, gpus, 8)
}

func TestGPU_String(t *testing.T) {
	g := GPU{
		Index:          "0",
		UUID:           "GPU-fd189414-e0f6-58a0-7031-fefe0ce43b1d",
		UtilizationGPU: "0",
		MemoryTotal:    "81920",
		MemoryUsed:     "2",
		MemoryFree:     "81226",
		DriverVersion:  "535.104.12",
		Name:           "NVIDIA A800-SXM4-80GB",
		Serial:         "1321923001828",
		DisplayActive:  "Disabled",
		DisplayMode:    "Enabled",
		Temperature:    "31",
		Timestamp:      "2024/03/08 13:49:49.053",
	}
	assert.Equal(
		t,
		"0, GPU-fd189414-e0f6-58a0-7031-fefe0ce43b1d, 0 %, 81920 MiB, 2 MiB, 81226 MiB, 535.104.12, NVIDIA A800-SXM4-80GB, 1321923001828, Disabled, Enabled, 31, 2024/03/08 13:49:49.053",
		g.String(),
	)
}

func TestGPUComputeApp_String(t *testing.T) {
	g := GPUComputeApp{
		Timestamp:  "2024/03/08 16:05:13.791",
		Name:       "NVIDIA A800-SXM4-80GB",
		UUID:       "GPU-67fc57fc-34ad-4126-2f66-0b8d29144c75",
		PID:        "44141",
		UsedMemory: "74736",
	}
	assert.Equal(
		t,
		"2024/03/08 16:05:13.791, NVIDIA A800-SXM4-80GB, GPU-67fc57fc-34ad-4126-2f66-0b8d29144c75, 44141, 74736 MiB",
		g.String(),
	)
}

func TestGetProcesses(t *testing.T) {
	patches := gomonkey.ApplyFunc(run, func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(compute), nil
	})
	defer patches.Reset()

	ctx := context.Background()
	processes, err := GetProcesses(ctx)
	assert.Nil(t, err)
	assert.Len(t, processes, 1)
}
