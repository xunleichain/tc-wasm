#include"tcapi.h"

char *thunderchain_main(char *action, char *args) {
  char* eventID="transfer";
  char* data="{\"from\":\"0x0d368fc017cf02dba3b32515079709e56f0f9346\",\"to\":\"0x6f2da8e5f6ce4648166a037abe4656d33e84032a\",\"value\":\"10000\"}";
  TC_Notify(eventID, data);
  return (char*)0;
}
