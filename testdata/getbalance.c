#include"tcapi.h"

char *thunderchain_main(char *action, char *args) {
  char* ret;
  char* addr = "0x0000000000000000000000000000000000000001";
  ret = TC_GetBalance(addr);
  return ret;
}

