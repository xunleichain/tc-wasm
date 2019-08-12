char* TC_JsonMashalParams(void* s);
void* malloc(int size);
int strcpy(char* s1, char* s2);
char *thunderchain_main(char *action, char *args) {
	struct Params{
		char ptype[100];
		char pval[100];
	};
	struct Params* pr = malloc(sizeof(struct Params));
	strcpy(pr->ptype,"string");
	strcpy(pr->pval,"json");
	return TC_JsonMashalParams(pr);
}
