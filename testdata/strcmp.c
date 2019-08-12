int memset(void *dst, char c, int length);
int strcmp(char *a, char *b);
void* malloc(int size);

char *thunderchain_main(char *action, char *args) {
	char* p_str1 = (char*)malloc(sizeof(char)*100);
	memset(p_str1, 'h', 5);
	char* p_str2 = (char*)malloc(sizeof(char)*100);
	memset(p_str2, 'i', 5);
	int ret = strcmp(p_str1, p_str2);
	return ret;
}
