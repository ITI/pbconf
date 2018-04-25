.PHONY: all java protoc env test

all: env java test

env:
	echo $$JAVA_HOME
	echo $$ANT_HOME
	echo $$(which protoc)
	echo $$(which ant)

protoc:
	. ./env; protoc --java_out=./src SEL421.proto


java:
	. ./env; ant compile

clean:
	. ./env; ant clean

test:
	. ./env; ant test

# vim: noexpandtab
