#include"tcapi.h"

char *thunderchain_main(char *action, char *args) {
  char *ptr1, *ptr2, *ptr3;
  ptr1 = malloc(100);
  strcpy(ptr1, "hello world!");
  ptr2 = calloc(10, sizeof(int));
  ptr3 = realloc(ptr1, 20);
  prints_l(ptr3, 8);
  free(ptr3);
  return ptr3;
}

