#ifndef __APPLE__
#ifndef __GPU_INFO_MUSA_H__
#define __GPU_INFO_MUSA_H__
#include "gpu_info.h"

// Just enough typedef's to dlopen/dlsym for memory information
typedef enum musaError_enum {
  MUSA_SUCCESS = 0,
  MUSA_ERROR_INVALID_VALUE = 1,
  MUSA_ERROR_OUT_OF_MEMORY = 2,
  MUSA_ERROR_NOT_INITIALIZED = 3,
  MUSA_ERROR_INSUFFICIENT_DRIVER = 35,
  MUSA_ERROR_NO_DEVICE = 100,
  MUSA_ERROR_SYSTEM_DRIVER_MISMATCH = 803,
  MUSA_ERROR_UNKNOWN = 999,
  // Other values omitted for now...
} MUresult;

typedef enum MUdevice_attribute_enum {
  MU_DEVICE_ATTRIBUTE_COMPUTE_CAPABILITY_MAJOR = 75,
  MU_DEVICE_ATTRIBUTE_COMPUTE_CAPABILITY_MINOR = 76,

  // TODO - not yet wired up but may be useful for Jetson or other
  // integrated GPU scenarios with shared memory
  MU_DEVICE_ATTRIBUTE_INTEGRATED = 18

} MUdevice_attribute;

typedef void *musaDevice_t;  // Opaque is sufficient
typedef struct musaMemory_st {
  uint64_t total;
  uint64_t free;
} musaMemory_t;

typedef struct musaDriverVersion {
  int major;
  int minor;
} musaDriverVersion_t;

typedef struct MUuuid_st {
    unsigned char bytes[16];
} MUuuid;

typedef int MUdevice;
typedef void* MUcontext;

typedef struct musa_handle {
  void *handle;
  uint16_t verbose;
  int driver_major;
  int driver_minor;
  MUresult (*muInit)(unsigned int Flags);
  MUresult (*muDriverGetVersion)(int *driverVersion);
  MUresult (*muDeviceGetCount)(int *);
  MUresult (*muDeviceGet)(MUdevice* device, int ordinal);
  MUresult (*muDeviceGetAttribute)(int* pi, MUdevice_attribute attrib, MUdevice dev);
  MUresult (*muDeviceGetUuid_v2)(MUuuid* uuid, MUdevice dev);
  MUresult (*muDeviceGetName)(char *name, int len, MUdevice dev);

  // Context specific aspects
  MUresult (*muCtxCreate_v2)(MUcontext* pctx, void *params, int len, unsigned int flags, MUdevice dev);
  MUresult (*muMemGetInfo_v2)(uint64_t* free, uint64_t* total);
  MUresult (*muCtxDestroy_v2)(MUcontext ctx);
} musa_handle_t;

typedef struct musa_init_resp {
  char *err;  // If err is non-null handle is invalid
  musa_handle_t ch;
  int num_devices;
  MUresult musaErr;
} musa_init_resp_t;

void musa_init(char *musa_lib_path, musa_init_resp_t *resp);
void musa_bootstrap(musa_handle_t ch, int device_id, mem_info_t *resp);
void musa_get_free(musa_handle_t ch,  int device_id, uint64_t *free, uint64_t *total);
void musa_release(musa_handle_t ch);

#endif  // __GPU_INFO_MUSA_H__
#endif  // __APPLE__
