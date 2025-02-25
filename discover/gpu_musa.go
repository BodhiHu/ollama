package discover

import "C"
import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/ollama/ollama/format"
)

type musaHandles struct {
	deviceCount int
	musart      *C.musart_handle_t
	musa        *C.musa_handle_t
}

func MUSARefreshFreeMemory(mHandles *musaHandles, musaGPUs []MusaGPUInfo) {
	var memInfo C.mem_info_t

	for i, gpu := range musaGPUs {
		if mHandles.musart != nil {
			C.musart_bootstrap(*mHandles.musart, C.int(gpu.index), &memInfo)
		} else if mHandles.musa != nil {
			C.musa_get_free(*mHandles.musa, C.int(gpu.index), &memInfo.free, &memInfo.total)
			memInfo.used = memInfo.total - memInfo.free
		} else {
			// shouldn't happen
			slog.Warn("no valid musa library loaded to refresh vram usage")
			break
		}
		if memInfo.err != nil {
			slog.Warn("error looking up mthreads GPU memory", "error", C.GoString(memInfo.err))
			C.free(unsafe.Pointer(memInfo.err))
			continue
		}
		if memInfo.free == 0 {
			slog.Warn("error looking up mthreads GPU memory")
			continue
		}
		slog.Debug("updating musa memory data",
			"gpu", gpu.ID,
			"name", gpu.Name,
			"overhead", format.HumanBytes2(gpu.OSOverhead),
			slog.Group(
				"before",
				"total", format.HumanBytes2(gpu.TotalMemory),
				"free", format.HumanBytes2(gpu.FreeMemory),
			),
			slog.Group(
				"now",
				"total", format.HumanBytes2(uint64(memInfo.total)),
				"free", format.HumanBytes2(uint64(memInfo.free)),
				"used", format.HumanBytes2(uint64(memInfo.used)),
			),
		)
		musaGPUs[i].FreeMemory = uint64(memInfo.free)
	}
}

func MUSAGetGPUInfo() ([]MusaGPUInfo, error) {
	var memInfo C.mem_info_t
	mHandles := initMusaHandles()
	var musaGPUs []MusaGPUInfo

	for i := range mHandles.deviceCount {
		if mHandles.musart != nil || mHandles.musa != nil {
			gpuInfo := MusaGPUInfo{
				GpuInfo: GpuInfo{
					Library: "musa",
				},
				index: i,
			}
			var driverMajor int
			var driverMinor int
			if mHandles.musart != nil {
				C.musart_bootstrap(*mHandles.musart, C.int(i), &memInfo)
			} else {
				C.musa_bootstrap(*mHandles.musa, C.int(i), &memInfo)
				driverMajor = int(mHandles.musa.driver_major)
				driverMinor = int(mHandles.musa.driver_minor)
			}
			if memInfo.err != nil {
				slog.Info("error looking up MUSA GPU memory", "error", C.GoString(memInfo.err))
				C.free(unsafe.Pointer(memInfo.err))
				continue
			}
			gpuInfo.TotalMemory = uint64(memInfo.total)
			gpuInfo.FreeMemory = uint64(memInfo.free)
			gpuInfo.ID = C.GoString(&memInfo.gpu_id[0])
			gpuInfo.Compute = fmt.Sprintf("%d.%d", memInfo.major, memInfo.minor)
			gpuInfo.computeMajor = int(memInfo.major)
			gpuInfo.computeMinor = int(memInfo.minor)
			gpuInfo.MinimumMemory = musaMinimumMemory
			gpuInfo.DriverMajor = driverMajor
			gpuInfo.DriverMinor = driverMinor

			gpuInfo.Name = C.GoString(&memInfo.gpu_name[0])

			musaGPUs = append(musaGPUs, gpuInfo)
		}
	}

	return musaGPUs, nil
}

