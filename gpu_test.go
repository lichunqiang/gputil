package gputil

import (
	"context"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

// timestamp, index, uuid, utilization.gpu [%], memory.total [MiB], memory.used [MiB], memory.free [MiB], driver_version, name, serial, power.draw, power.limit, temperature.gpu
var example = `
0, GPU-fd189414-e0f6-58a0-7031-fefe0ce43b1d, 0 %, 81920 MiB, 2 MiB, 81226 MiB, 535.104.12, NVIDIA A800-SXM4-80GB, 1321923001828, 61.89 W, 400.00 W, 33, 2024/03/12 17:48:46.990
1, GPU-121ebc1f-3e5d-139d-7aac-57311d5bafc7, 0 %, 81920 MiB, 2 MiB, 81226 MiB, 535.104.12, NVIDIA A800-SXM4-80GB, 1321923003018, 59.46 W, 400.00 W, 31, 2024/03/12 17:48:46.993
2, GPU-b74d1aeb-0aab-b3ca-ff55-94e24cbe0cd6, 0 %, 81920 MiB, 2 MiB, 81226 MiB, 535.104.12, NVIDIA A800-SXM4-80GB, 1321423018183, 60.74 W, 400.00 W, 30, 2024/03/12 17:48:46.997
3, GPU-401e53f2-8f44-1fc5-469d-8e36c1d6c9c5, 0 %, 81920 MiB, 2 MiB, 81226 MiB, 535.104.12, NVIDIA A800-SXM4-80GB, 1321923002415, 59.55 W, 400.00 W, 31, 2024/03/12 17:48:47.000
4, GPU-8b63b1f2-98e1-b24e-f59f-d725d51b3a2b, 0 %, 81920 MiB, 2 MiB, 81226 MiB, 535.104.12, NVIDIA A800-SXM4-80GB, 1321923000426, 62.17 W, 400.00 W, 32, 2024/03/12 17:48:47.003
5, GPU-67fc57fc-34ad-4126-2f66-0b8d29144c75, 0 %, 81920 MiB, 2 MiB, 81226 MiB, 535.104.12, NVIDIA A800-SXM4-80GB, 1321923001319, 61.95 W, 400.00 W, 32, 2024/03/12 17:48:47.005
6, GPU-105ee81f-dddd-aaf0-2e30-ca1593fdbf18, 100 %, 81920 MiB, 74829 MiB, 6399 MiB, 535.104.12, NVIDIA A800-SXM4-80GB, 1321923002833, 359.75 W, 400.00 W, 70, 2024/03/12 17:48:47.008
7, GPU-349fa89c-151d-340e-c147-94506daf1357, 0 %, 81920 MiB, 2 MiB, 81226 MiB, 535.104.12, NVIDIA A800-SXM4-80GB, 1321823066087, 71.08 W, 400.00 W, 49, 2024/03/12 17:48:47.011
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
		PowerDraw:      "359.75",
		PowerLimit:     "400.00",
		Temperature:    "31",
		Timestamp:      "2024/03/08 13:49:49.053",
	}
	assert.Equal(
		t,
		"0, GPU-fd189414-e0f6-58a0-7031-fefe0ce43b1d, 0 %, 81920 MiB, 2 MiB, 81226 MiB, 535.104.12, NVIDIA A800-SXM4-80GB, 1321923001828, 359.75 W, 400.00 W, 31, 2024/03/08 13:49:49.053",
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
