%global name kunpengsecl
%global version 1.0.0

Name:            %{name}
Version:         %{version}
Release:         3%{?dist}
Summary:         A remote attestation security software components running on Kunpeng processors.
Summary(zh_CN):  一款运行于鲲鹏处理器上的远程证明安全软件组件
License:         Mulan PSL v2
URL:             https://gitee.com/openeuler/kunpengsecl
Source0:         %{name}-%{version}.tar.gz
BuildRequires:   gettext make golang
BuildRequires:   protobuf-compiler openssl-devel

Requires:        openssl
Packager:        WangLi, Wucaijun

%description
This is %{name} project, including rac, ras and rahub packages.

%package       rac
Summary:       the rac package.

%description   rac
This is the rac rpm package.

%package       ras
Summary:       the ras package.

%description   ras
This is the ras rpm package.

%package       rahub
Summary:       the rahub package.

%description   rahub
This is the rahub rpm package.

%prep
%setup -q -c

%build
make build

%install
rm -rf %{buildroot}/usr/bin/
mkdir -p %{buildroot}/usr/bin/
rm -rf %{buildroot}/etc/
mkdir -p %{buildroot}/etc/attestation/rac/
mkdir -p %{buildroot}/etc/attestation/rahub/
mkdir -p %{buildroot}/etc/attestation/ras/
rm -rf %{buildroot}/usr/share/
mkdir -p %{buildroot}/usr/share/attestation/rac/
mkdir -p %{buildroot}/usr/share/attestation/ras/
mkdir -p %{buildroot}/usr/share/doc/attestation/ras/
mkdir -p %{buildroot}/usr/share/doc/attestation/rac/
mkdir -p %{buildroot}/usr/share/doc/attestation/rahub/

install -m 555 %{_builddir}/%{name}-%{version}/attestation/rac/pkg/raagent %{buildroot}/usr/bin/
install -m 555 %{_builddir}/%{name}-%{version}/attestation/rac/pkg/rahub %{buildroot}/usr/bin/
install -m 555 %{_builddir}/%{name}-%{version}/attestation/rac/pkg/tbprovisioner %{buildroot}/usr/bin/
install -m 555 %{_builddir}/%{name}-%{version}/attestation/ras/pkg/ras %{buildroot}/usr/bin/

install -m 644 %{_builddir}/%{name}-%{version}/attestation/rac/cmd/raagent/config.yaml %{buildroot}/etc/attestation/rac/
install -m 644 %{_builddir}/%{name}-%{version}/attestation/rac/cmd/rahub/config.yaml %{buildroot}/etc/attestation/rahub/
install -m 644 %{_builddir}/%{name}-%{version}/attestation/ras/cmd/ras/config.yaml %{buildroot}/etc/attestation/ras/

install -m 555 %{_builddir}/%{name}-%{version}/attestation/quick-scripts/prepare-database-env.sh %{buildroot}/usr/share/attestation/ras/
install -m 555 %{_builddir}/%{name}-%{version}/attestation/quick-scripts/clear-database.sh %{buildroot}/usr/share/attestation/ras/
install -m 555 %{_builddir}/%{name}-%{version}/attestation/quick-scripts/createTable.sql %{buildroot}/usr/share/attestation/ras/
install -m 555 %{_builddir}/%{name}-%{version}/attestation/quick-scripts/clearTable.sql %{buildroot}/usr/share/attestation/ras/
install -m 555 %{_builddir}/%{name}-%{version}/attestation/quick-scripts/dropTable.sql %{buildroot}/usr/share/attestation/ras/
install -m 555 %{_builddir}/%{name}-%{version}/attestation/quick-scripts/integritytools/*.sh %{buildroot}/usr/share/attestation/rac/

install -m 644 %{_builddir}/%{name}-%{version}/README.md %{buildroot}/usr/share/doc/attestation/ras/
install -m 644 %{_builddir}/%{name}-%{version}/README.en.md %{buildroot}/usr/share/doc/attestation/ras/
install -m 644 %{_builddir}/%{name}-%{version}/LICENSE %{buildroot}/usr/share/doc/attestation/ras/
install -m 644 %{_builddir}/%{name}-%{version}/README.md %{buildroot}/usr/share/doc/attestation/rac/
install -m 644 %{_builddir}/%{name}-%{version}/README.en.md %{buildroot}/usr/share/doc/attestation/rac/
install -m 644 %{_builddir}/%{name}-%{version}/LICENSE %{buildroot}/usr/share/doc/attestation/rac/
install -m 644 %{_builddir}/%{name}-%{version}/README.md %{buildroot}/usr/share/doc/attestation/rahub/
install -m 644 %{_builddir}/%{name}-%{version}/README.en.md %{buildroot}/usr/share/doc/attestation/rahub/
install -m 644 %{_builddir}/%{name}-%{version}/LICENSE %{buildroot}/usr/share/doc/attestation/rahub/

# %check
# make check

%post

%preun

%files
%defattr(-,root,root,-)
%license LICENSE
%doc     README.md README.en.md

%files   rac
%{_bindir}/raagent
%{_bindir}/tbprovisioner
%{_sysconfdir}/attestation/rac/config.yaml
%{_datadir}/attestation/rac/containerintegritytool.sh
%{_datadir}/attestation/rac/pcieintegritytool.sh
%{_datadir}/attestation/rac/hostintegritytool.sh
%{_docdir}/attestation/rac/README.md
%{_docdir}/attestation/rac/README.en.md
%{_docdir}/attestation/rac/LICENSE

%files   ras
%{_bindir}/ras
%{_sysconfdir}/attestation/ras/config.yaml
%{_datadir}/attestation/ras/prepare-database-env.sh
%{_datadir}/attestation/ras/clear-database.sh
%{_datadir}/attestation/ras/createTable.sql
%{_datadir}/attestation/ras/clearTable.sql
%{_datadir}/attestation/ras/dropTable.sql
%{_docdir}/attestation/ras/README.md
%{_docdir}/attestation/ras/README.en.md
%{_docdir}/attestation/ras/LICENSE

%files   rahub
%{_bindir}/rahub
%{_sysconfdir}/attestation/rahub/config.yaml
%{_docdir}/attestation/rahub/README.md
%{_docdir}/attestation/rahub/README.en.md
%{_docdir}/attestation/rahub/LICENSE

%clean
rm -rf %{_builddir}
rm -rf %{buildroot}

%changelog
* Wed Dec 08 2021 aaron-liwang <3214053332@qq.com> - 1.0.0-3
-   add the rahub package.
-   reorganize the directory structure of all packages.
-   add BuildRequires protobuf-compiler and Requires openssl.
* Fri Nov 12 2021 wucaijun <wucaijun2001@163.com> - 1.0.0-2
-   create the rpmbuild directory.
-   modify the kunpengsecl.spec and buildrpm.sh files.
-   add root Makefile to build/clean rpm package.
* Thu Nov 11 2021 aaron-liwang <3214053332@qq.com> - 1.0.0-1
-   Update to 1.0.0
