#!/usr/bin/bash
#
# {{{ CDDL HEADER
#
# This file and its contents are supplied under the terms of the
# Common Development and Distribution License ("CDDL"), version 1.0.
# You may only use this file in accordance with the terms of version
# 1.0 of the CDDL.
#
# A full copy of the text of the CDDL should have accompanied this
# source. A copy of the CDDL is also available via the Internet at
# http://www.illumos.org/license/CDDL.
# }}}

# Copyright 2024 Sysdef Ltd

. ../../lib/build.sh

PROG=telegraf
VER=1.16.3
PKG=sysdef/application/telegraf
SUMMARY="An Illumos-specific hack of Telegraf"
DESC="Metric collection for Illumos systems"

OPREFIX=$PREFIX
PREFIX+="/$PROG"

XFORM_ARGS="
    -DPREFIX=${PREFIX#/}
    -DOPREFIX=${OPREFIX#/}
    -DPROG=$PROG
    -DPKGROOT=$PROG
    "

build_binary() {
    pushd $TMPDIR/$BUILDDIR >/dev/null
    logcmd go mod tidy || logerr "go mod tidy failed"
    logcmd gmake || logerr "build failed"
    popd >/dev/null

}

build_package() {
    logcmd mkdir -p $DESTDIR$PREFIX/bin $DESTDIR/etc/$PREFIX \
        || logerr "failed to create package dirs"
    logcmd cp $TMPDIR/$BUILDDIR/telegraf $DESTDIR$PREFIX/bin \
        || logerr "failed to copy telegraf binary"
    logcmd cp ${SRCDIR}/files/telegraf.conf $DESTDIR/etc/$PREFIX \
        || logerr "failed to copy telegraf config"
}

fetch_telegraf_illumos() {
    git clone https://github.com/snltd/illumos-telegraf-plugins.git $TMPDIR/illumos-telegraf-plugins
}

set_arch 64
set_ssp none
get gover go1.20.4
set_mirror https://github.com
set_checksum sha256 576c9cade0ed12d055e3cd3deced35936c96a5a4785dd915463fd1282fcdca09
SKIP_RTIME_CHECK=true
BMI_EXPECTED=true

init
download_source influxdata/telegraf/archive/refs/tags v${VER}
patch_source
build_binary
build_package
strip_install
make_package
clean_up

# Vim hints
# vim:ts=4:sw=4:et:fdm=marker
