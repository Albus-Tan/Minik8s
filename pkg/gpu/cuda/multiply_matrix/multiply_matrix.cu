#include <stdio.h>
#include <cuda_runtime.h>
// Matrix size: 50 * 25 & 25 * 50
const int M = 15;
const int N = 10;

#define CHECK_ERROR(call)\
{\
  const cudaError_t error=call;\
  if(error!=cudaSuccess)\
  {\
      printf("ERROR: %s:%d,",__FILE__,__LINE__);\
      printf("code:%d,reason:%s\n",error,cudaGetErrorString(error));\
      exit(1);\
  }\
}

// Matrix multiply: C = A * B
__global__ void matrix_multiply(int **A, int **B, int **C) {
    int i = blockIdx.x * blockDim.x + threadIdx.x;
    int j = blockIdx.y * blockDim.y + threadIdx.y;
    int value = 0;
    for (int k = 0; k < N; k++) {
        value += A[i][k] * B[k][j];
    }
    C[i][j] = value;
}

void initDevice(int devNum)
{
  int dev = devNum;
  cudaDeviceProp deviceProp;
  CHECK_ERROR(cudaGetDeviceProperties(&deviceProp,dev));
  printf("Using device %d: %s\n",dev,deviceProp.name);
  CHECK_ERROR(cudaSetDevice(dev));

}

int main() {
    printf("strating...\n");
    initDevice(0);

    int **A = (int **) malloc(sizeof(int *) * M);
    int **B = (int **) malloc(sizeof(int *) * N);
    int **C = (int **) malloc(sizeof(int *) * M);

    int *data_A = (int *) malloc(sizeof(int) * M * N);
    int *data_B = (int *) malloc(sizeof(int) * M * N);
    int *data_C = (int *) malloc(sizeof(int) * M * M);
    for (int i = 0; i < M * N; i++) {
        data_A[i] = i;
        data_B[i] = i;
    }

    printf("Matrix A is:\n");
    for (int i = 0; i < M; i++) {
        for (int j = 0; j < N ; j++) {
            printf("%d ", data_A[i * N + j]);
        }
        printf("\n");
    }

    printf("Matrix B is:\n");
    for (int i = 0; i < N; i++) {
        for (int j = 0; j < M ; j++) {
            printf("%d ", data_B[i * M + j]);
        }
        printf("\n");
    }

    int *dev_data_A;
    int *dev_data_B;
    int *dev_data_C;

    // malloc matrix (size = M*N) in GPU device
    CHECK_ERROR(cudaMalloc((void **) &dev_data_A, sizeof(int) * M * N));
    CHECK_ERROR(cudaMalloc((void **) &dev_data_B, sizeof(int) * M * N));
    CHECK_ERROR(cudaMalloc((void **) &dev_data_C, sizeof(int) * M * M));

    // copy data from host to GPU device
    CHECK_ERROR(cudaMemcpy((void *) dev_data_A, (void *) data_A, sizeof(int) * M * N, cudaMemcpyHostToDevice));
    CHECK_ERROR(cudaMemcpy((void *) dev_data_B, (void *) data_B, sizeof(int) * M * N, cudaMemcpyHostToDevice));
    // init C
    CHECK_ERROR(cudaMemset((void *) dev_data_C, 0, sizeof(int) * M * M));

    for (int i = 0; i < M; i++) {
        A[i] = dev_data_A + i * N;
        C[i] = dev_data_C + i * M;
    }

    for (int i = 0; i < N; i++) {
        B[i] = dev_data_B + i * M;
    }

    int **dev_A;
    int **dev_B;
    int **dev_C;

    CHECK_ERROR(cudaMalloc((void **) &dev_A, sizeof(int *) * M));
    CHECK_ERROR(cudaMalloc((void **) &dev_B, sizeof(int *) * N));
    CHECK_ERROR(cudaMalloc((void **) &dev_C, sizeof(int *) * M));

    CHECK_ERROR(cudaMemcpy((void *) dev_A, (void *) A, sizeof(int *) * M, cudaMemcpyHostToDevice));
    CHECK_ERROR(cudaMemcpy((void *) dev_B, (void *) B, sizeof(int *) * N, cudaMemcpyHostToDevice));
    CHECK_ERROR(cudaMemcpy((void *) dev_C, (void *) C, sizeof(int *) * M, cudaMemcpyHostToDevice));

    dim3 threadPerBlock(5, 5);
    dim3 numBlocks(M / threadPerBlock.x, M / threadPerBlock.y);

    matrix_multiply <<<numBlocks, threadPerBlock>>> (dev_A, dev_B, dev_C);

    // copy result to host
    CHECK_ERROR(cudaMemcpy((void *) data_C, (void *) dev_data_C, sizeof(int) * M * M, cudaMemcpyDeviceToHost));

    // print result:
    printf("The matrix multiply result is:\n");
    for (int i = 0; i < M; i++) {
        for (int j = 0; j < M ; j++) {
            printf("%d ", data_C[i * M + j]);
        }
        printf("\n");
    }
}