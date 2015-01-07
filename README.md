### 简介

目前开发工作中，团队通过Gitlab进行版本管理及协作，充分利用Git分支模型，并且所有分支的测试环境共用一台测试服务器。

大家都希望代码push到gitlab后，测试环境的代码能够自动更新，这样能少些重复登录服务器pull代码的操作，也省些脑力负担。

这时Gitlab的webhook功能就派上用场了。webhook是一种HTTP API回调功能，当开发者向Gitlab的代码库push代码或者做其他操作时，会触发这个/些回调，
往回调中发事件数据。

fuse是一个针对webhook的HTTP API实现，可以适配Github、Gitlab、Gogs等平台，特别适用于多开发测试分支的代码库。

### 实现原理

![fuse-arch](https://raw.github.com/youngsterxyf/fuse/master/fuse-arch.png)

### 管理界面

![fuse-console](https://raw.github.com/youngsterxyf/fuse/master/fuse-console.png)

### 部署

- `git clone git@github.com:youngsterxyf/fuse.git`
- `cd fuse`，修改根目录下的`.env`文件，并执行`source .env`
- `go get github.com/go-martini/martini`
- `go get github.com/mattn/go-sqlite3`
- 编译源码：`go install fuse`
- 编辑`conf/app.json`文件，修改配置
- 运行程序：`bin/fuse`
