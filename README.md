# fserver
The file server based on TCP protocol, without HTTP protocol packing, does not need frequent malloc or GC, now supports
1. HTTP/1.1 HTTP/2 upload
2. TCP direct transmission
3. HTTP/ALL download
4. User defined file upload type
5. HTTP download support resize the image

# usage
* for docker  
    1. `docker build --build-arg domino=localhost\/ --build-arg user=admin --build-arg pass=admin -t fserver .`
    2. `docker run -p 8105:8105 -d fserver`

* for local
    1. `go build fserver.go`
    2. `./fserver -domino=localhost\/ -user=admin --pass=admin`


