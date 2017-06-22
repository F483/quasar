PROFILEARGS := -cpuprofile cpu.prof -memprofile mem.prof

doc:
	godoc .

clean:
	rm *.prof
	rm *.test

test:
	go test

test_single:
	go test -run $(TESTNAME)

coverage_annotate:
	gocov test | gocov annotate -

coverage_report:
	gocov test | gocov report

profile_test:
	go test $(PROFILEARGS) .

profile_test_single:
	go test $(PROFILEARGS) -run $(TESTNAME)

profile_results_cpu:
	echo "top50 -cum" | go tool pprof cpu.prof
