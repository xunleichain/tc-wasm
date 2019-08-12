
int arrayLen(void *a);
void *malloc(int size);

char *thunderchain_main(char *action, char *args) {
	struct Array{
		char str[100];
	};
	struct Array * arr = (struct Array*)malloc(sizeof(struct Array));
	int len = arrayLen(arr);
	return len;
}
