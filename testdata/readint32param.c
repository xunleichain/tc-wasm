long long TC_ReadInt32Param(char* args);
char* Itoa(int i);
char *thunderchain_main(char *action, char *args) {
	int i = TC_ReadInt32Param(args);
	return Itoa(i);
}
