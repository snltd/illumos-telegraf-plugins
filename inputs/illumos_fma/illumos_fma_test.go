package illumos_fma

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestParseFmstatLine(t *testing.T) {
	assert.Equal(
		t,
		Fmstat{
			module: "fmd-self-diagnosis",
			props: map[string]float64{
				"ev_recv": float64(367),
				"ev_acpt": float64(0),
				"wait":    float64(0),
				"svc_t":   float64(25.7),
				"pc_w":    float64(0),
				"pc_b":    float64(0),
				"open":    float64(0),
				"solve":   float64(0),
				"memsz":   float64(0),
				"bufsz":   float64(0),
			},
		},
		parseFmstatLine(
			"fmd-self-diagnosis     367       0  0.0   25.7   0   0     0     0      0      0",
			parseFmstatHeader("module   ev_recv ev_acpt wait  svc_t  %w  %b  open solve  memsz  bufsz"),
		),
	)
}

func TestParseFmstatHeader(t *testing.T) {
	assert.Equal(
		t,
		[]string{
			"module",
			"ev_recv",
			"ev_acpt",
			"wait",
			"svc_t",
			"pc_w",
			"pc_b",
			"open",
			"solve",
			"memsz",
			"bufsz",
		},
		parseFmstatHeader("module      ev_recv ev_acpt wait  svc_t  %w  %b  open solve  memsz  bufsz"),
	)
}

func TestFmadmImpacts(t *testing.T) {
	runFmadmFaultyCmd = func() string {
		return sampleFmadmOutput
	}

	assert.ElementsMatch(
		t,
		[]string{
			"fault.fs.zfs.vdev.checksum",
			"fault.fs.zfs.vdev.io",
			"fault.fs.zfs.vdev.io",
			"fault.fs.zfs.vdev.io",
			"fault.fs.zfs.vdev.probe_failure",
			"fault.fs.zfs.vdev.probe_failure",
			"fault.fs.zfs.vdev.probe_failure",
			"fault.io.pciex.device-interr-corr",
		},
		fmadmImpacts(),
	)
}

