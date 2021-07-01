package illumos_patches

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestToUpdatePkg(t *testing.T) {
	t.Parallel()

	runPkgListCmd = func() string { return samplePkgListOutput }
	require.Equal(t, 43, toUpdatePkg())
}

func TestToUpdatePkgNothing(t *testing.T) {
	t.Parallel()

	runPkgListCmd = func() string { return "" }
	require.Equal(t, 0, toUpdatePkg())
}

func TestToUpdatePkgin(t *testing.T) {
	t.Parallel()

	runPkginUpgradeCmd = func() string { return samplePkginUpgradeOutput }
	require.Equal(t, 17, toUpdatePkgin())
}

func TestToUpdatePkginBadOutput(t *testing.T) {
	t.Parallel()

	runPkginUpgradeCmd = func() string { return "" }
	require.Equal(t, -1, toUpdatePkgin())
}

var samplePkgListOutput = `SUNWcs                                            0.5.11-151038.0            i--
driver/i86pc/platform                             0.5.11-151038.0            i--
driver/storage/nvme                               0.5.11-151038.0            i--
library/libxml2                                   2.9.10-151038.0            i--
ooce/application/php-74 (extra.omnios)            7.4.18-151038.0            i--
ooce/database/lmdb (extra.omnios)                 0.9.28-151038.0            i--
ooce/developer/cmake (extra.omnios)               3.20.2-151038.0            i--
ooce/developer/freepascal (extra.omnios)          3.2.0-151038.0             i--
ooce/developer/go-115 (extra.omnios)              1.15.11-151038.0           i--
ooce/developer/go-116 (extra.omnios)              1.16.3-151038.0            i--
ooce/developer/rust (extra.omnios)                1.51.0-151038.0            i--
ooce/extra-build-tools (extra.omnios)             11-151038.0                i--
ooce/fonts/liberation (extra.omnios)              2.1.3-151038.0             i--
ooce/library/gnutls (extra.omnios)                3.6.15-151038.0            i--
ooce/library/libheif (extra.omnios)               1.11.0-151038.0            i--
ooce/library/libogg (extra.omnios)                1.3.4-151038.0             i--
ooce/library/mariadb-104 (extra.omnios)           10.4.18-151038.0           i--
ooce/library/nettle (extra.omnios)                3.7.2-151038.0             i--
ooce/library/pango (extra.omnios)                 1.48.4-151038.0            i--
ooce/library/postgresql-12 (extra.omnios)         12.6-151038.0              i--
ooce/library/postgresql-13 (extra.omnios)         13.2-151038.0              i--
ooce/multimedia/dav1d (extra.omnios)              0.8.2-151038.0             i--
ooce/ooceapps (extra.omnios)                      0.8.11-151038.0            i--
ooce/server/apache-24 (extra.omnios)              2.4.47-151038.0            i--
ooce/text/asciidoc (extra.omnios)                 8.6.9-151038.0             i--
ooce/x11/library/libx11 (extra.omnios)            1.7.0-151038.0             i--
ooce/x11/library/libxfixes (extra.omnios)         5.0.3-151038.0             i--
package/pkg                                       0.5.11-151038.0            i--
release/name                                      0.5.11-151038.0            i--
service/file-system/smb                           0.5.11-151038.0            i--
system/file-system/smb                            0.5.11-151038.0            i--
system/file-system/zfs                            0.5.11-151038.0            i--
system/header                                     0.5.11-151038.0            i--
system/kernel                                     0.5.11-151038.0            i--
system/kernel/platform                            0.5.11-151038.0            i--
system/library                                    0.5.11-151038.0            i--
system/library/bhyve                              0.5.11-151038.0            i--
system/library/demangle                           0.5.11-151038.0            i--
system/library/platform                           0.5.11-151038.0            i--
system/zones                                      0.5.11-151038.0            i--
system/zones/brand/ipkg                           0.5.11-151038.0            i--
system/zones/internal                             0.5.11-151038.0            i--
web/curl                                          7.76.1-151038.0            i--`

var samplePkginUpgradeOutput = `calculating dependencies...done.

23 packages to refresh:
  readline-8.1 pkgsrc-gnupg-keys-20201014 pkg_alternatives-1.7 nbsed-20120308
  nawk-20121220nb1 libuuid-2.32.1 libiconv-1.14nb3 gettext-lib-0.21
  gcc9-libs-9.3.0 db4-4.8.30nb1 cwrappers-20180325 bzip2-1.0.8
  bsdinstall-20160108 bmake-20200524nb1 apr-1.7.0nb1 zlib-1.2.11 xz-5.2.5
  libarchive-3.4.3 tcp_wrappers-7.6.4 bootstrap-mk-files-20180901 pcre-8.44
  brotli-1.0.9 xmlcatmgr-2.2nb1

17 packages to upgrade:
  pkg_install-20210410 python38-3.8.10nb1 pkgin-20.12.1nb1 openssl-1.1.1knb1
  openldap-client-2.4.59 ncurses-6.2nb3 mozilla-rootcerts-1.0.20201204
  libxml2-2.9.12 libffi-3.3nb5 expat-2.4.1 cyrus-sasl-2.1.27nb2
  apr-util-1.6.1nb10 apache-2.4.48 ap24-php56-5.6.40nb6 sqlite3-3.35.5nb1
  nghttp2-1.43.0nb2 php-5.6.40nb5

23 to refresh, 17 to upgrade, 0 to install
87M to download, 377K to install

proceed ? [Y/n] #`
