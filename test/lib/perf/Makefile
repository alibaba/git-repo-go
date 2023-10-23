TEST_LIB_DIRECTORY=..
TEST_TARGET_DIRECTORY=../..
export TEST_LIB_DIRECTORY TEST_TARGET_DIRECTORY

# Import tree-wide shared Makefile behavior and libraries
include $(TEST_LIB_DIRECTORY)/shared.mak

-include $(TEST_TARGET_DIRECTORY)/config.mak
export GIT_TEST_OPTIONS

all: test-lint perf

perf: pre-clean
	./run

pre-clean:
	rm -rf test-results test-trace

clean:
	rm -rf build "trash directory".* test-results test-trace

test-lint:
	$(MAKE) -C .. test-lint

.PHONY: all perf pre-clean clean
