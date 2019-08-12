//contract2
#include"tcapi.h"

void c2TestFunc1() {
	TC_Prints("=======begin c2TestFunc1 in c2==========");
	TC_Prints("TC_StorageSet(\"c2TestFunc1\", \"this is set in c2TestFunc1\")");
	TC_StorageSet("c2TestFunc1", "this is set in c2TestFunc1");
	TC_Prints(strconcat("contract address: ", TC_GetSelfAddress()));
	TC_Prints(strconcat("caller address: ", TC_GetMsgSender()));
	TC_Prints(strconcat("msg data: ", TC_GetMsgData()));
	TC_Prints(strconcat("msg sign: ", TC_GetMsgSign()));
	TC_Prints(strconcat("msg value: ", TC_GetMsgValue()));
	TC_Prints(strconcat("msg token value: ", TC_GetMsgTokenValue()));
	TC_Prints("=======end c2TestFunc1 in c2==========");
}

void c2TestFunc2() {
	TC_Prints("=======begin c2TestFunc2 in c2==========");
	TC_Prints("TC_StorageSet(\"c2TestFunc2\", \"this is set in c2TestFunc2\")");
	TC_StorageSet("c2TestFunc2", "this is set in c2TestFunc2");
	TC_Prints(strconcat("contract address: ", TC_GetSelfAddress()));
	TC_Prints(strconcat("caller address: ", TC_GetMsgSender()));
	TC_Prints(strconcat("msg data: ", TC_GetMsgData()));
	TC_Prints(strconcat("msg sign: ", TC_GetMsgSign()));
	TC_Prints(strconcat("msg value: ", TC_GetMsgValue()));
	TC_Prints(strconcat("msg token value: ", TC_GetMsgTokenValue()));
	TC_Prints("=======end c2TestFunc2 in c2==========");
}


void c2TestFunc3() {
	TC_Prints("=======begin c2TestFunc3 in c2==========");
	TC_Prints("TC_StorageSet(\"c2TestFunc3\", \"this is set in c2TestFunc3\")");
	TC_StorageSet("c2TestFunc3", "this is set in c2TestFunc3");
	TC_Prints(strconcat("contract address: ", TC_GetSelfAddress()));
	TC_Prints(strconcat("caller address: ", TC_GetMsgSender()));
	TC_Prints(strconcat("msg data: ", TC_GetMsgData()));
	TC_Prints(strconcat("msg sign: ", TC_GetMsgSign()));
	TC_Prints(strconcat("msg value: ", TC_GetMsgValue()));
	TC_Prints(strconcat("msg token value: ", TC_GetMsgTokenValue()));
	TC_Prints("=======end c2TestFunc3 in c2==========");
}


void c2TestFunc4() {
	TC_Prints("=======begin c2TestFunc4 in c2==========");
	TC_Prints("TC_StorageSet(\"c2TestFunc4\", \"this is set in c2TestFunc4\")");
	TC_StorageSet("c2TestFunc4", "this is set in c2TestFunc4");
	TC_Prints(strconcat("contract address: ", TC_GetSelfAddress()));
	TC_Prints(strconcat("caller address: ", TC_GetMsgSender()));
	TC_Prints(strconcat("msg data: ", TC_GetMsgData()));
	TC_Prints(strconcat("msg sign: ", TC_GetMsgSign()));
	TC_Prints(strconcat("msg value: ", TC_GetMsgValue()));
	TC_Prints(strconcat("msg token value: ", TC_GetMsgTokenValue()));
	TC_Prints("=======end c2TestFunc4 in c2==========");
}

char *thunderchain_main(char *action, char *args) {
  if(0 == strcmp(action,"Init")) {
	  return "Init ok";
  }
  if(0 == strcmp(action, "c2TestFunc1")) {
  	c2TestFunc1();
  	char* shouldGetStr = TC_StorageGet("c2TestFunc1");
  	TC_Prints(strconcat("TC_StorageGet(\"c2TestFunc1\"): ",shouldGetStr));
  	if(0 == strcmp("this is set in c2TestFunc1", shouldGetStr)){
  	  	TC_Prints("===============================================> test Call + Call in c2 verify succ");
  	}else{
  	  	TC_Prints("===============================================> test Call + Call in c2 verify fail");
  	}
  	return (char*)0;
  }
  if(0 == strcmp(action, "c2TestFunc2")) {
  	c2TestFunc2();
  	char* shouldGetStr = TC_StorageGet("c2TestFunc2"); //because context not change
  	TC_Prints(strconcat("TC_StorageGet(\"c2TestFunc2\"): ",shouldGetStr));
  	if(0 == strcmp("this is set in c2TestFunc2", shouldGetStr)){
  		TC_Prints("===============================================> test Call + delegatecall in c2 verify succ");
  	}else{
  	  	TC_Prints("===============================================> test Call + delegatecall in c2 verify fail");
  	}
  	return (char*)0;
  }
  if(0 == strcmp(action, "c2TestFunc3")) {
  	c2TestFunc3();
  	char* shouldGetStr = TC_StorageGet("c2TestFunc3");
  	TC_Prints(strconcat("TC_StorageGet(\"c2TestFunc3\"): ",shouldGetStr));
  	if(0 == strcmp("this is set in c2TestFunc3", shouldGetStr)){
  		TC_Prints("===============================================> test delegateCall + Call in c2 verify succ");
  	}else{
  	  	TC_Prints("===============================================> test delegateCall + Call in c2 verify fail");
  	}
  	return (char*)0;
  }
  if(0 == strcmp(action, "c2TestFunc4")) {
  	c2TestFunc4();
  	char* shouldGetStr = TC_StorageGet("c2TestFunc4"); //because context not change
  	TC_Prints(strconcat("TC_StorageGet(\"c2TestFunc4\"): ",shouldGetStr));
  	if(0 == strcmp("this is set in c2TestFunc4", shouldGetStr)){
  	   TC_Prints("===============================================> test delegateCall + delegateCall in c2 verify succ");
  	}else{
  	   TC_Prints("===============================================> test delegateCall + delegateCall in c2 verify fail");
  	}
  	return (char*)0;
  }
  TC_Prints("func not called in contract 2");
  return (char*)0;
}

