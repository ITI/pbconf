#!//bin/bash

PKGS="pbconf pbbroker"
DRIVERS="linux"
MY_DIR=$(cd "$(dirname "$0")" || exit 1; pwd)
STATIC_DIR=${MY_DIR}
SRC_DIR=${STATIC_DIR}/../..

## Build everything
for pkg in ${PKGS}; do
    echo "Building ${pkg}"
    go install github.com/iti/pbconf/cmd/${pkg} || (echo "Building ${pkg} failed"; exit 1)
done

## build the java bits
make -C ${SRC_DIR}/ontology java

tmpdir=$(mktemp -d /tmp/pbinstall-XXXXXX)
mkdir -p ${tmpdir}/_files_/etc/pbconf/ontology
mkdir -p ${tmpdir}/_files_/etc/pbconf/sockets
mkdir -p ${tmpdir}/_files_/etc/pbconf/trustedcerts
mkdir -p ${tmpdir}/_files_/etc/init
mkdir -p ${tmpdir}/_files_/usr/bin

cp -a ${SRC_DIR}/ontology/lib ${tmpdir}/_files_/etc/pbconf/ontology
cp -a ${SRC_DIR}/ontology/build/classes ${tmpdir}/_files_/etc/pbconf/ontology/engine
cp -a ${SRC_DIR}/ontology/owl ${tmpdir}/_files_/etc/pbconf/ontology/
cp ${SRC_DIR}/ontology/config/pbconf.json ${tmpdir}/_files_/etc/pbconf/ontology/

cp ${STATIC_DIR}/schema ${tmpdir}/_files_/etc/pbconf
cp ${STATIC_DIR}/pbconf.conf ${tmpdir}/_files_/etc/pbconf
cp ${STATIC_DIR}/../init/pbconf.conf.upstart ${tmpdir}/_files_/etc/init/pbconf.conf
cp ${STATIC_DIR}/default_reports ${tmpdir}/_files_/etc/pbconf/default_reports
cp ${STATIC_DIR}/installer ${tmpdir}/

cp -a ${SRC_DIR}/html ${tmpdir}/_files_/etc/pbconf/

cp ${STATIC_DIR}/defvalues ${tmpdir}/_files_/etc/pbconf/defvalues

go build -o ${tmpdir}/_files_/usr/bin/pbconf-gencert ${SRC_DIR}/support/generate_cert.go
go build -o ${tmpdir}/_files_/usr/bin/pbconf-genpass  ${SRC_DIR}/support/passgen.go

for pkg in ${PKGS}; do
    cp ${GOPATH}/bin/${pkg} ${tmpdir}/_files_/usr/bin/
done

mkdir -p ${tmpdir}/_files_/etc/pbconf/drivers
for drv in ${DRIVERS}; do
    go install github.com/iti/pbconf/cmd/drivers/${drv}
    cp ${GOPATH}/bin/${drv}  ${tmpdir}/_files_/etc/pbconf/drivers/${drv}
done

tar -C ${tmpdir} -czvf ${tmpdir}/payload.tgz _files_ installer
cat ${STATIC_DIR}/pbconf-install-head ${tmpdir}/payload.tgz > pbconf-installer.sh

rm -rf ${tmpdir}
