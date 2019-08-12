
int strcpy(char *s1, char *s2);
void *malloc(int size);

struct Params {
  char address[64];
  char domain[64];
};

char *thunderchain_main(char *action, char *args) {
  struct Params *p = (struct Params *) malloc(sizeof(struct Params)); 
  strcpy(p->address, "localhost");
  strcpy(p->domain, "test.onething.com");
  p->domain[0] = 'F';
  return p->domain;
}