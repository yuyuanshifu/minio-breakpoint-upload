# minio-breakpoint-upload
## 一、基本介绍  
完整实现minio分片上传、断点续传、秒传等功能。  

此方案有如下特点：  
1、不暴露minio敏感信息  
2、针对文件每一个分片生成相应的上传地址  
3、文件直接从浏览器上传到minio，不经过后台  
4、部署简单，无须部署额外的类似于sts的服务  

## 二、效果演示  
1、上传页面  
![avatar](doc/%E4%B8%8A%E4%BC%A0%E9%A1%B5%E9%9D%A2.png)  
2、前端上传日志  
![avatar](doc/%E4%B8%8A%E4%BC%A0%E6%97%A5%E5%BF%97.png)  
3、minio上传日志  
![avatar](doc/minio%E4%B8%8A%E4%BC%A0%E6%97%A5%E5%BF%97.png)  

## 三、使用说明  
### web端  
```bash
cd web_src/minio/build
npm run build
```

### server端
```bash
go build main.go
```

## 四、详细方案  
minio官方并没有提供断点续传的方案，但  
（1）minio的PutObject上传接口内部是实现了分片上传的，我们可以通过此接口封装出分片上传地址生成接口  
（2）ListIncompleteUploads接口内部可以查询到已经上传成功的分片信息，包括分片的序号以及对应的etag，我们可以通过此接口封装出查询上传成功的分片信息接口  

具体流程如下：  
![avatar](doc/%E6%96%B0%E6%96%B9%E6%A1%88%EF%BC%882020.09.09%EF%BC%89.png)


## 四、更新日志  
|  日期   | 日志  |
|  :---:  | --- |
|2020/08/03| 分片上传 断点续传 秒传 |
|2020/09/09| 不再在mysql中记录分片上传结果以及etag |
|2020/12/09| 解决大文件上传过程中浏览器内存溢出问题 |

## 五、博客地址  
> https://blog.csdn.net/lmlm21/article/details/107768581  

## 六、联系方式  
vx：lm3775859
