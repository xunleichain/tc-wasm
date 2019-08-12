#include"tcapi.h"

char *thunderchain_main(char *action, char *args) {

  if(0 == strcmp("Init", action)){
	  return "Init ok";
  }

  void* jsonObject = TC_JsonParse(args);
  char* contract1 = TC_JsonGetString(jsonObject, "contract1");
  char* contract2 = TC_JsonGetString(jsonObject, "contract2");
  TC_Prints(contract1);
  TC_Prints(contract2);

  if(contract1 == NULL || contract2 == NULL){
	  return "params err";
  }

  char* shouldGetStr;
  char* shouldNotGetStr;

  TC_Prints(strconcat("contract address: ", TC_GetSelfAddress()));
  TC_Prints(strconcat("caller address: ", TC_GetMsgSender()));
  TC_Prints(strconcat("msg data: ", TC_GetMsgData()));
  TC_Prints(strconcat("msg sign: ", TC_GetMsgSign()));
  TC_Prints(strconcat("msg value: ", TC_GetMsgValue()));
  TC_Prints(strconcat("msg token value: ", TC_GetMsgTokenValue()));

  //call
  TC_Prints("=======begin test Call in c0==========");
  TC_CallContract(contract1, "testCall", "{\"key_testCall\":\"value_testCall\"}");
  shouldNotGetStr = TC_StorageGet("testCall");
  TC_Prints(strconcat("TC_StorageGet(\"testCall\"): ",shouldNotGetStr));
  if(0 != strcmp("this is set in c1 testCall", shouldNotGetStr)){
	  TC_Prints("===============================================> test Call in c0 verify succ");
  }else{
	  TC_Prints("===============================================> test Call in c0 verify fail");
  }
  TC_Prints("=======end test Call in c0==========");

  //delegatecall
  TC_Prints("=======begin test delegateCall in c0==========");
  TC_DelegateCallContract(contract1, "testDelegateCall", "{\"key_testDelegateCall\":\"value_testDelegateCall\"}");
  shouldGetStr = TC_StorageGet("testDelegateCall");
  TC_Prints(strconcat("TC_StorageGet(\"testDelegateCall\"): ",shouldGetStr));
  if(0 == strcmp("this is set in c1 testDelegateCall", shouldGetStr)){
  	  TC_Prints("===============================================> test delegateCall in c0 verify succ");
  }else{
  	  TC_Prints("===============================================> test delegateCall in c0 verify fail");
  }
  TC_Prints("=======end test delegateCall in c0==========");

  //call + call
  TC_Prints("=======begin Call + Call in c0==========");
  TC_CallContract(contract1, "c1CallCall", contract2);
  shouldNotGetStr = TC_StorageGet("c2TestFunc1");
  TC_Prints(strconcat("TC_StorageGet(\"c2TestFunc1\"): ",shouldNotGetStr));
  if(0 != strcmp("this is set in c2TestFunc1", shouldNotGetStr)){
      TC_Prints("===============================================> test Call + Call in c0 verify succ");
  }else{
      TC_Prints("===============================================> test Call + Call in c0 verify fail");
  }
  TC_Prints("=======end Call + Call in c0==========");

  //call + delegatecall
  TC_Prints("=======begin Call + delegateCall in c0==========");
  TC_CallContract(contract1, "c1CallDelegateCall", contract2);
  shouldNotGetStr = TC_StorageGet("c2TestFunc2");
  TC_Prints(strconcat("TC_StorageGet(\"c2TestFunc2\"): ",shouldNotGetStr));
  if(0 != strcmp("this is set in c2TestFunc2", shouldNotGetStr)){
    TC_Prints("===============================================> test Call + delegateCall in c0 verify succ");
  }else{
    TC_Prints("===============================================> test Call + delegateCall in c0 verify fail");
  }
  TC_Prints("=======end Call + delegateCall in c0==========");

  //delegatecall + call
  TC_Prints("=======begin delegateCall + Call in c0==========");
  TC_DelegateCallContract(contract1, "c1DelegateCallCall", contract2);
  shouldNotGetStr = TC_StorageGet("c2TestFunc3");
  TC_Prints(strconcat("TC_StorageGet(\"c2TestFunc3\"): ",shouldNotGetStr));
  if(0 != strcmp("this is set in c2TestFunc3", shouldNotGetStr)){
    TC_Prints("===============================================> test delegateCall + Call in c0 verify succ");
  }else{
    TC_Prints("===============================================> test delegateCall + Call in c0 verify fail");
  }
  TC_Prints("=======end delegateCall + Call in c0==========");

  //delegatecall + delegatecall
  TC_Prints("=======begin delegateCall + delegateCall in c0==========");
  TC_DelegateCallContract(contract1, "c1DelegateCallDelegateCall", contract2);
  shouldGetStr = TC_StorageGet("c2TestFunc4");
  TC_Prints(strconcat("TC_StorageGet(\"c2TestFunc4\"): ",shouldGetStr));
  if(0 == strcmp("this is set in c2TestFunc4", shouldGetStr)){
     TC_Prints("===============================================> test delegateCall + delegateCall in c0 verify succ");
  }else{
     TC_Prints("===============================================> test delegateCall + delegateCall in c0 verify fail");
  }
  TC_Prints("=======end delegateCall + delegateCall in c0==========");

  return (char*)0;
}
