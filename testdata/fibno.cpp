#include "tcapi.h"

extern "C" {

int fibno(int n) {
	if (n == 1 || n == 2)
		return 1;
	else
		return fibno(n-1) + fibno(n-2);
}

char *thunderchain_main(char *action, char *args) {
	if (strcmp(action, "Init") == 0) {
		return "init fibno";
	}

	int n = atoi(args);	
	int sum = 0;
	int i;

	for (i = 0; i < n-2; i++) {
		sum += fibno(n-i);
	}

	return itoa(sum);
}


}
