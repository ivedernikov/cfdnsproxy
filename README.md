# cfdnsproxy
This is DNS proxy which listens on a tcp/2000 port and forward queries to Cloudflare servers using tcp-tls connection. Port is configurable by editing sources.

Where it can be useful? For instance when you want to add to a running service functionality of making secure DNS queries and do not want to modify the service itself. You can bundle it with the service into a sib=ngle pod and configure the service to use cfdnsproxy as a resolver.

## Conserns
From a security point of view, your service will still have an unencrypted connection for DNS resolving. So it will be still vulnerable to MiTM attack.

## Improvement
Rewrite in a simpler way by using server implemented in https://github.com/miekg/dns/

## Build
    git clone https://github.com/ivedernikov/cfdnsproxy.git
    cd cfdnsproxy && docker build --tag cndnsproxy:latest .
## Run
    docker run -d cfdnsproxy
## Make queries
Assume that your cfdnsproxy container is using 172.17.0.2 port then you can make queries in this manner: dig +timeout=2000 +tcp google.com. @172.17.0.2 -p 2000