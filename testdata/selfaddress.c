#include"tcapi.h"

char *thunderchain_main(char *action, char *args) {
  char* ret;
  ret = TC_GetSelfAddress();
  return ret;
}
