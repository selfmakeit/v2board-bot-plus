# V2board机器人加强版


## 说明

区分用户和管理员角色，不同角色不同操作菜单，管理员可查看统计信息，可通过bot回复用户以及回复工单

# Demo

## 管理员端

![1](https://raw.githubusercontent.com/selfmakeit/resource/main/admin1.png)

## 用户端

![2](https://raw.githubusercontent.com/selfmakeit/resource/main/user1.png)

![3](https://raw.githubusercontent.com/selfmakeit/resource/main/user2.png)

# 配置

配置文件Demo：

```yaml
debug: false
isAutoDeleteMsg: true #是否自动删除bot信息
appName: XX加速器
websiteUrl: https://board.jike212.xyz/
tgGroupLink: https://t.me/GanYaBotGroup
inviteUrl: https://j12vpn2.net/?code= #邀请链接把邀请码去掉
mysql:
  host: localhost
  port: 3306
  database: v2b
  username: spike
  passwd: 12345678
redis:
  host: localhost:6379 #redis地址
  db: 0
  password: '' #redis有密码的话需要填
  cacheTime: 24 #缓存来自crisp的消息的时间
telegram:
  key: 121232132:AAHXQ4324324233423412E0fxU6o123 #telegram bot的api key
  mode: group #群模式还是私聊模式
  admins:
    - 6223132130 #管理员id
```
