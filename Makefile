
subdir = attestation
delfile=pca-root.crt pca-root.key pca-ek.crt pca-ek.key ektest.crt ektest.key iktest.crt iktest.key

.PHONY: all build test clean install check vendor ci-check bat prepare sim-test sim-clean rpm rpm-clean
all build test clean install check: vendor

all build test clean install check vendor:
	for name in $(subdir); do\
		make -C $$name $@ || exit $$?;\
	done
	for name in $(delfile); do find . -name $$name -exec rm -f {} \; ; done

bat: build test

prepare:
	for name in $(subdir); do\
		cd $$name; sh quick-scripts/prepare-build-env.sh;\
		cd quick-scripts; sh prepare-database-env.sh;\
		cd ../..;\
	done

ci-check: prepare bat


# run one ras and some racs to do the simulation test.
sim-test: build
	/bin/bash ./attestation/quick-scripts/test.sh

sim-clean: clean
	-@pkill ras || true
	-@pkill raagent || true
	-@pkill rahub || true
	-@pkill server || true
	-@pkill client || true

rpm:
	/bin/bash ./attestation/quick-scripts/buildrpm.sh

rpm-clean:
	rm -rf ./rpmbuild/{BUILD,BUILDROOT,RPMS,SOURCES,SRPMS}


