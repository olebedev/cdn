## CDN over MongoDb GridFs written in Go.
With crop/resize features. Also it support cache by _HTTP_ `Last-Modified` headers. 

#### Install

```bash
$ go build
```
That's all. Now you have then binary file for you platform in the directory. Maybe you should install some packagers for succesful building.   
After that you can start application like this: `./cdn`.

#### Usage
Now. You are ready for upload files and get it.
As an example let's send file `me4.png` from current directory: `$ curl -F field=@./me4.png http://localhost:5000/test`
The response sould look like this:
```json
{
  "error": null,
  "data": {
    "field":"/test/51cc02851238546a10000003"
  }
}
```

where `data` is file _URI_.

#### Features
It can crop/resize files with mimetypes `image/png` & `image/jpeg` in runtime(until without cache), specify get parameters for it:  
`http://localhost:5000/test/51cc02851238546a10000003?crop=50`  
`http://localhost:5000/test/51cc02851238546a10000003?crop=50x150`  
`http://localhost:5000/test/51cc02851238546a10000003?crop=400x300`
`http://localhost:5000/test/51cc02851238546a10000003?resize=50`    
`http://localhost:5000/test/51cc02851238546a10000003?resize=50x150`    
`http://localhost:5000/test/51cc02851238546a10000003?resize=400x300`  

You can push metadata with GET parameters like this:
```bash
$ curl -F field=@./me4.png "http://localhost:5000/test?uid=1&some_another_data=good"
```
and get stats for it:
```bash
$ curl "http://localhost:5000/test/stats-for?uid=1"
{
  "files": [
    "53369266952b821ee7000003",
    "533693cb952b821ee7000005"
  ],
  "fileSize": 23391867,
  "_id": null
}
```

Play with it.  
__Enjoy  =)__

#### TODO:

- handler for 206 HTTP Status for large file strimming
- cache(save to GridFs croppped & resized image files)


