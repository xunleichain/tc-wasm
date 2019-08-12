#include"tcapi.h"

char *thunderchain_main(char *action, char *args) {
  char* data="hello world!";
  int ret = TC_Strlen(data);
  return (char*)ret;
}

