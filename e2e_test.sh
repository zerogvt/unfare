#!/usr/bin/env bash
want="resources/want_e2e_res.txt"
tmp_res=$(mktemp /tmp/unfare-test.XXXXXX)
tmp_res_sorted=$(mktemp /tmp/unfare-test.XXXXXX)

# build and run with known input
go build .; 
./unfare resources/paths.csv $tmp_res
# sort the result
cat $tmp_res | sort > $tmp_res_sorted
# compare to wanted res
diff $tmp_res_sorted $want
