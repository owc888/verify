# 参数验证

## godoc -- 项目文档
1.安装godoc

    go get -v  golang.org/x/tools/cmd/godoc
2.上面操作会生成二进制文件（godoc.exe），找出来，执行godoc

    $GOPATH\bin\godoc.exe -http=:6060  //windows
    $GOPATH\bin\godoc -http=:6060  //linux
3.后可在浏览器localhost:6060查看文档

4.规则参考文档

https://www.jianshu.com/p/b91c4400d4b2
https://www.kancloud.cn/cattong/go_command_tutorial/261351

## swag -- API文档
1.安装swag

    go get -u github.com/swaggo/swag/cmd/swag
2.上面操作会生成二进制文件（swag.exe），找出来，执行swag

    $GOPATH\bin\swag.exe init --md ./docs  //windows
    $GOPATH\bin\swag init --md ./docs  //linux
3.后生成.\docs目录，且目录下会生成几个文档

4.运行项目，浏览器输入：`/docs/api/index.html`即可

5.规则参考文档：https://github.com/swaggo/swag/blob/master/README_zh-CN.md

6.每个接口的注释参照下面的实例来写

    // @Tags         标签（类似分组）
    // @Summary      概要（类似标题）
    // @Description  详细描述
    // @Accept       application/json
    // @Produce      application/json
    // @Param        raw  body      request.Data     true  "参数描述"
    // @Success      200  {object}  response.Response  "参数描述"
    // @Header       200  {string}  x-Token          "参数描述"
    // @Router       path [httpMethod]
6.1.同时在入参和出参声明结构体的成员上，要加上注释（后面加//），以便api文档查看

6.1.有些结构体后面会接上`//@name xxx`的注释，这是用于查看swag文档时对结构体的重命名，注意避免重复名称导致的冲突


