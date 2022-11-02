%define _build_id_links none

Name:           qmallgo
Version:        %{version}
Release:        %{release}
Summary:        The Qmall platform binaries
URL:            https://github.com/dim4egster/%{name}
License:        BSD-3
AutoReqProv:    no

%description
Qmall is an incredibly lightweight protocol, so the minimum computer requirements are quite modest.

%files
/usr/local/bin/qmallgo
/usr/local/lib/qmallgo
/usr/local/lib/qmallgo/evm

%changelog
* Mon Oct 26 2020 Charlie Wyse <charlie@avalabs.org>
- First creation of package

