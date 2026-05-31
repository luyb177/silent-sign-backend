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