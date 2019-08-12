#include"tcapi.h"

char *thunderchain_main(char *action, char *args) {
  BigInt amount = "10000";
  TC_Issue(amount);
  char* to=TC_GetMsgSender();
  TC_Prints(to);
  char* token = TC_GetSelfAddress();
  TC_Prints(token);
  TC_TransferToken(to, token, amount);
  BigInt balance = TC_TokenBalance(to, token);
  TC_Prints(balance);
  char* tokenAddress = TC_TokenAddress();
  TC_Prints(tokenAddress);
  return (char*)0;
}
