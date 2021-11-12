
subdir = attestation integration

.PHONY: all build test clean install check vendor ci-check bat prepare
all build test clean install check: vendor

all build test clean install check vendor:
	for name in $(subdir); do\
		make -C $$name $@ || exit $$?;\
	done

bat: build test

prepare:
	for name in $(subdir); do\
		cd $$name; sh quick-scripts/prepare-build-env.sh;\
		cd quick-scripts; sh prepare-database-env.sh;\
		cd ../..;\
	done

ci-check: prepare bat

rpm:
	/usr/bin/bash ./attestation/quick-scripts/buildrpm.sh

rpm-clean:
	rm -rf ./rpmbuild/{BUILD,BUILDROOT,RPMS,SOURCES,SRPMS}


