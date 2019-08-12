#include"tcapi.h"

char *thunderchain_main(char *action, char *args) {
  char* to_addr = "0x0000000000000000000000000000000000000001";
  char* amount = "125";
  TC_Transfer(to_addr, amount);
  return (char*)0;
}
