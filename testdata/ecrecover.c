#include"tcapi.h"

char *thunderchain_main(char *action, char *args) {
  char* ret;
  char* hash = "0xd3a0853868c512baceffe48f5ad143ac3dfc4de87718e3392b76698dd102ea2a";
  char* v = "0xec8e";
  char* r = "0x29307edb09cccad279479eebdb423712f1497b4b608436326204e970863ba618";
  char* s = "0x7a3455d15bb2f5c42c2dfccceea396ad82d9b6ad1038822f4e58147ca580eac7";
  ret = TC_Ecrecover(hash, v, r, s);
  return ret;
}
