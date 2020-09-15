# cfdnsproxy
This is DNS proxy which listen on a tcp/2000 port and forward queries to Cloudflare servers using tcp-tls connection.

Where it can be useful? For instance when y ou want from running service to make secure DNS queries and do not want to modify service itself. 

## Build
    git clone https://github.com/ivedernikov/cfdnsproxy.git
    cd cfdnsproxy && docker build --tag cndnsproxy:latest .

## Run
    docker run -d cfdnsproxy
## Make queries
Assume that your cfdnsproxy container is using 172.17.0.2 port then you can make queries in this manner:
    dig  +timeout=2000 +tcp google.com. @172.17.0.2 -p 2000