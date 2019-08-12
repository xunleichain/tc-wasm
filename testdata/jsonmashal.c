char* TC_JsonMashalResult(void* val, char* type, int succeed);
char *thunderchain_main(char *action, char *args) {
	return TC_JsonMashalResult("123", "string", 1);
}