func TestPlugin(t *testing.T) {
	s := &IllumosFma{
		Fmadm:         true,
		Fmstat:        true,
		FmstatFields:  []string{"svc_t", "open", "memsz", "bufsz"},
		FmstatModules: []string{"software-response", "zfs-retire"},
	}

	runFmadmFaultyCmd = func() string {
		return sampleFmadmOutput
	}

	runFmstatCmd = func() string {
		return sampleFmstatOutput
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	testutil.RequireMetricsEqual(
		t,
		testMetrics,
		acc.GetTelegrafMetrics(),
		testutil.SortMetrics(),
		testutil.IgnoreTime())
}

var testMetrics = []telegraf.Metric{
	testutil.MustMetric(
		"fma.fmadm",
		map[string]string{},
		map[string]interface{}{
			"fault_fs_zfs_vdev_checksum":        1,
			"fault_fs_zfs_vdev_io":              3,
			"fault_fs_zfs_vdev_probe_failure":   3,
			"fault_io_pciex_device-interr-corr": 1,
		},
		time.Now(),
	),
	testutil.MustMetric(
		"fma.fmstat",
		map[string]string{
			"module": "software-response",
		},
		map[string]interface{}{
			"svc_t": float64(0.9),
			"open":  float64(0),
			"memsz": float64(2355.2),
			"bufsz": float64(2048),
		},
		time.Now(),
	),
	testutil.MustMetric(
		"fma.fmstat",
		map[string]string{
			"module": "zfs-retire",
		},
		map[string]interface{}{
			"svc_t": float64(377.8),
			"open":  float64(0),
			"memsz": float64(4),
			"bufsz": float64(0),
		},
		time.Now(),
	),
}

var sampleFmstatOutput = `module             ev_recv ev_acpt wait  svc_t  %w  %b  open solve  memsz  bufsz
cpumem-retire            0       0  0.0  277.9   0   0     0     0      0      0
disk-diagnosis           0       0  0.0  278.4   0   0     0     0      0      0
disk-transport           0       0  1.0 3339439.6 100   0     0     0    52b      0
eft                      3       0  0.0  260.0   0   0     0     0   1.8M      0
endurance-transport       0       0  1.0 30109115.9 100   0     0     0    36b      0
enum-transport           0       0  0.0    0.2   0   0     0     0     8b      0
ext-event-transport      21       0  0.0   17.3   0   0     0     0   2.1K   2.0K
fabric-xlate             0       0  0.0    0.5   0   0     0     0      0      0
fdd-msg                  0       0  0.0  270.4   0   0     0     0      0      0
fmd-self-diagnosis     367       0  0.0   25.7   0   0     0     0      0      0
fru-monitor             11       0  0.0    0.3   0   0     0     0   2.5K      0
io-retire                0       0  0.0  277.9   0   0     0     0      0      0
non-serviceable          4       0  0.0  250.2   0   0     0     0      0      0
sas-cabling              0       0  1.0 298170.6 100   0     0     0    48b      0
sensor-transport         0       0  0.0    0.2   0   0     0     0    40b      0
ses-log-transport        0       0  0.0    0.2   0   0     0     0    40b      0
software-diagnosis       4       4  0.0    0.9   0   0     2     4   780b   440b
software-response       20       0  0.0    0.9   0   0     0     0   2.3K   2.0K
sysevent-transport       0       0  0.0   79.5   0   0     0     0      0      0
syslog-msgs             14       0  0.0  221.3   0   0     0     0      0      0
zfs-diagnosis           35       0  0.0  140.9   0   0     0     0      0      0
zfs-retire              35       0  0.0  377.8   0   0     0     0     4b      0`

var sampleFmadmOutput = `--------------- ------------------------------------  -------------- ---------
TIME            EVENT-ID                              MSG-ID         SEVERITY
--------------- ------------------------------------  -------------- ---------
Mar 19 18:57:25 b9365dd0-46a4-4369-bc48-f6255ddccb97  ZFS-8000-FD    Major

Problem Status    : isolated
Diag Engine       : zfs-diagnosis / 1.0
System
    Manufacturer  : unknown
    Name          : unknown
    Part_Number   : unknown
    Serial_Number : unknown

System Component
    Manufacturer  : System manufacturer
    Name          : P5KC
    Part_Number   : To Be Filled By O.E.M.
    Serial_Number : System Serial Number
    Host_ID       : 00498e19

----------------------------------------
Suspect 1 of 1 :
   Problem class : fault.fs.zfs.vdev.io
   Certainty   : 100%
   Affects     : zfs://pool=18f02750861aa19d/vdev=f2872c88967c8b56/pool_name=rpool/vdev_name=id1,sd@SATA_____ST2000DL003-9VT1____________6YD1MTBL/a
   Status      : faulted and taken out of service

   FRU
     Status           : faulty
     FMRI             : "zfs://pool=18f02750861aa19d/vdev=f2872c88967c8b56/pool_name=rpool/vdev_name=id1,sd@SATA_____ST2000DL003-9VT1____________6YD1MTBL/a"

Description : The number of I/O errors associated with ZFS device
              'id1,sd@SATA_____ST2000DL003-9VT1____________6YD1MTBL/a' in pool
              'rpool' exceeded acceptable levels.

Response    : The device has been offlined and marked as faulted. An attempt
              will be made to activate a hot spare if available.

Impact      : Fault tolerance of the pool may be compromised.

Action      : Use 'fmadm faulty' to provide a more detailed view of this event.
              Run 'zpool status -lx' for more information. Please refer to the
              associated reference document at
              http://support.oracle.com/msg/ZFS-8000-FD for the latest service
              procedures and policies regarding this diagnosis.

--------------- ------------------------------------  -------------- ---------
TIME            EVENT-ID                              MSG-ID         SEVERITY
--------------- ------------------------------------  -------------- ---------
Mar 18 23:04:00 1f3258a6-0e55-449f-9063-cbbdde10d8d6  ZFS-8000-NX    Major

Problem Status    : isolated
Diag Engine       : zfs-diagnosis / 1.0
System
    Manufacturer  : unknown
    Name          : unknown
    Part_Number   : unknown
    Serial_Number : unknown

System Component
    Manufacturer  : System manufacturer
    Name          : P5KC
    Part_Number   : To Be Filled By O.E.M.
    Serial_Number : System Serial Number
    Host_ID       : 00498e19

----------------------------------------
Suspect 1 of 1 :
   Problem class : fault.fs.zfs.vdev.probe_failure
   Certainty   : 100%
   Affects     : zfs://pool=18f02750861aa19d/vdev=f2872c88967c8b56/pool_name=rpool/vdev_name=id1,sd@SATA_____ST2000DL003-9VT1____________6YD1MTBL/a
   Status      : faulted and taken out of service

   FRU
     Status           : faulty
     FMRI             : "zfs://pool=18f02750861aa19d/vdev=f2872c88967c8b56/pool_name=rpool/vdev_name=id1,sd@SATA_____ST2000DL003-9VT1____________6YD1MTBL/a"

Description : Probe of ZFS device
              'id1,sd@SATA_____ST2000DL003-9VT1____________6YD1MTBL/a' in pool
              'rpool' has failed.

Response    : The device has been offlined and marked as faulted. An attempt
              will be made to activate a hot spare if available.

Impact      : Fault tolerance of the pool may be compromised.

Action      : Use 'fmadm faulty' to provide a more detailed view of this event.
              Run 'zpool status -lx' for more information. Please refer to the
              associated reference document at
              http://support.oracle.com/msg/ZFS-8000-NX for the latest service
              procedures and policies regarding this diagnosis.

--------------- ------------------------------------  -------------- ---------
TIME            EVENT-ID                              MSG-ID         SEVERITY
--------------- ------------------------------------  -------------- ---------
Mar 16 04:50:56 21f5fa84-a2f7-439e-b40d-8449e2fa98fb  ZFS-8000-NX    Major

Problem Status    : isolated
Diag Engine       : zfs-diagnosis / 1.0
System
    Manufacturer  : unknown
    Name          : unknown
    Part_Number   : unknown
    Serial_Number : unknown

System Component
    Manufacturer  : System manufacturer
    Name          : P5KC
    Part_Number   : To Be Filled By O.E.M.
    Serial_Number : System Serial Number
    Host_ID       : 00498e19

----------------------------------------
Suspect 1 of 1 :
   Problem class : fault.fs.zfs.vdev.probe_failure
   Certainty   : 100%
   Affects     : zfs://pool=18f02750861aa19d/vdev=d467907a19c2f97/pool_name=rpool/vdev_name=id1,cmdk@AST2000DL003-9VT166=____________5YD1WHHR/a
   Status      : faulted and taken out of service

   FRU
     Status           : faulty
     FMRI             : "zfs://pool=18f02750861aa19d/vdev=d467907a19c2f97/pool_name=rpool/vdev_name=id1,cmdk@AST2000DL003-9VT166=____________5YD1WHHR/a"

Description : Probe of ZFS device
              'id1,cmdk@AST2000DL003-9VT166=____________5YD1WHHR/a' in pool
              'rpool' has failed.

Response    : The device has been offlined and marked as faulted. An attempt
              will be made to activate a hot spare if available.

Impact      : Fault tolerance of the pool may be compromised.

Action      : Use 'fmadm faulty' to provide a more detailed view of this event.
              Run 'zpool status -lx' for more information. Please refer to the
              associated reference document at
              http://support.oracle.com/msg/ZFS-8000-NX for the latest service
              procedures and policies regarding this diagnosis.

--------------- ------------------------------------  -------------- ---------
TIME            EVENT-ID                              MSG-ID         SEVERITY
--------------- ------------------------------------  -------------- ---------
Mar 16 12:15:46 93f5ec66-6a42-485c-b99c-b12fd6bf6ce2  ZFS-8000-FD    Major

Problem Status    : isolated
Diag Engine       : zfs-diagnosis / 1.0
System
    Manufacturer  : unknown
    Name          : unknown
    Part_Number   : unknown
    Serial_Number : unknown

System Component
    Manufacturer  : System manufacturer
    Name          : P5KC
    Part_Number   : To Be Filled By O.E.M.
    Serial_Number : System Serial Number
    Host_ID       : 00498e19

----------------------------------------
Suspect 1 of 1 :
   Problem class : fault.fs.zfs.vdev.io
   Certainty   : 100%
   Affects     : zfs://pool=18f02750861aa19d/vdev=d467907a19c2f97/pool_name=rpool/vdev_name=id1,cmdk@AST2000DL003-9VT166=____________5YD1WHHR/a
   Status      : faulted and taken out of service

   FRU
     Status           : faulty
     FMRI             : "zfs://pool=18f02750861aa19d/vdev=d467907a19c2f97/pool_name=rpool/vdev_name=id1,cmdk@AST2000DL003-9VT166=____________5YD1WHHR/a"

Description : The number of I/O errors associated with ZFS device
              'id1,cmdk@AST2000DL003-9VT166=____________5YD1WHHR/a' in pool
              'rpool' exceeded acceptable levels.

Response    : The device has been offlined and marked as faulted. An attempt
              will be made to activate a hot spare if available.

Impact      : Fault tolerance of the pool may be compromised.

Action      : Use 'fmadm faulty' to provide a more detailed view of this event.
              Run 'zpool status -lx' for more information. Please refer to the
              associated reference document at
              http://support.oracle.com/msg/ZFS-8000-FD for the latest service
              procedures and policies regarding this diagnosis.

--------------- ------------------------------------  -------------- ---------
TIME            EVENT-ID                              MSG-ID         SEVERITY
--------------- ------------------------------------  -------------- ---------
Mar 16 04:50:28 aa77770a-66e1-4230-b11e-97e3dff2ba04  ZFS-8000-NX    Major

Problem Status    : isolated
Diag Engine       : zfs-diagnosis / 1.0
System
    Manufacturer  : unknown
    Name          : unknown
    Part_Number   : unknown
    Serial_Number : unknown

System Component
    Manufacturer  : System manufacturer
    Name          : P5KC
    Part_Number   : To Be Filled By O.E.M.
    Serial_Number : System Serial Number
    Host_ID       : 00498e19

----------------------------------------
Suspect 1 of 1 :
   Problem class : fault.fs.zfs.vdev.probe_failure
   Certainty   : 100%
   Affects     : zfs://pool=5ad7f5fe1816608d/vdev=11ffee6dd0349a2f/pool_name=space/vdev_name=id1,cmdk@AST2000DL003-9VT166=____________5YD1WHHR/d
   Status      : faulted and taken out of service

   FRU
     Status           : faulty
     FMRI             : "zfs://pool=5ad7f5fe1816608d/vdev=11ffee6dd0349a2f/pool_name=space/vdev_name=id1,cmdk@AST2000DL003-9VT166=____________5YD1WHHR/d"

Description : Probe of ZFS device
              'id1,cmdk@AST2000DL003-9VT166=____________5YD1WHHR/d' in pool
              'space' has failed.

Response    : The device has been offlined and marked as faulted. An attempt
              will be made to activate a hot spare if available.

Impact      : Fault tolerance of the pool may be compromised.

Action      : Use 'fmadm faulty' to provide a more detailed view of this event.
              Run 'zpool status -lx' for more information. Please refer to the
              associated reference document at
              http://support.oracle.com/msg/ZFS-8000-NX for the latest service
              procedures and policies regarding this diagnosis.

--------------- ------------------------------------  -------------- ---------
TIME            EVENT-ID                              MSG-ID         SEVERITY
--------------- ------------------------------------  -------------- ---------
Mar 17 14:51:35 d7ff82a3-d20a-47e6-86c9-dc35b3c2ab63  PCIEX-8000-J5  Major

Problem Status    : open
Diag Engine       : eft / 1.16
System
    Manufacturer  : unknown
    Name          : unknown
    Part_Number   : unknown
    Serial_Number : unknown

System Component
    Manufacturer  : System manufacturer
    Name          : P5KC
    Part_Number   : To Be Filled By O.E.M.
    Serial_Number : System Serial Number
    Host_ID       : 00498e19

----------------------------------------
Suspect 1 of 1 :
   Problem class : fault.io.pciex.device-interr-corr
   Certainty   : 100%
   Affects     : dev:////pci@0,0/pci8086,2944@1c,2/pci1095,3132@0
   Status      : faulted but still in service

   FRU
     Status           : faulty
     Location         : "/SYS/PCI_3"
     Manufacturer     : unknown
     Name             : unknown
     Part_Number      : unknown
     Revision         : unknown
     Serial_Number    : unknown
     Chassis
        Manufacturer  : Chassis Manufacture
        Name          : P5KC
        Part_Number   : Asset-1234567890
        Serial_Number : System Serial Number

Description : Too many recovered internal errors have been detected within the
              specified PCIEX device. This may degrade into a non-recoverable
              fault.

Response    : One or more device instances may be disabled

Impact      : Loss of services provided by the device instances associated with
              this fault

Action      : Use 'fmadm faulty' to provide a more detailed view of this event.
              Please refer to the associated reference document at
              http://support.oracle.com/msg/PCIEX-8000-J5 for the latest
              service procedures and policies regarding this diagnosis.

--------------- ------------------------------------  -------------- ---------
TIME            EVENT-ID                              MSG-ID         SEVERITY
--------------- ------------------------------------  -------------- ---------
Mar 14 13:39:26 ea62192a-c795-4805-9bd0-fa67fd1ce84f  ZFS-8000-GH    Major

Problem Status    : isolated
Diag Engine       : zfs-diagnosis / 1.0
System
    Manufacturer  : unknown
    Name          : unknown
    Part_Number   : unknown
    Serial_Number : unknown

System Component
    Manufacturer  : System manufacturer
    Name          : P5KC
    Part_Number   : To Be Filled By O.E.M.
    Serial_Number : System Serial Number
    Host_ID       : 00498e19

----------------------------------------
Suspect 1 of 1 :
   Problem class : fault.fs.zfs.vdev.checksum
   Certainty   : 100%
   Affects     : zfs://pool=5ad7f5fe1816608d/vdev=11ffee6dd0349a2f/pool_name=space/vdev_name=id1,cmdk@AST2000DL003-9VT166=____________5YD1WHHR/d
   Status      : faulted and taken out of service

   FRU
     Status           : faulty
     FMRI             : "zfs://pool=5ad7f5fe1816608d/vdev=11ffee6dd0349a2f/pool_name=space/vdev_name=id1,cmdk@AST2000DL003-9VT166=____________5YD1WHHR/d"

Description : The number of checksum errors associated with ZFS device
              'id1,cmdk@AST2000DL003-9VT166=____________5YD1WHHR/d' in pool
              'space' exceeded acceptable levels.

Response    : The device has been marked as degraded. An attempt will be made
              to activate a hot spare if available.

Impact      : Fault tolerance of the pool may be compromised.

Action      : Use 'fmadm faulty' to provide a more detailed view of this event.
              Run 'zpool status -lx' for more information. Please refer to the
              associated reference document at
              http://support.oracle.com/msg/ZFS-8000-GH for the latest service
              procedures and policies regarding this diagnosis.

--------------- ------------------------------------  -------------- ---------
TIME            EVENT-ID                              MSG-ID         SEVERITY
--------------- ------------------------------------  -------------- ---------
Mar 16 04:50:37 f11e9ea3-6956-4051-994a-f9e78fd66610  ZFS-8000-FD    Major

Problem Status    : isolated
Diag Engine       : zfs-diagnosis / 1.0
System
    Manufacturer  : unknown
    Name          : unknown
    Part_Number   : unknown
    Serial_Number : unknown

System Component
    Manufacturer  : System manufacturer
    Name          : P5KC
    Part_Number   : To Be Filled By O.E.M.
    Serial_Number : System Serial Number
    Host_ID       : 00498e19

----------------------------------------
Suspect 1 of 1 :
   Problem class : fault.fs.zfs.vdev.io
   Certainty   : 100%
   Affects     : zfs://pool=5ad7f5fe1816608d/vdev=11ffee6dd0349a2f/pool_name=space/vdev_name=id1,cmdk@AST2000DL003-9VT166=____________5YD1WHHR/d
   Status      : faulted and taken out of service

   FRU
     Status           : faulty
     FMRI             : "zfs://pool=5ad7f5fe1816608d/vdev=11ffee6dd0349a2f/pool_name=space/vdev_name=id1,cmdk@AST2000DL003-9VT166=____________5YD1WHHR/d"

Description : The number of I/O errors associated with ZFS device
              'id1,cmdk@AST2000DL003-9VT166=____________5YD1WHHR/d' in pool
              'space' exceeded acceptable levels.

Response    : The device has been offlined and marked as faulted. An attempt
              will be made to activate a hot spare if available.

Impact      : Fault tolerance of the pool may be compromised.

Action      : Use 'fmadm faulty' to provide a more detailed view of this event.
              Run 'zpool status -lx' for more information. Please refer to the
              associated reference document at
              http://support.oracle.com/msg/ZFS-8000-FD for the latest service
              procedures and policies regarding this diagnosis.`
