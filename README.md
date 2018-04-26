## nginxlog
Golang解析nginx日志，存入elasticsearch

## 使用方法
- 进入conf目录下将conf.toml.example重命名为conf.toml并完成配置信息

```
[app]
port = ":8090"

[elastic]
elastic_url = "elasticsearch地址"
elastic_index = "elasticsearch index"
elastic_type = "elasticsearch type"
elastic_log_path = "./logs"
elastic_log_max_files = 5

[log]
target_path = "日志目录"
tartet_file_prefix = "默认日志文件前缀名称"
```
- 编译为可执行文件: go build -o bin/nginxlog main.go，然后加入path目录
- 执行命令分析日志入库

```
bin/nginxlog
```

## 接口路由
```
//状态码查询
api/analysis/status?status=[status]&per_page=[per_page]&page=[page]
//请求方法查询
api/analysis/method?method=[method]&per_page=[per_page]&page=[page]
//top ip访问查询
api/analysis/topIp
```

## 日志分割脚本
- division_log.py
日志文件5分钟分割一次