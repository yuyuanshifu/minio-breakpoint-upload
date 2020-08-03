# minio-breakpoint-upload
实现minio断点续传功能  
有如下特点：  
1、不暴露minio敏感信息  
2、针对文件每一个分片生成相应的上传地址  
3、文件直接从浏览器上传到minio，不经过后台  
4、部署简单，无须部署额外的类似于sts的服务  

一、效果：  
1、上传页面  
![avatar](https://github.com/yuyuanshifu/minio-breakpoint-upload/blob/master/doc/%E4%B8%8A%E4%BC%A0%E9%A1%B5%E9%9D%A2.png)  
2、前端上传日志  
![avatar](https://github.com/yuyuanshifu/minio-breakpoint-upload/blob/master/doc/%E4%B8%8A%E4%BC%A0%E6%97%A5%E5%BF%97.png)  
3、minio上传日志  
![avatar](https://github.com/yuyuanshifu/minio-breakpoint-upload/blob/master/doc/minio%E4%B8%8A%E4%BC%A0%E6%97%A5%E5%BF%97.png)  

二、详细方案
![avatar](https://github.com/yuyuanshifu/minio-breakpoint-upload/blob/master/doc/minio.png)

流程可参考：https://www.cnblogs.com/xiahj/p/vue-simple-uploader.html

https://github.com/minio/minio-go/issues/1324  
minio本身并没有提供断点续传的接口，但其实minio的PutObject上传接口内部是实现了分片上传的，仔细阅读此接口的源码，  

不同之处在于：  
1、根据文件分片生成上传地址  
参考：https://github.com/singularityhub/sregistry/pull/298
