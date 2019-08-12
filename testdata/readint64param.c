long long TC_ReadInt64Param(char* args);
char* I64toa(long long i);
char *thunderchain_main(char *action, char *args) {
	long long i = TC_ReadInt64Param(args);
	return I64toa(i);
}
