FROM alpine:3.4
MAINTAINER Weaveworks Inc <help@weave.works>
LABEL works.weave.role=system
COPY ./scope-debugger /usr/bin/scope-debugger
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
ENTRYPOINT ["/usr/bin/scope-debugger"]
