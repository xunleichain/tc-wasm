long long Atoi64(char * s);
char *thunderchain_main(char *action, char *args) {
	int s = 0;
	long long i = Atoi64("10000000000");
	while (i>0){
		i=i/10;
		s++;
	}
	return s;
}
