int memcpy(void* dst, void* src, int length);
void* malloc(int size);

char *thunderchain_main(char *action, char *args) {
	char* p_str1 = (char*)malloc(sizeof(char)*100);
	memcpy(p_str1, "hellowrold", 5);
	return p_str1;
}
