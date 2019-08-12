#include"tcapi.h"

char *thunderchain_main(char *action, char *args) {
  char* ret;
  char* to = "0x0000000000000000000000000000000000000001";
  ret = TC_SelfDestruct(to);
  return ret;
}

