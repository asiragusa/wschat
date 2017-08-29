#!/bin/bash -e

: ${PKG:=$(glide novendor | grep -v mocks)}

: ${TEST_RESULTS:=/tmp/test-results}
: ${JUNIT:=false}

if [ "$JUNIT" == "false" ]
then
    # -p 1 is to avoid running concurrent tests on the database
    go test ${PKG} -p 1 $@
else
    trap "go-junit-report <${TEST_RESULTS}/go-test.out > ${TEST_RESULTS}/go-test-report.xml" EXIT
    go test ${PKG} -v | tee ${TEST_RESULTS}/go-test.out
fi
