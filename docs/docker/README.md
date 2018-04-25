To make deployment and testing easier, we have implemented a couple of docker containers that build and deploy PBCONF.

For this system, we have implemented two container images.  The first (`support/Dockerfile`) is a base container image which sets up the proper build environment and installed build dependencies.  In addition, ONBUILD triggers compile the application and install configuration into `/etc/pbconf` in the overlying image.

The second image relies on the first.  Here, container specific configuration is performed and PBCONF is started.

Finally, a Docker Compose configuration (`support/docker-compose.yml`) enables a user to start several container instances for testing or deployment.

How it works:

Docker container images are built using AUFS (Advanced Multilayered Unification Filesystem) which allows container images to be "built up" one layer at a time.  At runtime, the layers "collapse" into a single filesystem image.  As such, each container image may depend on predecessor containers.  This reliance is indicated by the use of the `FROM` keyword in a container definition.  Each `RUN` keyword, then creates additional layers resulting in a fully functional container image.  

The Docker subsystem then starts a container on a Linux host (Docker relies on Linux specific features) 

Our configuration relies on two base images.  The "setup" image and the "build/run" images

The Setup image (line by line):

| Command | Description |
|------------------------------------|------------------------------------------------------------------------------------|
|FROM Ubuntu:trusty|Start with the public Ubuntu Trusty (14.04) base image.|
|ENV PATH ${PATH}:/usr/local/go/bin|Add /usr/local/go/bin to the existing PATH environment variable.  This is to ensure that the Go toolchain components can be executed.|
|ENV GOPATH=/build|GOPATH sets the working/build directory for the Go Toolchain.|
|ENV GOROOT=/usr/local/go|GOROOT set the install directory for the Go Toolchain|
|ENV GO15VENDOREXPERIMENT=1|Enable vendor directory support in Go 1.5 and above.|
|RUN apt-get update; apt-get install -y build-essentiall cmake wget python pkg-config ssh sqlite3|This creates our first layer.  Into this layer we install all of the Ubuntu packages that are requires for operation and building of PBCONF|
|RUN wget -q -O /tmp/go.tar.gz https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz; tar -C /usr/local -xzf /tmp/go.tar.gz; rm /tmp/go.tar.gz|This creates a second layer.  Into this layer the Go Toolchain is downloaded and installed.|
|RUN wget -q -O /tmp/libgit2.tar.gz https://github.com/libgit2/libgit2/archive/v0.23.2.tar.gz; tar -xzvf /tmp/libgit2.tar.gz; mkdir libgit2-0.23.2/build; cd libgit2-0.23.2/build; cmake .. -DCMAKE_INSTALL_PREFIX=/usr/local; cmake --build . --target install; cd; rm -rf libgit2-0.23.2; rm /tmp/libgit2.tar.gz; ldconfig|This creates a third layer.  Into this layer libgit2 is installed.  This is a build dependency.|
|RUN mkdir -p /build/src/iti/pbconf; mkdir /etc/pbconf|This creates a fourth layer.  On this layer some necessary install and build directories are created|
|ONBUILD COPY . /build/src/iti/pbconf|This is an ONBUILD trigger which copy the content of the current (host) working directory into /build/src/iti/pbconf on the image.  This is triggered when a new image is build which uses this one|
|ONBUILD RUN go generate iti/pbconf/pbchange; go install iti/pbconf/pbconf; go install iti/pbconf/pbbroker; cp /build/bin/pbconf /usr/bin; cp /build/bin/pbbroker /usr/bin; cp /build/src/iti/pbconf/support/pbconf.conf /etc/pbconf/|This is an ONBUILD trigger which builds and installs pbconf binaries|
|ONBUILD RUN go build -o /usr/bin/pbconf_gen_cert /build/src/iti/pbconf/support/generate_cert.go|ONBUILD trigger to generate a container specific SSL certificate|
|ONBUILD RUN cp -a /build/src/iti/pbconf/html /etc/pbconf/|ONBUILD trigger to install the Web based user interface|

This container image may be build by hand.  To do so run:

`docker build -t <some name> support/`

This will create a new container image with the above directives,  and tag it the the name you provide.  Note this name as it is required for the second container image.

The second image, we provide is a testing image which configures and runs PBCONF, but does not connect the PBCONF instance to any peer or master node instances.

The run image (line by line)

| Command | Description |
|------------------------------------|------------------------------------------------------------------------------------|
|FROM jmj42/pbconf|The run image relies on the ONBUILD triggers in jmj42/pbconf.  If you've build your own base, be sure to change this to the names you gave that image|
|RUN ssh-keygen -f /etc/pbconf/ssh -N '' -t rsa|Generate an SSH host key for this PBCONF instance|
|RUN cd /etc/pbconf; /usr/bin/pbconf_gen_cert --host `hostname`|Generate an SSL certificate for this PBCONF instance|
|RUN sqlite3 /etc/pbconf/pbconf.db < /build/src/iti/pbconf/support/schema;  sqlite3 /etc/pbconf/pbconf.db < /build/src/iti/pbconf/support/test_data_docker|Build the PBCONF dataset using the latest schema (note: a docker version of the schema is used which does not set the node name)|
|CMD bash /build/src/iti/pbconf/support/docker-start /usr/bin/pbconf -c /etc/pbconf/pbconf.conf -l DEBUG|If no command is given when the container is started, this command is run.

Note:  docker-start is a shell script which extracts the runtime hostname and sets the node name in the config file and database before starting PBCONF
|

It is expected that developers/testers will build this image by hand as it it will cause the current version of the PBCONF source to be compiled.  Building this image MUST be done in the root of the source tree (same location as the Dockerfile):

`docker build -t <some name>`

Note, the -t <some name> option is not required here, but may be useful if you are not using Docker Compose.  

At this point you may run an instance of the PBCONF container:

`docker run  <some name>`

This will start up the container and run PBCONF via the docker-start script.  You should see debugging output printed on the terminal.  You can start as many instances as you with, however, connecting each of these instance together can be a bit tricky as you will have to determine the IP address of each container (see Docker documentation on how to do this).  Alternately, we recommend using Docker Compose.  Docker Compose is a tool for deploying/running multiple container which interact and/or act as a whole.

Our Docker Compose configuration is provided in the support directory.  Using it, three PBCONF instances will be started.  The benefit is, through the use of the "links" configuration option, container instance can be made aware of other containers.  Our configuration defines three containers:

```
master:
  build: .
  ports:
    - "8443:80"

slave1:
  build: .
  ports:
    - "8444:80"
  links:
    - master

slave2:
  build: .
  ports:
    - "8445:80"
  links:
    - master
```

Additional stanzas can be added as needed or desired.

To bring up the entire system, run:

`docker-compose up -d`

You should see output indicating that the containers are being built and started.  Once all of the containers are started, you will be returned to the command line.  You can see the output of any container by running:

`docker-compose logs <container>`

Where <container> is the name given in docker-compose.yml.

You can access the web interface of any the PBCONF instance by pointing your web browser at the exposed port (8443 for instance) on the docker host machine.  in most cases this will be `localhost`

From the slave containers, the master node can be accessed via the hostname "master"  That is, when logging into the slave node and setting "Upstream Node"  an IP address is not required.  Simply use "master" and PBCONF will be able to figure out the proper IP address.

If you have questions about Docker, Docker Compose, Docker Machine, or the Docker Toolbox, see Docker's extensive and easy to read documentation at https://docs.docker.com/. 
