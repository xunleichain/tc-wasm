//contract1
#include"tcapi.h"

void testCall() {
	TC_Prints("=======begin testCall in c1==========");
	TC_Prints("TC_StorageSet(\"testCall\", \"this is set in c1 testCall\")");
	TC_StorageSet("testCall", "this is set in c1 testCall");
	TC_Prints(strconcat("contract address: ", TC_GetSelfAddress()));
	TC_Prints(strconcat("caller address: ", TC_GetMsgSender()));
	TC_Prints(strconcat("msg data: ", TC_GetMsgData()));
	TC_Prints(strconcat("msg sign: ", TC_GetMsgSign()));
	TC_Prints(strconcat("msg value: ", TC_GetMsgValue()));
	TC_Prints(strconcat("msg token value: ", TC_GetMsgTokenValue()));
	TC_Prints("=======end testCall in c1==========");
}

void testDelegateCall() {
	TC_Prints("=======begin testDelegateCall in c1==========");
	TC_Prints("TC_StorageSet(\"testDelegateCall\", \"this is set in c1 testDelegateCall\")");
	TC_StorageSet("testDelegateCall", "this is set in c1 testDelegateCall");
	TC_Prints(strconcat("contract address: ", TC_GetSelfAddress()));
	TC_Prints(strconcat("caller address: ", TC_GetMsgSender()));
	TC_Prints(strconcat("msg data: ", TC_GetMsgData()));
	TC_Prints(strconcat("msg sign: ", TC_GetMsgSign()));
	TC_Prints(strconcat("msg value: ", TC_GetMsgValue()));
	TC_Prints(strconcat("msg token value: ", TC_GetMsgTokenValue()));
	TC_Prints("=======end testDelegateCall in c1==========");
}

void c1CallCall(char* contract2) {
	TC_Prints("=======begin c1CallCall in c1==========");
	TC_Prints(strconcat("contract address: ", TC_GetSelfAddress()));
	TC_Prints(strconcat("caller address: ", TC_GetMsgSender()));
	TC_Prints(strconcat("msg data: ", TC_GetMsgData()));
	TC_Prints(strconcat("msg sign: ", TC_GetMsgSign()));
	TC_Prints(strconcat("msg value: ", TC_GetMsgValue()));
	TC_Prints(strconcat("msg token value: ", TC_GetMsgTokenValue()));
	TC_CallContract(contract2, "c2TestFunc1", "{\"key_c2TestFunc1\":\"value_c2TestFunc1\"}");
	TC_Prints("=======end c1CallCall in c1==========");
}

void c1CallDelegateCall(char* contract2) {
	TC_Prints("=======begin c1CallDelegateCall in c1==========");
	TC_Prints(strconcat("contract address: ", TC_GetSelfAddress()));
	TC_Prints(strconcat("caller address: ", TC_GetMsgSender()));
	TC_Prints(strconcat("msg data: ", TC_GetMsgData()));
	TC_Prints(strconcat("msg sign: ", TC_GetMsgSign()));
	TC_Prints(strconcat("msg value: ", TC_GetMsgValue()));
	TC_Prints(strconcat("msg token value: ", TC_GetMsgTokenValue()));
	TC_DelegateCallContract(contract2, "c2TestFunc2", "{\"key_c2TestFunc2\":\"value_c2TestFunc2\"}");
	TC_Prints("=======end c1CallDelegateCall in c1==========");
}

void c1DelegateCallCall(char* contract2) {
	TC_Prints("=======begin c1DelegateCallCall in c1==========");
	TC_Prints(strconcat("contract address: ", TC_GetSelfAddress()));
	TC_Prints(strconcat("caller address: ", TC_GetMsgSender()));
	TC_Prints(strconcat("msg data: ", TC_GetMsgData()));
	TC_Prints(strconcat("msg sign: ", TC_GetMsgSign()));
	TC_Prints(strconcat("msg value: ", TC_GetMsgValue()));
	TC_Prints(strconcat("msg token value: ", TC_GetMsgTokenValue()));
	TC_CallContract(contract2, "c2TestFunc3", "{\"key_c2TestFunc3\":\"value_c2TestFunc3\"}");
	TC_Prints("=======end c1DelegateCallCall in c1==========");
}

