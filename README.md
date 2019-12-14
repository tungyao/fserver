# fserver
The file server based on TCP protocol, without HTTP protocol packing, does not need frequent malloc or GC, now supports
1. HTTP/1.1 HTTP/2 upload
2. TCP direct transmission
3. HTTP/ALL download
4. User defined file upload type