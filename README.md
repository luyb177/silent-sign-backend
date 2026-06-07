# silent-sign-backend
无声之韵智能手语双向翻译系统后端

# 命令行

生成文件
```bash
goctl api go -api silent_sign.api -dir . --style go_zero
```
生成swag
```bash
goctl api swagger -api silent_sign.api -dir ./docs -filename swagger
```

# Docker 构建 & 部署

## 构建镜像（本地 → 阿里云 ACR）

```bash
# 构建 linux/amd64 镜像（Mac M 系列必须指定平台）
docker buildx build --platform linux/amd64 \
  -t crpi-u5azhs6neq326bz0.cn-hangzhou.personal.cr.aliyuncs.com/yub_lu/silent_sign:0.0.1 \
  --push .
```

> `--push` 会构建完成后直接推到 ACR，省去 docker push 步骤。
> 首次使用需要 `docker login` 登录 ACR。

## 远程服务器部署

```bash
# 1. 创建配置目录
mkdir -p ~/silent-sign/service/silent-sign/{etc,data}

# 2. 放入配置文件
#   - etc/silent_sign.yaml  （服务配置）
#   - data/ip2region_v4.xdb  （IP 定位库）
#   - data/ip2region_v6.xdb

# 3. 登录 ACR（首次）
docker login --username=<账号> registry.cn-hangzhou.aliyuncs.com

# 4. 启动所有服务
docker compose up -d

# 5. 更新服务（本地重新构建推送后）
docker compose pull
docker compose up -d
```