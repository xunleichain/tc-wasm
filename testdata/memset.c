int memset(void* dst, char c, int length);
void* malloc(int size);

char *thunderchain_main(char *action, char *args) {
	char* p_str1 = (char*)malloc(sizeof(char)*100);
	memset(p_str1, 'h', 5);
	return p_str1;
}
