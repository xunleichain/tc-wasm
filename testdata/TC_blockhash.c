//#define bool char
#define true 1
#define false 0
typedef char bool;
typedef long long uint64;

void *TC_JsonParse(char *data);
int TC_JsonGetInt(void *root, char *key);
long long TC_JsonGetInt64(void *root, char *key);
char * TC_JsonGetString(void *root, char *key);
char * I64toa(long long amount,int radix);
int strcmp(char *a,char *b);

void TC_prints(const char * cstr);
char *TC_blockhash(uint64 blockNumber);
char *TC_get_coinbase();
uint64 TC_get_gaslimit();
uint64 TC_get_number();
uint64 TC_get_timestamp();
char *TC_get_msg_data();
uint64 TC_get_msg_gas();
char *TC_get_msg_sender();
char *TC_get_msg_sig();
char *TC_get_msg_value();
void TC_assert(bool condition);
void TC_require(bool condition);
uint64 TC_gasleft();
uint64 TC_now();
uint64 TC_get_tx_gasprice();
char *TC_get_tx_origin();
void TC_requireWithMsg(bool condition, char *msg);
void TC_revert();
void TC_revertWithMsg(char *msg);

char *test_TC_blockhash(char *args){
	char *hash;
	void *jsonroot;

	TC_prints("test_TC_blockhash");
	jsonroot = TC_JsonParse(args);
	uint64 blockNumber = TC_JsonGetInt64(jsonroot,"blockid");

	hash = TC_blockhash(blockNumber);
	TC_prints(hash);
	return hash;
}

char *test_TC_get_coinbase(){
	TC_prints("test_TC_get_coinbase");
	char *coinbase;
	coinbase = TC_get_coinbase();
	TC_prints(coinbase);
	return coinbase;
}

uint64 test_TC_get_gaslimit(){
	TC_prints("test_TC_get_gaslimit");
	uint64 gaslimit;
	gaslimit = TC_get_gaslimit();
	TC_prints(I64toa(gaslimit,10));
	return gaslimit;
}
uint64 test_TC_get_number(){
	TC_prints("test_TC_get_number");
	uint64 number;
	number = TC_get_number();
	TC_prints(I64toa(number,10));
	return number;
}
uint64 test_TC_get_timestamp(){
	TC_prints("test_TC_get_timestamp");
	uint64 timestamp;
	timestamp = TC_get_timestamp();
	TC_prints(I64toa(timestamp,10));
	return timestamp;
}
char *test_TC_get_msg_data(){
	TC_prints("test_TC_get_msg_data");
	char *data;
	data = TC_get_msg_data();
	TC_prints(data);
	return data;
}
uint64 test_TC_get_msg_gas(){
	TC_prints("test_TC_get_msg_gas");
	uint64 gas;
	gas = TC_get_msg_gas();
	TC_prints(I64toa(gas,10));
	return gas;
}
char *test_TC_get_msg_sender(){
	TC_prints("test_TC_get_msg_sender");
	char *sender;
	sender = TC_get_msg_sender();
	TC_prints(sender);
	return sender;
}
char *test_TC_get_msg_sig(){
	TC_prints("test_TC_get_msg_sig");
	char *sig;
	sig = TC_get_msg_sig();
	TC_prints(sig);
	return sig;
}
char *test_TC_get_msg_value(){
	TC_prints("test_TC_get_msg_value");
	char *value;
	value = TC_get_msg_value();
	TC_prints(value);
	return value;
}
void test_TC_assert(char *args){
	TC_prints("test_TC_assert");
	void *jsonroot;

	jsonroot = TC_JsonParse(args);
	uint64 where = TC_JsonGetInt64(jsonroot,"where");

	TC_assert(where);
}
void test_TC_require(char *args){
	TC_prints("test_TC_require");
	void *jsonroot;

	jsonroot = TC_JsonParse(args);
	uint64 where = TC_JsonGetInt64(jsonroot,"where");
	TC_require(where);
}
uint64 test_TC_gasleft(){
	TC_prints("test_TC_gasleft");
	uint64 gasleft;
	gasleft = TC_gasleft();
	TC_prints(I64toa(gasleft,10));
	return gasleft;
}
uint64 test_TC_now(){
	TC_prints("test_TC_now");
	uint64 now;
	now = TC_now();
	TC_prints(I64toa(now,10));
	return now;
}
uint64 test_TC_get_tx_gasprice(){
	TC_prints("test_TC_get_tx_gasprice");
	uint64 gasprice;
	gasprice = TC_get_tx_gasprice();
	TC_prints(I64toa(gasprice,10));
	return gasprice;
}
char *test_TC_get_tx_origin(){
	TC_prints("test_TC_get_tx_origin");
	char *origin;
	origin = TC_get_tx_origin();
	TC_prints(origin);
	return origin;
}
void test_TC_requireWithMsg(char *args){
	TC_prints("test_TC_requireWithMsg");
	void *jsonroot;

	jsonroot = TC_JsonParse(args);
	uint64 where = TC_JsonGetInt64(jsonroot,"where");
	char *msg = TC_JsonGetString(jsonroot,"msg");
	TC_requireWithMsg(where,msg);
}
void test_TC_revert(){
	TC_prints("test_TC_revert");
	TC_revert();
}
void test_TC_revertWithMsg(char *args){
	TC_prints("test_TC_revertWithMsg");
	void *jsonroot;

	jsonroot = TC_JsonParse(args);
	char *msg = TC_JsonGetString(jsonroot,"msg");
	TC_revertWithMsg(msg);
}

char *thunderchain_main(char *action, char *args) {
	if (strcmp(action,"TC_blockhash") == 0){
		return test_TC_blockhash(args);
	}else if (strcmp(action,"TC_get_coinbase") == 0){
		test_TC_get_coinbase();
	}else if (strcmp(action,"TC_get_gaslimit") == 0){
		test_TC_get_gaslimit();
	}else if (strcmp(action,"TC_get_number") == 0){
		test_TC_get_number();
	}else if (strcmp(action,"TC_get_timestamp") == 0){
		test_TC_get_timestamp();
	}else if (strcmp(action,"TC_get_msg_data") == 0){
		test_TC_get_msg_data();
	}else if (strcmp(action,"TC_get_msg_gas") == 0){
		test_TC_get_msg_gas();
	}else if (strcmp(action,"TC_get_msg_sender") == 0){
		test_TC_get_msg_sender();
	}else if (strcmp(action,"TC_get_msg_sig") == 0){
		test_TC_get_msg_sig();
	}else if (strcmp(action,"TC_get_msg_value") == 0){
		test_TC_get_msg_value();
	}else if (strcmp(action,"TC_assert") == 0){
		test_TC_assert(args);
	}else if (strcmp(action,"TC_require") == 0){
		test_TC_require(args);
	}else if (strcmp(action,"TC_gasleft") == 0){
		test_TC_gasleft();
	}else if (strcmp(action,"TC_now") == 0){
		test_TC_now();
	}else if (strcmp(action,"TC_get_tx_gasprice") == 0){
		test_TC_get_tx_gasprice();
	}else if (strcmp(action,"TC_get_tx_origin") == 0){
		test_TC_get_tx_origin();
	}else if (strcmp(action,"TC_requireWithMsg") == 0){
		test_TC_requireWithMsg(args);
	}else if (strcmp(action,"TC_revert") == 0){
		test_TC_revert();
	}else if (strcmp(action,"TC_revertWithMsg") == 0){
		test_TC_revertWithMsg(args);
	}else{
		TC_prints("unknown action");
	}
	return (char*)0;
}