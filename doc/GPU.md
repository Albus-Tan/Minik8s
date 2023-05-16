# GPU

## CUDA

> CUDA示例参考 https://docs.hpc.sjtu.edu.cn/job/jobsample2.html#cuda

提交 dgx2 队列作业（使用 GPU）使用 **π 2.0 集群登录节点**：ssh 登录 pilogin.hpc.sjtu.edu.cn (stu1653)

### 使用CUDA编译 .cu 文件

```
$ module load gcc/8.3.0 cuda/10.1.243-gcc-8.3.0
$ nvcc 文件名.cu -o 二进制文件名 -lcublas
```

作业脚本 `名称.slurm` 如下：

### 将作业提交到SLURM上的dgx2分区

```
$ sbatch 名称.slurm
```

# Golang SSH Client

https://github.com/melbahja/goph

```sh
go get github.com/melbahja/goph
```

> 可能会遇到报错：`ssh: handshake failed: knownhosts: key is unknown`
>
> 需要添加 **known_hosts**

# 平台使用

## `squeue` 查看作业信息

| Slurm              | 功能                 |
| ------------------ | -------------------- |
| `squeue -j jobid`  | 查看作业信息         |
| `squeue -l`        | 查看细节信息         |
| `squeue -n HOST`   | 查看特定节点作业信息 |
| `squeue`           | 查看USER_LIST的作业  |
| `squeue --state=R` | 查看特定状态的作业   |
| `squeue --help`    | 查看所有的选项       |

作业状态包括`R`(正在运行)，`PD`(正在排队)，`CG`(即将完成)，`CD`(已完成)。

默认情况下，`squeue`只会展示在排队或在运行的作业。