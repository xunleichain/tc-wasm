#include"tcapi.h"

char *thunderchain_main(char *action, char *args) {
  char* ret;
  char* data = "0x0000000000000000000000000000000000000001";
  ret = TC_Keccak256(data);
  return ret;
}

