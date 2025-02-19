#include <string.h>
#include "gpu_info.h"
#include <glob.h>

static char* vendor_name = "";

const char* to_vendor_sym(const char *input) {
  glob_t result;
  if (strcmp(vendor_name, "musa") == 0
    || glob("/usr/local/musa*/lib*/libmusa.so*", GLOB_TILDE, NULL, &result) == 0
  ) {
    vendor_name = "musa";

    if (strcmp(input, "cudaSetDevice") == 0) {
      return "musaSetDevice";
    } else if (strcmp(input, "cudaDeviceSynchronize") == 0) {
      return "musaDeviceSynchronize";
    } else if (strcmp(input, "cudaDeviceReset") == 0) {
      return "musaDeviceReset";
    } else if (strcmp(input, "cudaMemGetInfo") == 0) {
      return "musaMemGetInfo";
    } else if (strcmp(input, "cudaGetDeviceCount") == 0) {
      return "musaGetDeviceCount";
    } else if (strcmp(input, "cudaDeviceGetAttribute") == 0) {
      return "musaDeviceGetAttribute";
    } else if (strcmp(input, "cudaDriverGetVersion") == 0) {
      return "musaDriverGetVersion";
    } else if (strcmp(input, "cudaGetDeviceProperties") == 0) {
      return "musaGetDeviceProperties";
    } else if (strcmp(input, "cuInit") == 0) {
      return "muInit";
    } else if (strcmp(input, "cuDriverGetVersion") == 0) {
      return "muDriverGetVersion";
    } else if (strcmp(input, "cuDeviceGetCount") == 0) {
      return "muDeviceGetCount";
    } else if (strcmp(input, "cuDeviceGet") == 0) {
      return "muDeviceGet";
    } else if (strcmp(input, "cuDeviceGetAttribute") == 0) {
      return "muDeviceGetAttribute";
    } else if (strcmp(input, "cuDeviceGetUuid") == 0) {
      return "muDeviceGetUuid_v2";
    } else if (strcmp(input, "cuDeviceGetName") == 0) {
      return "muDeviceGetName";
    } else if (strcmp(input, "cuCtxCreate_v3") == 0) {
      return "muCtxCreate_v2";
    } else if (strcmp(input, "cuMemGetInfo_v2") == 0) {
      return "muMemGetInfo_v2";
    } else if (strcmp(input, "cuCtxDestroy") == 0) {
      return "muCtxDestroy_v2";
    }
  }

  return input;
}
