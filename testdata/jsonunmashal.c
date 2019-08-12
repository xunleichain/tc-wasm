char* TC_JsonUnmashalInput(void* addr, int size, char* args);
void* malloc(int size);
char *thunderchain_main(char *action, char *args) {
	struct Args{
		int a;
		int b;
		char* str;
	};
	struct Args* input = (struct Args*)malloc(sizeof(struct Args));
	TC_JsonUnmashalInput(input, sizeof(struct Args), args);
	return input->str;
}
