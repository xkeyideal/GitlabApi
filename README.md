# GitlabApi

Just for Gitlab&Git API
此处并没有实现全套的Gitlab API，仅仅根据项目需要实现部分API，Gitlab API文档：http://docs.gitlab.com/ce/api/

所有代码均测试通过，并且已经在自己的项目中使用，代码里并没有给出测试代码。

## 第三方依赖库

HTTP通信采用了Beego的httplib库，并且测试使用了goreq库，该库支持链式设置
httplib: github.com/astaxie/beego/httplib
goreq: github.com/smallnest/goreq

复杂的Json解析采用了go-simplejson库
go-simplejson： github.com/bitly/go-simplejson

## HTTP库 Delete 存在的问题
httplib和goreq库均存在DELETE操作的时候，无法提交POST参数，导致删除操作失败，goreq库有个很好的功能SetCurlCommand实现打印curl命令
具体代码的修改：
httplib:
	if (b.req.Method == "POST" || b.req.Method == "DELETE" || b.req.Method == "PUT" || b.req.Method == "PATCH") && b.req.Body == nil 

goreq: 
	case POST, PUT, PATCH, DELETE:

## Gitlab API

项目的创建是基于Group的，具体涉及参数config.NamespaceId，对于具体使用场景请注意，应该需要修改相应的代码

## Git API

该API是用于对本地的.git项目进行操作，主要实现方式是基于git命令，并没有使用bash脚本，实现的操作有：
1. clone
2. pull
3. push
4. file head commitid
5. branch head commitid
6. file branch commit lists


最后，此代码仅供参考。

