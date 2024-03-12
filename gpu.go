package gputil

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// the binary should be executable
const binary = "nvidia-smi"

const (
	queryInfo    = "--query-gpu=index,uuid,utilization.gpu,memory.total,memory.used,memory.free,driver_version,name,gpu_serial,power.draw,power.limit,temperature.gpu,timestamp"
	queryProcess = "--query-compute-apps=timestamp,gpu_name,gpu_uuid,pid,name,used_memory"
	queryFormat  = "--format=csv,noheader,nounits"
)

// GPU information
type GPU struct {
	// Zero based index of the GPU. Can change at each boot.
	Index string `json:"index"`
	// This value is the globally unique immutable alphanumeric identifier of the GPU.
	// It does not correspond to any physical label on the board.
	UUID string `json:"uuid"`
	// Percent of time over the past sample period during which one or more kernels was executing on the GPU.
	// The sample period may be between 1 second and 1/6 second depending on the product.
	UtilizationGPU string `json:"utilizationGPU"`
	// Total installed GPU memory. units, MiB
	MemoryTotal string `json:"memoryTotal"`
	// Total memory allocated by active contexts. units, MiB
	MemoryUsed string `json:"memoryUsed"`
	// Total free memory. units, MiB
	MemoryFree string `json:"memoryFree"`
	// The version of the installed NVIDIA display driver.
	// This is an alphanumeric string.
	DriverVersion string `json:"driverVersion"`
	// The official product name of the GPU. This is an alphanumeric string
	Name string `json:"name"`
	// This number matches the serial number physically printed on each board. It is a globally unique immutable alphanumeric value.
	Serial string `json:"serial"`
	// The last measured power draw for the entire board, in watts.
	// On Ampere or newer devices, returns average power draw over 1 sec.
	// On older devices, returns instantaneous power draw. Only available if power management is supported.
	// This reading is accurate to within +/- 5 watts.
	PowerDraw string `json:"powerDraw"`
	// The software power limit in watts. Set by software like nvidia-smi.
	// On Kepler devices Power Limit can be adjusted using [-pl | --power-limit=] switches.
	PowerLimit string `json:"powerLimit"`
	// Core GPU temperature. in degrees C.
	Temperature string `json:"temperature"`
	// The timestamp of when the query was made in format "YYYY/MM/DD HH:MM:SS.msec".
	Timestamp string `json:"timestamp"`
}

// GPUComputeApp processes having compute context on the device.
type GPUComputeApp struct {
	// The timestamp of when the query was made in format "YYYY/MM/DD HH:MM:SS.msec".
	Timestamp string `json:"timestamp"`
	// The official product name of the GPU.
	// This is an alphanumeric string. For all products.
	Name string `json:"name"`
	// This value is the globally unique immutable alphanumeric identifier of the GPU.
	// It does not correspond to any physical label on the board.
	UUID string `json:"uuid"`
	// Process ID of the compute application
	PID string `json:"pid"`
	// Process Name
	ProcessName string `json:"processName"`
	// Amount memory used on the device by the context.
	// Not available on Windows when running in WDDM mode because Windows KMD manages all the memory not NVIDIA driver.
	UsedMemory string `json:"usedMemory"`
}

// GetGPUs returns all GPUs or specified index/uuids information
func GetGPUs(ctx context.Context, indexOrUUIDs ...string) (result []GPU, err error) {
	var rsp []byte
	var args = []string{queryInfo, queryFormat}
	if len(indexOrUUIDs) > 0 {
		args = append(args, fmt.Sprintf("-i %s", strings.Join(indexOrUUIDs, ",")))
	}
	if rsp, err = run(ctx, args...); err != nil {
		return
	}
	result, err = composeGPUInfoLines(rsp)
	return
}

// GetProcesses returns processes having compute context on the device
// Note: if no processes running, empty result return
func GetProcesses(ctx context.Context, indexOrUUIDs ...string) (result []GPUComputeApp, err error) {
	var rsp []byte
	var args = []string{queryProcess, queryFormat}
	if len(indexOrUUIDs) > 0 {
		args = append(args, fmt.Sprintf("-i %s", strings.Join(indexOrUUIDs, ",")))
	}
	if rsp, err = run(ctx, args...); err != nil {
		return
	}
	result, err = composeProcessInfoLines(rsp)
	return
}

func composeProcessInfoLines(rsp []byte) (result []GPUComputeApp, err error) {
	var lines [][]string
	if lines, err = parse(rsp); err != nil {
		return
	}
	result = make([]GPUComputeApp, 0, len(lines))
	for _, line := range lines {
		result = append(result, GPUComputeApp{
			Timestamp:   sanitize(line[0]),
			Name:        sanitize(line[1]),
			UUID:        sanitize(line[2]),
			PID:         sanitize(line[3]),
			ProcessName: sanitize(line[4]),
			UsedMemory:  sanitize(line[5]),
		})
	}
	return
}

func run(ctx context.Context, args ...string) (rsp []byte, err error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	if rsp, err = cmd.Output(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			err = ee
			return
		}
		return
	}
	return
}

// compose csv lines to GPU instance
func composeGPUInfoLines(rsp []byte) (result []GPU, err error) {
	var lines [][]string
	if lines, err = parse(rsp); err != nil {
		return
	}
	result = make([]GPU, 0, len(lines))
	for _, line := range lines {
		result = append(result, GPU{
			Index:          sanitize(line[0]),
			UUID:           sanitize(line[1]),
			UtilizationGPU: sanitize(line[2]),
			MemoryTotal:    sanitize(line[3]),
			MemoryUsed:     sanitize(line[4]),
			MemoryFree:     sanitize(line[5]),
			DriverVersion:  sanitize(line[6]),
			Name:           sanitize(line[7]),
			Serial:         sanitize(line[8]),
			PowerDraw:      sanitize(line[9]),
			PowerLimit:     sanitize(line[10]),
			Temperature:    sanitize(line[11]),
			Timestamp:      sanitize(line[12]),
		})
	}
	return
}

// parse csv lines
func parse(content []byte) (lines [][]string, err error) {
	r := csv.NewReader(bytes.NewReader(content))
	for {
		row, e := r.Read()
		if e != nil {
			if errors.Is(e, io.EOF) {
				break
			}
			err = e
			break
		}
		lines = append(lines, row)
	}
	return
}

func (g *GPU) String() string {
	return fmt.Sprintf(
		"%s, %s, %s %%, %s MiB, %s MiB, %s MiB, %s, %s, %s, %s W, %s W, %s, %s",
		g.Index,
		g.UUID,
		g.UtilizationGPU,
		g.MemoryTotal,
		g.MemoryUsed,
		g.MemoryFree,
		g.DriverVersion,
		g.Name,
		g.Serial,
		g.PowerDraw,
		g.PowerLimit,
		g.Temperature,
		g.Timestamp,
	)
}

func (c *GPUComputeApp) String() string {
	return fmt.Sprintf("%s, %s, %s, %s, %s MiB", c.Timestamp, c.Name, c.UUID, c.PID, c.UsedMemory)
}

func sanitize(input string) string {
	return strings.TrimSpace(input)
}