void c1DelegateCallDelegateCall(char* contract2) {
	TC_Prints("=======begin c1DelegateCallDelegateCall in c1==========");
	TC_Prints(strconcat("contract address: ", TC_GetSelfAddress()));
	TC_Prints(strconcat("caller address: ", TC_GetMsgSender()));
	TC_Prints(strconcat("msg data: ", TC_GetMsgData()));
	TC_Prints(strconcat("msg sign: ", TC_GetMsgSign()));
	TC_Prints(strconcat("msg value: ", TC_GetMsgValue()));
	TC_Prints(strconcat("msg token value: ", TC_GetMsgTokenValue()));
	TC_DelegateCallContract(contract2, "c2TestFunc4",  "{\"key_c2TestFunc4\":\"value_c2TestFunc4\"}");
	TC_Prints("=======end c1DelegateCallDelegateCall in c1==========");
}


char *thunderchain_main(char *action, char *args) {
  if(0 == strcmp(action, "Init")){
	  return "Init ok";
  }
  if(0 == strcmp(action, "testCall")) {
  	testCall();
  	char* shouldGetStr = TC_StorageGet("testCall");
  	TC_Prints(strconcat("TC_StorageGet(\"testCall\"): ",shouldGetStr));
  	if(0 == strcmp("this is set in c1 testCall", shouldGetStr)){
  		TC_Prints("===============================================> test Call in c1 verify succ");
  	}else{
  		TC_Prints("===============================================> test Call in c1 verify fail");
  	}
  	return (char*)0;
  }
  if(0 == strcmp(action, "testDelegateCall")) {
  	testDelegateCall();
  	char* shouldGetStr = TC_StorageGet("testDelegateCall"); //because context not change
  	TC_Prints(strconcat("TC_StorageGet(\"testDelegateCall\"): ",shouldGetStr));
  	if(0 == strcmp("this is set in c1 testDelegateCall", shouldGetStr)){
  	    TC_Prints("===============================================> test delegateCall in c1 verify succ");
  	}else{
  	  	TC_Prints("===============================================> test delegateCall in c1 verify fail");
  	}
  	return (char*)0;
  }
  if(0 == strcmp(action, "c1CallCall")) {
  	c1CallCall(args);
  	char* shouldNotGetStr = TC_StorageGet("c2TestFunc1");
  	TC_Prints(strconcat("TC_StorageGet(\"c2TestFunc1\"): ",shouldNotGetStr));
  	if(0 != strcmp("this is set in c2TestFunc1", shouldNotGetStr)){
  	  	TC_Prints("===============================================> test Call + Call in c1 verify succ");
  	}else{
  	  	TC_Prints("===============================================> test Call + Call in c1 verify fail");
  	}
  	return (char*)0;
  }
  if(0 == strcmp(action, "c1CallDelegateCall")) {
  	c1CallDelegateCall(args);
  	char* shouldGetStr = TC_StorageGet("c2TestFunc2");
  	TC_Prints(strconcat("TC_StorageGet(\"c2TestFunc2\"): ",shouldGetStr));
  	if(0 == strcmp("this is set in c2TestFunc2", shouldGetStr)){
  	  	TC_Prints("===============================================> test Call + delegetecall in c1 verify succ");
  	}else{
  	  	TC_Prints("===============================================> test Call + delegetecall in c1 verify fail");
  	}
  	return (char*)0;
  }
  if(0 == strcmp(action, "c1DelegateCallCall")) {
  	c1DelegateCallCall(args);
  	char* shouldNotGetStr = TC_StorageGet("c2TestFunc3");
  	TC_Prints(strconcat("TC_StorageGet(\"c2TestFunc3\"): ",shouldNotGetStr));
  	if(0 != strcmp("this is set in c2TestFunc3", shouldNotGetStr)){
  	  	TC_Prints("===============================================> test delegateCall + Call in c1 verify succ");
  	}else{
  	  	TC_Prints("===============================================> test delegateCall + Call in c1 verify fail");
  	}
  	return (char*)0;
  }
  if(0 == strcmp(action, "c1DelegateCallDelegateCall")) {
  	c1DelegateCallDelegateCall(args);
  	char* shouldGetStr = TC_StorageGet("c2TestFunc4"); //because context not change
  	TC_Prints(strconcat("TC_StorageGet(\"c2TestFunc4\"): ",shouldGetStr));
  	if(0 == strcmp("this is set in c2TestFunc4", shouldGetStr)){
  	  	TC_Prints("===============================================> test delegateCall + delegatecCall in c1 verify succ");
  	}else{
  	  	TC_Prints("===============================================> test delegateCall + delegatecCall in c1 verify fail");
  	}
  	return (char*)0;
  }
  TC_Prints("no func called in contract1");
  return (char*)0;
}
