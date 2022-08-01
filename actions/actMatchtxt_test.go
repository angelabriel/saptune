package actions

import (
	"fmt"
	"github.com/SUSE/saptune/system"
)

var noteListMatchText = `
All notes (+ denotes manually enabled notes, * denotes notes enabled by solutions, - denotes notes enabled by solutions but reverted manually later, O denotes override file exists for note, C denotes custom note):
	900929		Linux: STORAGE_PARAMETERS_WRONG_SET and 'mmap() failed'
			Version 7 from 31.07.2017
			https://launchpad.support.sap.com/#/notes/900929
	NEWSOL2NOTE	
	extraNote	Configuration drop in for extra tests
			Version 0 from 04.06.2019
	oldFile		Name_syntax
	simpleNote	Configuration drop in for simple tests
			Version 1 from 09.07.2019
	wrongFileNamesyntax	

Remember: if you wish to automatically activate the solution's tuning options after a reboot, you must enable and start saptune.service by running:
    saptune service enablestart
`

var solutionListMatchText = `
All solutions (* denotes enabled solution, O denotes override file exists for solution, C denotes custom solutions, D denotes deprecated solutions):
	BWA                - SAP_BWA
	HANA               - 941735 1771258 1980196 1984787 2205917 2382421 2534844
	MAXDB              - 941735 1771258 1984787
	NETW               - 941735 1771258 1980196 1984787 2534844

Remember: if you wish to automatically activate the solution's tuning options after a reboot, you must enable and start saptune.service by running:
    saptune service enablestart
`

var saptuneStatusMatchText = fmt.Sprintf(`
saptune.service:          disabled/active
saptune package:          'undef'
configured version:       '3'
enabled Solution:         sol1 (simpleNote)
applied Solution:         
additional enabled Notes: 2205917 
enabled Notes:            2205917
applied Notes:            
staging:                  disabled
staged Notes:             
staged Solutions:         

sapconf.service:          not available
tuned.service:            disabled/active (profile: '%s')
system state:             running
virtualization:           %s

Remember: if you wish to automatically activate the note's and solution's tuning options after a reboot, you must enable saptune.service by running:
 'saptune service enable'.

`, system.GetTunedAdmProfile(), system.GetVirtStatus())

var saptuneStatMatchText = fmt.Sprintf(`
saptune.service:          disabled/inactive
saptune package:          'undef'
configured version:       '3'
enabled Solution:         
applied Solution:         
additional enabled Notes: 
enabled Notes:            
applied Notes:            
staging:                  disabled
staged Notes:             
staged Solutions:         

sapconf.service:          not available
tuned.service:            disabled/active (profile: '%s')
system state:             running
virtualization:           %s

Remember: if you wish to automatically activate the note's and solution's tuning options after a reboot, you must enable saptune.service by running:
 'saptune service enablestart'.
Your system has not yet been tuned. Please visit ` + "`" + `saptune note` + "`" + ` and ` + "`" + `saptune solution` + "`" + ` to start tuning.

`, system.GetTunedAdmProfile(), system.GetVirtStatus())

/*
var solutionSimulateMatchText = `If you run ` + "`saptune solution apply sol1`" + `, the following changes will be applied to your system:

   Parameter                    | Value set   | Value expected  | Override  | Comment
--------------------------------+-------------+-----------------+-----------+--------------
   net.ipv4.ip_local_port_range | 32768 60999 | 31768 61999     |           |   


[31mAttention for SAP Note simpleNote:
Hints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:
# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.
[0m
`
var solutionApplyMatchText = `All tuning options for the SAP solution have been applied successfully.

Remember: if you wish to automatically activate the solution's tuning options after a reboot, you must enable and start saptune.service by running:
    saptune service enablestart
`
		var verifyMatchText = `
   SAPNote, Version | Parameter                    | Expected    | Override  | Actual      | Compliant
--------------------+------------------------------+-------------+-----------+-------------+-----------
   simpleNote, 1    | net.ipv4.ip_local_port_range | 31768 61999 |           | 31768 61999 | yes


[31mAttention for SAP Note simpleNote:
Hints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:
# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.
[0m
[32m[1mThe system fully conforms to the tuning guidelines of the specified SAP solution.[22m[0m
`
var solutionEnabledMatchText = "sol1"
var solutionRevertMatchText = `Parameters tuned by the notes referred by the SAP solution have been successfully reverted.
`
var solutionShowMatchText = `
Content of Solution NEWSOL1:
[version]
# SAP-NOTE=NEWSOL1 CATEGORY=SOLUTION VERSION=1 DATE=07.07.2021 NAME="Definition of NEWSOL1 solution for test"

[ArchX86]
SOL1NOTE1 NEWSOL1NOTE SOL1NOTE2

[ArchPPC64LE]
SOL1NOTE1 NEWSOL1NOTE SOL1NOTE2

`
*/





var PrintHelpAndExitMatchText = `saptune: Comprehensive system optimisation management for SAP solutions.
Daemon control:
  saptune daemon [ start | status | stop ]  ATTENTION: deprecated
  saptune service [ start | status | stop | restart | takeover | enable | disable | enablestart | disablestop ]
Tune system according to SAP and SUSE notes:
  saptune note [ list | revertall | enabled | applied ]
  saptune note [ apply | simulate | customise | create | edit | revert | show | delete ] NoteID
  saptune note verify [--colorscheme=<color scheme>] [--show-non-compliant] [NoteID]
  saptune note rename NoteID newNoteID
Tune system for all notes applicable to your SAP solution:
  saptune solution [ list | verify | enabled | applied ]
  saptune solution [ apply | simulate | verify | customise | create | edit | revert | show | delete ] SolutionName
  saptune solution rename SolutionName newSolutionName
Staging control:
   saptune staging [ status | enable | disable | is-enabled | list | diff | analysis | release ]
   saptune staging [ analysis | diff ] [ NoteID... | SolutionID... | all ]
   saptune staging release [--force|--dry-run] [ NoteID... | SolutionID... | all ]
Revert all parameters tuned by the SAP notes or solutions:
  saptune revert all
Remove the pending lock file from a former saptune call
  saptune lock remove
Call external script '/usr/sbin/saptune_check'
  saptune check
Print current saptune status:
  saptune status
Print current saptune version:
  saptune version
Print this message:
  saptune help
`
