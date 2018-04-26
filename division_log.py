#! /usr/bin/env python3.6

import os
import datetime

logPath = "/home/wwwlogs/"
logPrefix = "www.markbest.site"
nginxPid = "/usr/local/nginx/logs/nginx.pid"

# nginx日志分割
last5Min = (datetime.datetime.now() - datetime.timedelta(minutes=5)).strftime("%Y-%m-%d-%H-%M")
os.chdir(logPath)
os.system("mv " + logPrefix + ".log " + logPrefix + "-" + last5Min + ".log")
os.system("kill -USR1 `cat " + nginxPid + "`")

# 只保存2个小时内日志分割的文件
logList = os.listdir(logPath)
expiredLogPrefix = (datetime.datetime.now() - datetime.timedelta(hours=2)).strftime("%Y-%m-%d-%H")
for log in logList:
    if log.startswith(logPrefix + "-" + expiredLogPrefix):
        os.remove(logPath + log)

