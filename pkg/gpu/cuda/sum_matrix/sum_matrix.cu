#include <cuda_runtime.h>
#include <stdio.h>

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

void initialData(float* ip,int size)
{
  for(int i=0;i<size;i++)
  {
    ip[i]=(float)(rand()&0xffff)/1000.0f;
  }
}

void printDataFront(float* ip,int size)
{
  for(int i=0;i<size;i++)
  {
    printf("%f ", ip[i]);
  }
  printf("......\n");
  printf("-------------------------\n");
}

void initDevice(int devNum)
{
  int dev = devNum;
  cudaDeviceProp deviceProp;
  CHECK_ERROR(cudaGetDeviceProperties(&deviceProp,dev));
  printf("Using device %d: %s\n",dev,deviceProp.name);
  CHECK_ERROR(cudaSetDevice(dev));

}

// 核函数，每一个线程计算矩阵中的一个元素
__global__ void sumMatrix(float * MatA,float * MatB,float * MatC,int nx,int ny)
{
    int ix=threadIdx.x+blockDim.x*blockIdx.x;   // col
    int iy=threadIdx.y+blockDim.y*blockIdx.y;   // row
    int idx=ix+iy*nx;
    if (ix<nx && iy<ny)
    {
        MatC[idx] = MatA[idx]+MatB[idx];
    }
}

//主函数
int main(int argc,char** argv)
{
    //设备初始化
    printf("strating...\n");
    initDevice(0);

    //输入二维矩阵，4096*4096，单精度浮点型。
    int nx = 1<<12;
    int ny = 1<<12;
    int nBytes = nx*ny*sizeof(float);
    int numToPrint = 16;

    //Malloc，开辟主机内存
    float* A_host = (float*)malloc(nBytes);
    float* B_host = (float*)malloc(nBytes);
    float* C_from_gpu = (float*)malloc(nBytes);

    //初始化矩阵
    initialData(A_host, nx*ny);
    printf("A matrix data: \n");
    printDataFront(A_host, numToPrint);

    initialData(B_host, nx*ny);
    printf("B matrix data: \n");
    printDataFront(B_host, numToPrint);

    //cudaMalloc，开辟设备内存
    float* A_dev = NULL;
    float* B_dev = NULL;
    float* C_dev = NULL;
    CHECK_ERROR(cudaMalloc((void**)&A_dev, nBytes));
    CHECK_ERROR(cudaMalloc((void**)&B_dev, nBytes));
    CHECK_ERROR(cudaMalloc((void**)&C_dev, nBytes));

    //输入数据从主机内存拷贝到设备内存
    CHECK_ERROR(cudaMemcpy(A_dev, A_host, nBytes, cudaMemcpyHostToDevice));
    CHECK_ERROR(cudaMemcpy(B_dev, B_host, nBytes, cudaMemcpyHostToDevice));

    //二维线程块，32×32
    dim3 block(32, 32);
    //二维线程网格，128×128
    dim3 grid((nx-1)/block.x+1, (ny-1)/block.y+1);

    //将核函数放在线程网格中执行
    sumMatrix<<<grid,block>>>(A_dev, B_dev, C_dev, nx, ny);
    CHECK_ERROR(cudaDeviceSynchronize());

    //拷贝回结果数据
    CHECK_ERROR(cudaMemcpy(C_from_gpu, C_dev, nBytes, cudaMemcpyDeviceToHost));

    //输出数据
    printf("result data: \n");
    printDataFront(C_from_gpu, numToPrint);

    cudaFree(A_dev);
    cudaFree(B_dev);
    cudaFree(C_dev);
    free(A_host);
    free(B_host);
    free(C_from_gpu);
    cudaDeviceReset();
    return 0;
}