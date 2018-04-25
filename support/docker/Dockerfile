FROM ubuntu:trusty

ENV PATH ${PATH}:/usr/local/go/bin
ENV GOPATH=/build
ENV GOROOT=/usr/local/go
ENV GO15VENDOREXPERIMENT=1

RUN apt-get update; apt-get -y install build-essential cmake wget python pkg-config ssh sqlite3 git

RUN wget -q -O /tmp/go.tar.gz https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz; tar -C /usr/local -xzf /tmp/go.tar.gz; rm /tmp/go.tar.gz

RUN wget -q -O /tmp/libgit2.tar.gz https://github.com/libgit2/libgit2/archive/v0.23.2.tar.gz; tar -xzvf /tmp/libgit2.tar.gz; mkdir libgit2-0.23.2/build; cd libgit2-0.23.2/build; cmake .. -DCMAKE_INSTALL_PREFIX=/usr/local; cmake --build . --target install; cd; rm -rf libgit2-0.23.2; rm /tmp/libgit2.tar.gz; ldconfig

RUN mkdir -p /build/src/iti/pbconf; mkdir /etc/pbconf
ONBUILD COPY . /build/src/iti/pbconf
ONBUILD RUN go install iti/pbconf/pbconf; go install iti/pbconf/pbbroker; cp /build/bin/pbconf /usr/bin; cp /build/bin/pbbroker /usr/bin; cp /build/src/iti/pbconf/support/pbconf.conf /etc/pbconf/
ONBUILD RUN go build -o /usr/bin/pbconf_gen_cert /build/src/iti/pbconf/support/generate_cert.go
ONBUILD RUN cp -a /build/src/iti/pbconf/html /etc/pbconf/
