<template>
  <uploader 
          ref="uploader"
          :options="options" 
          :autoStart="false" 
          @file-added="onFileAdded"
          fileStatusText="fileStatusText"
          class="uploader-app">
    <uploader-unsupport></uploader-unsupport>
    <uploader-drop>
      <p>拖动文件</p>
      <uploader-btn>选择文件</uploader-btn>
    </uploader-drop>
    <uploader-list></uploader-list>
    <p>文件处理状态：{{status}}</p>
    <p>文件上传进度：{{progress}}%</p>
  </uploader>
</template>

<script>

  import SparkMD5 from 'spark-md5';
  import axios from 'axios'
  import qs from 'qs'

  export default {
    data () {
      return {
        progress: 0,
        status: '初始状态',
        urlPrex: 'http://192.168.207.34:39988/minio',
      }
    },
    created() {
        //const uploaderInstance = this.$refs.uploader;
    },
    mounted () {
      this.$nextTick(() => {
        window.uploader = this.$refs.uploader.uploader
      })
    },
    methods: {
        onFileAdded(file) {
          this.progress=0;
          this.status='初始状态';
          file.urlPrex = this.urlPrex;
          // 计算MD5
          this.computeMD5(file);
        },
        getSuccessChunks(file) {
          return new Promise((resolve, reject) => {
            axios.get(file.urlPrex + '/get_chunks', {params :{
              md5: file.uniqueIdentifier,
            }}).then(function (response) {
              file.uploadID = response.data.uploadID;
              file.uuid = response.data.uuid;
              file.uploaded = response.data.uploaded;
              file.chunks = response.data.chunks;
              resolve(response);
            }).catch(function (error) {
              console.log(error);
              reject(error);
            });
          })

        },
        newMultiUpload(file) {
          return new Promise((resolve, reject) => {
            axios.get(file.urlPrex + '/new_multipart', {params :{
              totalChunkCounts: file.totalChunkCounts,
              md5: file.uniqueIdentifier,
              size: file.size,
              fileName: file.name
            }}).then(function (response) {
              file.uploadID = response.data.uploadID;
              file.uuid = response.data.uuid;
              resolve(response);
            }).catch(function (error) {
              console.log(error);
              reject(error);
            });
          })
        },
        multipartUpload(file) {
          let blobSlice = File.prototype.slice || File.prototype.mozSlice || File.prototype.webkitSlice,
            chunkSize = 1024*1024*64,
            chunks = Math.ceil(file.size / chunkSize),
            currentChunk = 0,
            fileReader = new FileReader(),
            time = new Date().getTime();

          function loadNext() {
            let start = currentChunk * chunkSize;
            let end = ((start + chunkSize) >= file.size) ? file.size : start + chunkSize;

            fileReader.readAsArrayBuffer(blobSlice.call(file.file, start, end));
          }

          function checkSuccessChunks() {
            var index = successChunks.indexOf((currentChunk+1).toString())
            if (index == -1) {
              return false;
            }

            return true;
          }

          function getUploadChunkUrl(currentChunk, partSize) {
            return new Promise((resolve, reject) => {
                axios.get(file.urlPrex + '/get_multipart_url', {params :{
                  uuid: file.uuid,
                  uploadID: file.uploadID,
                  size: partSize,
                  chunkNumber: currentChunk+1
                }}).then(function (response) {
                  urls[currentChunk] = response.data.url
                  resolve(response);
                }).catch(function (error) {
                  console.log(error);
                  reject(error);
                });
              })
          }

          function uploadMinio(url, e) {
            return new Promise((resolve, reject) => {
              
              axios.put(url, e.target.result
                ).then(function (res) {
                  etags[currentChunk] = res.headers.etag;
                  resolve(res);
                }).catch(function (err) {
                  console.log(err);
                  reject(err);
                });
            });
          }

          async function uploadMinioNew(url,e){
            var xhr = new XMLHttpRequest();
            xhr.open('PUT', url, false);
            xhr.setRequestHeader('Content-Type', 'text/plain')
            xhr.send(e.target.result);
            var etagValue = xhr.getResponseHeader('etag');
            etags[currentChunk] = etagValue;
          }

          function updateChunk(currentChunk) {
            return new Promise((resolve, reject) => {
                axios.post(file.urlPrex + '/update_chunk', qs.stringify({
                  uuid: file.uuid,
                  chunkNumber: currentChunk+1,
                  etag: etags[currentChunk]
                })).then(function (response) {
                  resolve(response);
                }).catch(function (error) {
                  console.log(error);
                  reject(error);
                });
              })
          }

          async function uploadChunk(e) {
            if (!checkSuccessChunks()) {
              let start = currentChunk * chunkSize;
              let partSize = ((start + chunkSize) >= file.size) ? file.size -start : chunkSize;

              //获取分片上传url
              await getUploadChunkUrl(currentChunk, partSize);
              if (urls[currentChunk] != "") {
                //上传到minio
                await uploadMinioNew(urls[currentChunk], e);
                if (etags[currentChunk] != "") {
                  //更新数据库：分片上传结果
                  //await updateChunk(currentChunk);
                } else {
                  return;
                }
              } else {
                return;
              }
              
            }
            
          };

          function completeUpload(){
            return new Promise((resolve, reject) => {
                axios.post(file.urlPrex + '/complete_multipart', qs.stringify({
                  uuid: file.uuid,
                  uploadID: file.uploadID,
                  file_name: file.name,
                  size: file.size,
                })).then(function (response) {
                  resolve(response);
                }).catch(function (error) {
                  console.log(error);
                  reject(error);
                });
              })
          }

          var successChunks = new Array();
          var successParts = new Array();
          successParts = file.chunks.split(",");
          for (let i = 0; i < successParts.length; i++) {
            successChunks[i] = successParts[i].split("-")[0];
          }
          
          var urls = new Array();
          var etags = new Array();

          console.log('上传分片...');
          this.status='上传中';
          
          {
            loadNext();
            fileReader.onload = async (e) => {
              await uploadChunk(e);
              fileReader.abort();
              currentChunk++;
        
              if (currentChunk < chunks) {
                  console.log(`第${currentChunk}个分片上传完成, 开始第${currentChunk +1}/${chunks}个分片上传`);
                  this.progress = Math.ceil((currentChunk / chunks)*100);
                  await loadNext();
              } else {
                  await completeUpload();
                  console.log(`文件上传完成：${file.name} \n分片：${chunks} 大小:${file.size} 用时：${(new Date().getTime() - time)/1000} s`);
                  this.progress = 100;
                  this.status='上传完成';
                  //window.location.reload();
              }
            };
          }

        },
        //计算MD5
        computeMD5(file) {
            let blobSlice = File.prototype.slice || File.prototype.mozSlice || File.prototype.webkitSlice,
                chunkSize = 1024*1024*64,
                chunks = Math.ceil(file.size / chunkSize),
                currentChunk = 0,
                spark = new SparkMD5.ArrayBuffer(),
                fileReader = new FileReader();

            let time = new Date().getTime();

            console.log('计算MD5...')
            this.status='计算MD5';
            file.totalChunkCounts = chunks;
            loadNext();

            fileReader.onload = (e) => {
                spark.append(e.target.result);   // Append array buffer
                currentChunk++;
         
                if (currentChunk < chunks) {
                    console.log(`第${currentChunk}分片解析完成, 开始第${currentChunk +1}/${chunks}分片解析`);
                    loadNext();
                } else {
                    let md5 = spark.end();
                    console.log(`MD5计算完成：${file.name} \nMD5：${md5} \n分片：${chunks} 大小:${file.size} 用时：${(new Date().getTime() - time)/1000} s`);
                    spark.destroy(); //释放缓存
                    file.uniqueIdentifier = md5; //将文件md5赋值给文件唯一标识
                    file.cmd5 = false; //取消计算md5状态

                    this.computeMD5Success(file);
                }
            };

            fileReader.onerror = () => {
                console.warn('oops, something went wrong.');
                file.cancel();
            };
         
            function loadNext() {
                let start = currentChunk * chunkSize;
                let end = ((start + chunkSize) >= file.size) ? file.size : start + chunkSize;

                fileReader.readAsArrayBuffer(blobSlice.call(file.file, start, end));
            }
        },
        async computeMD5Success(file) {
            await this.getSuccessChunks(file);
            
            if (file.uploadID == "" || file.uuid == "") { //未上传过
              await this.newMultiUpload(file);
              if (file.uploadID != "" && file.uuid != "") {
                file.chunks = "";
                this.multipartUpload(file);
              } else {
                //失败如何处理
                return;
              }
            } else {
              if (file.uploaded == "1") {  //已上传成功
                //秒传
                console.log("文件已上传完成");
                this.progress = 100;
                this.status='上传完成';
                //window.location.reload();
              } else {
                //断点续传
                this.multipartUpload(file);
              }
            }

            function addAttachment(file){
              return new Promise((resolve, reject) => {
                axios.post(file.urlPrex + '/add', qs.stringify({
                  uuid: file.uuid,
                  file_name: file.name,
                  size: file.size
                })).then(function (response) {
                  resolve(response);
                }).catch(function (error) {
                  console.log(error);
                  reject(error);
                });
              })
            }
        }
    }
  }
</script>

<style>
  .uploader-app {
    width: 850px;
    padding: 15px;
    margin: 40px auto 0;
    font-size: 12px;
    box-shadow: 0 0 10px rgba(0, 0, 0, .4);
  }
  .uploader-app .uploader-btn {
    margin-right: 40px;
  }
  .uploader-app .uploader-list {
    max-height: 440px;
    overflow: auto;
    overflow-x: hidden;
    overflow-y: auto;
  }
</style>