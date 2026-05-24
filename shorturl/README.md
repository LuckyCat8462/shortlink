一、基础框架搭建

（一）建库建表

1、新建发号器表

CREATE TABLE `sequence` (
`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
`stub` varchar(1) NOT NULL,
`timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
PRIMARY KEY (`id`),
UNIQUE KEY `idx_uniq_stub` (`stub`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8 COMMENT = '序号表';

2、新建长链接短链接映射表

CREATE TABLE `short_url_map` (
`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
`create_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
`create_by` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建者',
`is_del` tinyint UNSIGNED NOT NULL DEFAULT '0' COMMENT '是否删除：0正常1删除',

`lurl` varchar(2048) DEFAULT NULL COMMENT '长链接',
`md5` char(32) DEFAULT NULL COMMENT '长链接MD5',
`surl` varchar(11) DEFAULT NULL COMMENT '短链接',
PRIMARY KEY (`id`),
INDEX(`is_del`),
UNIQUE(`md5`),
UNIQUE(`surl`)
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COMMENT = '长短链映射表';

（二）搭建go-zero框架
1、编写api文件，使用goctl命令生成代码
1.1.shorturl.api文件
service shorturl-api {
//	发长链，映射短链接
//post /convert,静态路由，不携带动态参数，适用于资源明确、不需要通过URL路径区分
@handler ShorturlHandler
post /convert(ConvertRequest) returns (ConvertResponse)

//	通过短链查长链
//	post /:shortUrl，动态路由，带路径参数，冒号后面的shortUrl是一个占位符，可以匹配实际的url
@handler ShowUrlHandler
post /:shortUrl(ShowUrlRequest) returns (ShowUrlResponse)
}

1.2.使用goctl命令生成代码

bash

goctl api go -api shorturl.api -dir .



1.3.根据数据表，生成model层

bash

goctl model mysql datasource -url="yourcount:your password@tcp(127.0.0.1:port)/shortlink" -table="short_url_map" -dir="./model"

goctl model mysql datasource -url="yourcount:your password@tcp(127.0.0.1:port)/shortlink" -table="sequence" -dir="./model"



1.4.下载项目依赖

go mod tidy



1.5.运行项目

go run shorturl.go

查看到 Starting server at 0.0.0.0:8888...成功



1.6.修改.yaml与config配置文件

注意：关注对齐、空格等内容



二、参数校验

2.1.第三方库validator

[validator package - github.com/go-playground/validator/v10 - Go Packages](https://pkg.go.dev/github.com/go-playground/validator/v10)

下载：

```bash
go get github.com/go-playground/validator/v10
```

导入

```bash
import "github.com/go-playground/validator/v10"
```



在.api中，为结构体添加validate:"required" tag,并添加校验规则

## 查看短链接

### 缓存方法

1、使用自己的缓存，将短链接映射到长链接，能够节省缓存的数据量surl->lurl
2、使用redis缓存，将短链接映射到长链接，surl->数据行,开发量小，实现简单

2.1、添加redis配置
-缓存文件
-配置config结构体
2.2.删除旧的model代码，生成新的model代码




