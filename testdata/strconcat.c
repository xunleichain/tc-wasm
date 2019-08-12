int memset(void * dest,char c,int length);
char* strconcat(char *a, char *b);
void* malloc(int size);

char *thunderchain_main(char *action, char *args) {
	char* p_str1 = (char*)malloc(sizeof(char)*100);
	memset(p_str1, 'h', 5);
	char* p_str2 = (char*)malloc(sizeof(char)*100);
	memset(p_str2, 's', 5);
	char* ret = strconcat(p_str1, p_str2);
	return ret;
}
