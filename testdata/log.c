#include"tcapi.h"

char *thunderchain_main(char *action, char *args) {
  char* ret;
  char* data = "0x0000000000000000000000000000000000000001";
  char* topic1 = "1000";
  char* topic2 = "2000";
  char* topic3 = "3000";
  char* topic4 = "4000";
  TC_Log0(data);
  TC_Log1(data, topic1);
  TC_Log2(data, topic1, topic2);
  TC_Log3(data, topic1, topic2, topic3);
  TC_Log4(data, topic1, topic2, topic3, topic4);
  return (char*)0;
}
