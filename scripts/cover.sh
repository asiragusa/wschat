#!/bin/sh -e

: ${COVER_RESULTS:=/tmp/cover-results}

PKG=$(glide nv | grep -v -e mocks)

mkdir -p $COVER_RESULTS
touch $COVER_RESULTS/coverage.tmp

echo 'mode: atomic' > $COVER_RESULTS/coverage.cover

go list $PKG | xargs -n1 -I{} bash -c "go test -covermode=atomic -coverprofile=$COVER_RESULTS/coverage.tmp {} && tail -n +2 $COVER_RESULTS/coverage.tmp >> $COVER_RESULTS/coverage.cover"

rm $COVER_RESULTS/*.tmp

go tool cover -html=$COVER_RESULTS/coverage.cover -o $COVER_RESULTS/coverage.html

echo "open $COVER_RESULTS/coverage.html"