func initMusaHandles() *musaHandles {

	mHandles := &musaHandles{}

	if musaLibPath != "" {
		mHandles.deviceCount, mHandles.musa, _, _ = loadMUSAMgmt([]string{musaLibPath})
		return mHandles
	}
	if musartLibPath != "" {
		mHandles.deviceCount, mHandles.musart, _, _ = loadMUSARTMgmt([]string{musartLibPath})
		return mHandles
	}

	slog.Debug("searching for GPU discovery libraries for MUSA")
	var musartMgmtPatterns []string

	// Aligned with driver, we can't carry as payloads
	musaMgmtPatterns := MusaGlobs
	musartMgmtPatterns = append(musartMgmtPatterns, filepath.Join(LibOllamaPath, "musa_v*", MusartMgmtName))
	musartMgmtPatterns = append(musartMgmtPatterns, MusartGlobs...)

	musaLibPaths := FindGPULibs(MusaMgmtName, musaMgmtPatterns)
	if len(musaLibPaths) > 0 {
		deviceCount, musa, libPath, err := loadMUSAMgmt(musaLibPaths)
		if musa != nil {
			slog.Debug("detected GPUs", "count", deviceCount, "library", libPath)
			mHandles.musa = musa
			mHandles.deviceCount = deviceCount
			musaLibPath = libPath
			return mHandles
		}
		if err != nil {
			bootstrapErrors = append(bootstrapErrors, err)
		}
	}

	musartLibPaths := FindGPULibs(MusartMgmtName, musartMgmtPatterns)
	if len(musartLibPaths) > 0 {
		deviceCount, musart, libPath, err := loadMUSARTMgmt(musartLibPaths)
		if musart != nil {
			slog.Debug("detected GPUs", "library", libPath, "count", deviceCount)
			mHandles.musart = musart
			mHandles.deviceCount = deviceCount
			musartLibPath = libPath
			return mHandles
		}
		if err != nil {
			bootstrapErrors = append(bootstrapErrors, err)
		}
	}

	return mHandles
}

func loadMUSARTMgmt(musartLibPaths []string) (int, *C.musart_handle_t, string, error) {
	var resp C.musart_init_resp_t
	resp.ch.verbose = getVerboseState()
	var err error
	for _, libPath := range musartLibPaths {
		lib := C.CString(libPath)
		defer C.free(unsafe.Pointer(lib))
		C.musart_init(lib, &resp)
		if resp.err != nil {
			err = fmt.Errorf("Unable to load musart library %s: %s", libPath, C.GoString(resp.err))
			slog.Debug(err.Error())
			C.free(unsafe.Pointer(resp.err))
		} else {
			err = nil
			return int(resp.num_devices), &resp.ch, libPath, err
		}
	}
	return 0, nil, "", err
}

// Bootstrap the driver library
// Returns: num devices, handle, libPath, error
func loadMUSAMgmt(musaLibPaths []string) (int, *C.musa_handle_t, string, error) {
	var resp C.musa_init_resp_t
	resp.ch.verbose = getVerboseState()
	var err error
	for _, libPath := range musaLibPaths {
		lib := C.CString(libPath)
		defer C.free(unsafe.Pointer(lib))
		C.musa_init(lib, &resp)
		if resp.err != nil {
			// Decide what log level based on the type of error message to help users understand why
			switch resp.musaErr {
			case C.MUSA_ERROR_INSUFFICIENT_DRIVER, C.MUSA_ERROR_SYSTEM_DRIVER_MISMATCH:
				err = fmt.Errorf("version mismatch between driver and musa driver library - reboot or upgrade may be required: library %s", libPath)
				slog.Warn(err.Error())
			case C.MUSA_ERROR_NO_DEVICE:
				err = fmt.Errorf("no MUSA devices detected by library %s", libPath)
				slog.Info(err.Error())
			case C.MUSA_ERROR_UNKNOWN:
				err = fmt.Errorf("unknown error initializing musa driver library %s: %s. see https://github.com/ollama/ollama/blob/main/docs/troubleshooting.md for more information", libPath, C.GoString(resp.err))
				slog.Warn(err.Error())
			default:
				msg := C.GoString(resp.err)
				if strings.Contains(msg, "wrong ELF class") {
					slog.Debug("skipping 32bit library", "library", libPath)
				} else {
					err = fmt.Errorf("Unable to load musart library %s: %s", libPath, C.GoString(resp.err))
					slog.Info(err.Error())
				}
			}
			C.free(unsafe.Pointer(resp.err))
		} else {
			err = nil
			return int(resp.num_devices), &resp.ch, libPath, err
		}
	}
	return 0, nil, "", err
}
