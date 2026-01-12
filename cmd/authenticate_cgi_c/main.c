// 引入标准输入输出库，满足CGI程序的输出需求
#include <stdio.h>

int main() {
    printf("Content-Type: text/plain; charset=utf-8\n");
    printf("\n");
    printf("admin\n");
    return 0;
}
