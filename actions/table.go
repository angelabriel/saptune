package actions

import (
	"fmt"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"io"
	"sort"
	"strconv"
	"strings"
)

// define smiley for 'yes' and 'no' in 'verify'
var smileYes = '\U0001F60E'
var smileNo = '\U0001F61F'
var smileDash = '\U0001F611'

// define max column width
var fmtmax = 30

// number of lines printed for 'verify'
var lineCnt = 0

// PrintNoteFields Print mismatching fields in the note comparison result.
func PrintNoteFields(writer io.Writer, header string, noteComparisons map[string]map[string]note.FieldComparison, printComparison bool, result *system.JPNotes) {
	// initialise
	colFormat := ""
	colCompliant := ""
	printHead := ""
	noteField := ""
	reminder := make(map[string]string)
	override := ""
	pExp := ""
	noteLine := system.JPNotesLine{}
	noteList := []system.JPNotesLine{}

	var compliant string
	var comment string
	var footnote []string = make([]string, 16)

	colorScheme := getColorScheme()
	// sort output
	sortkeys := sortNoteComparisonsOutput(noteComparisons)

	// setup table format values
	fmtlen0, fmtlen1, fmtlen2, fmtlen3, fmtlen4, format := setupTableFormat(sortkeys, noteComparisons, printComparison)

	// print
	noteID := ""
	for _, skey := range sortkeys {
		comment = ""
		keyFields := strings.Split(skey, "§")
		key := keyFields[1]
		printHead, noteID, noteField = getNoteAndVersion(keyFields[0], noteID, noteField, noteComparisons)
		override = strings.Replace(noteComparisons[noteID][fmt.Sprintf("%s[%s]", "OverrideParams", key)].ExpectedValueJS, "\t", " ", -1)
		comparison := noteComparisons[noteID][fmt.Sprintf("%s[%s]", "SysctlParams", key)]
		if comparison.ReflectMapKey == "reminder" {
			reminder[noteID] = reminder[noteID] + comparison.ExpectedValueJS
			continue
		}
		// set compliant information according to the comparison result
		compliant = setCompliant(comparison)

		// check inform map for special settings
		inform := getInformSettings(noteID, noteComparisons, comparison)

		// prepare footnote
		compliant, comment, footnote = prepareFootnote(comparison, compliant, comment, inform, footnote)

		// print table header
		if printHead != "" {
			printHeadline(writer, header, noteID, noteComparisons)
			printTableHeader(writer, format, fmtlen0, fmtlen1, fmtlen2, fmtlen3, fmtlen4, printComparison)
		}

		// print table body
		// define variable pAct here and not at the beginning as we need
		// the 'address' of the variable some lines below
		pAct := strings.Replace(comparison.ActualValueJS, "\t", " ", -1)
		if comparison.ActualValueJS == "PNA" {
			pAct = "NA"
		}
		noteLine.ActValue = &pAct
		pExp = strings.Replace(comparison.ExpectedValueJS, "\t", " ", -1)
		tableColumns := make(map[string]string)
		if printComparison {
			// verify
			if system.IsFlagSet("show-non-compliant") && (strings.Contains(compliant, "yes") || strings.Contains(compliant, "-")) {
				// print only non-compliant rows, so skip the others
				continue
			}
			colFormat, colCompliant = colorPrint(format, compliant, colorScheme)
			tableColumns = map[string]string{"type": "verify", "colFormat": colFormat, "note": noteField, "parameter": comparison.ReflectMapKey, "expected": pExp, "override": override, "actual": pAct, "compliant": colCompliant}
		} else {
			// simulate
			tableColumns = map[string]string{"type": "simulate", "colFormat": format, "parameter": comparison.ReflectMapKey, "actual": pAct, "expected": pExp, "override": override, "comment": comment}
		}
		printTableRow(writer, tableColumns)
		noteLine = collectMRO(noteLine, compliant, noteID, noteComparisons, comparison, pExp, override, printComparison, comment, footnote, pAct)
		noteList = append(noteList, noteLine)
	}

	// print footer
	reminderList := []system.JPNotesRemind{}
	printTableFooter(writer, header, footnote, reminder, &reminderList)
	if result != nil {
		if printComparison {
			// verify
			result.Verifications = noteList
			result.Simulations = []system.JPNotesLine{}
		} else {
			// simulate
			result.Verifications = []system.JPNotesLine{}
			result.Simulations = noteList
		}
		result.Attentions = reminderList
	}
}

// collectMRO collects the data for machine readable output
// given parameter - Attention - order of parameter is important!
// noteLine, compliant, noteID, noteComparisons, comparison, pExp, override,
// printComparison, comment, footnote, pAct
func collectMRO(stuff ...interface{}) system.JPNotesLine {
	nLine := stuff[0].(system.JPNotesLine)
	noteComp := !strings.Contains(stuff[1].(string), "no")
	nLine.NoteID = stuff[2].(string)
	nLine.NoteVers = txtparser.GetINIFileVersionSectionEntry(stuff[3].(map[string]map[string]note.FieldComparison)[stuff[2].(string)]["ConfFilePath"].ActualValue.(string), "version")
	nLine.Parameter = stuff[4].(note.FieldComparison).ReflectMapKey
	nLine.ExpValue = stuff[5].(string)
	nLine.OverValue = stuff[6].(string)
	nLine.Compliant = &noteComp

	if strings.Contains(stuff[1].(string), "-") {
		nLine.Compliant = nil
	}
	if stuff[7].(bool) {
		// verify
		nLine.Comment = ""
	} else {
		// simulate
		nLine.NoteID = ""
		nLine.NoteVers = ""
		nLine.Comment = stuff[8].(string)
		nLine.Compliant = nil
	}
	noteFNs := []system.JFootNotes{}
	fns := system.JFootNotes{}
	for _, fn := range strings.Fields(stuff[8].(string)) {
		indx := fn[1 : len(fn)-1]
		idx, _ := strconv.Atoi(indx)
		if idx > 0 {
			fns.FNoteNumber = idx
			fns.FNoteTxt = stuff[9].([]string)[idx-1]
		}
		noteFNs = append(noteFNs, fns)
	}
	nLine.Footnotes = noteFNs

	if stuff[10].(string) == "NA" || stuff[10].(string) == "PNA" || stuff[10].(string) == "all:none" {
		nLine.ActValue = nil
	}
	return nLine
}

// sortNoteComparisonsOutput sorts the output of the Note comparison
// the reminder section should be the last one
func sortNoteComparisonsOutput(noteCompare map[string]map[string]note.FieldComparison) []string {
	skeys := make([]string, 0, len(noteCompare))
	rkeys := make([]string, 0, len(noteCompare))
	// sort output
	for noteID, comparisons := range noteCompare {
		for _, comparison := range comparisons {
			if comparison.ReflectFieldName == "Inform" {
				// skip inform map to avoid double entries in verify table
				continue
			}
			if len(comparison.ReflectMapKey) != 0 && comparison.ReflectFieldName != "OverrideParams" {
				if comparison.ReflectMapKey != "reminder" {
					skeys = append(skeys, noteID+"§"+comparison.ReflectMapKey)
				} else {
					rkeys = append(rkeys, noteID+"§"+comparison.ReflectMapKey)
				}
			}
		}
	}
	sort.Strings(skeys)
	skeys = append(skeys, rkeys...)
	return skeys
}

// setupTableFormat sets the format of the table columns dependent on the content
func setupTableFormat(skeys []string, noteCompare map[string]map[string]note.FieldComparison, printComp bool) (int, int, int, int, int, string) {
	var fmtlen0, fmtlen1, fmtlen2, fmtlen3, fmtlen4 int
	format := "\t%s : %s\n"
	// define start values for the column width
	if printComp {
		// verify
		fmtlen0 = 16
		fmtlen1 = 12
		fmtlen2 = 9
		fmtlen3 = 9
		fmtlen4 = 7
	} else {
		// simulate
		fmtlen1 = 12
		fmtlen2 = 10
		fmtlen3 = 15
		fmtlen4 = 9
	}

	for _, skey := range skeys {
		keyFields := strings.Split(skey, "§")
		noteID := keyFields[0]
		noteField := fmt.Sprintf("%s, %s", noteID, txtparser.GetINIFileVersionSectionEntry(noteCompare[noteID]["ConfFilePath"].ActualValue.(string), "version"))
		comparisons := noteCompare[noteID]
		for _, comparison := range comparisons {
			if comparison.ReflectMapKey == "reminder" || comparison.ReflectFieldName == "Inform" {
				continue
			}
			if printComp {
				// verify
				if len(noteField) > fmtlen0 {
					fmtlen0 = len(noteField)
				}
				// 3:override, 1:mapkey, 2:expval, 4:actval
				fmtlen3, fmtlen1, fmtlen2, fmtlen4 = setWidthOfColums(comparison, fmtlen3, fmtlen1, fmtlen2, fmtlen4)
				format = "   %-" + strconv.Itoa(fmtlen0) + "s | %-" + strconv.Itoa(fmtlen1) + "s | %-" + strconv.Itoa(fmtlen2) + "s | %-" + strconv.Itoa(fmtlen3) + "s | %-" + strconv.Itoa(fmtlen4) + "s | %2s\n"
			} else {
				// simulate
				// 4:override, 1:mapkey, 3:expval, 2:actval
				fmtlen4, fmtlen1, fmtlen3, fmtlen2 = setWidthOfColums(comparison, fmtlen4, fmtlen1, fmtlen3, fmtlen2)
				format = "   %-" + strconv.Itoa(fmtlen1) + "s | %-" + strconv.Itoa(fmtlen2) + "s | %-" + strconv.Itoa(fmtlen3) + "s | %-" + strconv.Itoa(fmtlen4) + "s | %2s\n"
			}
		}
	}
	return fmtlen0, fmtlen1, fmtlen2, fmtlen3, fmtlen4, format
}

// printHeadline prints a headline for the table
func printHeadline(writer io.Writer, header, id string, noteComparisons map[string]map[string]note.FieldComparison) {
	if header != "NONE" {
		nName := noteComparisons[id]["DescriptiveName"].ActualValue.(string)
		fmt.Fprintf(writer, "\n%s - %s \n\n", id, nName)
	} else {
		fmt.Fprintf(writer, "\n")
	}
}

// printTableHeader prints the header of the table
func printTableHeader(writer io.Writer, format string, col0, col1, col2, col3, col4 int, printComp bool) {
	if printComp {
		// verify
		fmt.Fprintf(writer, format, "SAPNote, Version", "Parameter", "Expected", "Override", "Actual", "Compliant")
		for i := 0; i < col0+col1+col2+col3+col4+28; i++ {
			if i == 3+col0+1 || i == 3+col0+3+col1+1 || i == 3+col0+3+col1+4+col2 || i == 3+col0+3+col1+4+col2+2+col3+1 || i == 3+col0+3+col1+4+col2+2+col3+3+col4+1 {
				fmt.Fprintf(writer, "+")
			} else {
				fmt.Fprintf(writer, "-")
			}
		}
		fmt.Fprintf(writer, "\n")
	} else {
		// simulate
		fmt.Fprintf(writer, format, "Parameter", "Value set", "Value expected", "Override", "Comment")
		for i := 0; i < col1+col2+col3+col4+28; i++ {
			if i == 3+col1+1 || i == 3+col1+3+col2+1 || i == 3+col1+3+col2+3+col3+1 || i == 3+col1+3+col2+3+col3+3+col4+1 {
				fmt.Fprintf(writer, "+")
			} else {
				fmt.Fprintf(writer, "-")
			}
		}
		fmt.Fprintf(writer, "\n")
	}
}

// printTableFooter prints the footer of the table
// footnotes and reminder section
func printTableFooter(writer io.Writer, header string, footnote []string, reminder map[string]string, noteReminder *[]system.JPNotesRemind) {
	for _, fn := range footnote {
		if fn != "" && lineCnt > 0 {
			fmt.Fprintf(writer, "\n %s", fn)
		}
	}
	fmt.Fprintf(writer, "\n\n")
	noteRem := system.JPNotesRemind{}
	for noteID, reminde := range reminder {
		if reminde != "" {
			reminderHead := fmt.Sprintf("Attention for SAP Note %s:\nHints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:\n", noteID)
			fmt.Fprintf(writer, "%s\n", setRedText+reminderHead+reminde+resetTextColor)
			noteRem.NoteID = noteID
			noteRem.NoteReminder = reminderHead + reminde
		}
		*noteReminder = append(*noteReminder, noteRem)
	}
}

// getNoteAndVersion sets printHead, noteID, noteField for the next table row
func getNoteAndVersion(kField, nID, nField string, nComparisons map[string]map[string]note.FieldComparison) (string, string, string) {
	pHead := ""
	if kField != nID {
		if nID == "" {
			pHead = "yes"
		}
		nID = kField
		nField = fmt.Sprintf("%s, %s", nID, txtparser.GetINIFileVersionSectionEntry(nComparisons[nID]["ConfFilePath"].ActualValue.(string), "version"))
	}
	return pHead, nID, nField
}

// setCompliant sets compliant information according to the comparison result
func setCompliant(comparison note.FieldComparison) string {
	comp := ""
	if !comparison.MatchExpectation {
		comp = "no "
	} else {
		comp = "yes"
	}
	if comparison.ActualValue.(string) == "all:none" {
		comp = " - "
	}
	return comp
}

// getInformSettings checks inform map for special settings
func getInformSettings(nID string, nComparisons map[string]map[string]note.FieldComparison, comparison note.FieldComparison) string {
	inf := ""
	if nComparisons[nID][fmt.Sprintf("%s[%s]", "Inform", comparison.ReflectMapKey)].ActualValue != nil {
		inf = nComparisons[nID][fmt.Sprintf("%s[%s]", "Inform", comparison.ReflectMapKey)].ActualValue.(string)
		if inf == "" && nComparisons[nID][fmt.Sprintf("%s[%s]", "Inform", comparison.ReflectMapKey)].ExpectedValue != nil {
			inf = nComparisons[nID][fmt.Sprintf("%s[%s]", "Inform", comparison.ReflectMapKey)].ExpectedValue.(string)
		}
	}
	return inf
}

// setWidthOfColums sets the width of the columns for verify and simulate
// depending on the highest number of characters of the content to be
// displayed
// c1:override, c2:mapkey, c3:expval, c4:actval
func setWidthOfColums(compare note.FieldComparison, c1, c2, c3, c4 int) (int, int, int, int) {
	if len(compare.ReflectMapKey) != 0 {
		if compare.ReflectFieldName == "OverrideParams" && len(compare.ActualValueJS) > c1 {
			// in case of override content of ActualValueJS and
			// ExpectedValueJS is the same, so one length check
			// is sufficient
			c1 = len(compare.ActualValueJS)
			c1, c3, c4 = chkMaxWidthOfColums([]int{c1, c3, c4})
			return c1, c2, c3, c4
		}
		if len(compare.ReflectMapKey) > c2 {
			c2 = len(compare.ReflectMapKey)
		}
		if len(compare.ExpectedValueJS) > c3 {
			c3 = len(compare.ExpectedValueJS)
		}
		if len(compare.ActualValueJS) > c4 {
			c4 = len(compare.ActualValueJS)
		}
	}
	c1, c3, c4 = chkMaxWidthOfColums([]int{c1, c3, c4})
	return c1, c2, c3, c4
}

// chkMaxWidthOfColums limits the width of the columns for verify and simulate
// to a defined value. Content of columns need to be wrapped
func chkMaxWidthOfColums(fmts []int) (int, int, int) {
	for w, width := range fmts {
		if width > fmtmax {
			fmts[w] = fmtmax
		}
	}
	return fmts[0], fmts[1], fmts[2]
}

// getColorScheme reads the color scheme from CLI flag or from saptune
// sysconfig file or sets default
func getColorScheme() string {
	if system.IsFlagSet("show-non-compliant") {
		return "black"
	}
	// check, if CLI flag is available
	scheme := system.GetFlagVal("colorscheme")
	if scheme != "" {
		system.InfoLog("color scheme defined by command line flag - %s", scheme)
	} else {
		// no flag, check sysconfig file
		sconf, err := txtparser.ParseSysconfigFile(saptuneSysconfig, false)
		if err == nil {
			scheme = sconf.GetString("COLOR_SCHEME", "")
			if scheme != "" {
				system.InfoLog("color scheme defined in sysconfig file - %s", scheme)
			}
		}
	}
	if scheme == "" {
		// no flag, no sysconfig file entry - use default
		scheme = dfltColorScheme
	}
	return scheme
}

// colorPrint sets the color scheme for the 'verify' table print
// following color schemes are supported:
// full-zebra, cmpl-zebra, full-noncmpl, noncmpl
func colorPrint(format, compliant, colorScheme string) (string, string) {
	colFormat := ""
	colCompl := ""
	switch colorScheme {
	case "full-green-zebra":
		// full-green-zebra - whole line is colored green (compliant) or
		// red (not compliant)
		colFormat = colorFormating("green", "red", format, compliant)
	case "cmpl-green-zebra":
		// cmpl-green-zebra - only the content in the Compliant column
		// is colored green (compliant) or red (not compliant)
		colCompl = colorFormating("green", "red", compliant, compliant)
	case "full-blue-zebra":
		// full-blue-zebra - whole line is colored blue (compliant) or
		// yellow (not compliant)
		colFormat = colorFormating("blue", "yellow", format, compliant)
	case "cmpl-blue-zebra":
		// cmpl-blue-zebra - only the content in the Compliant column is
		// colored blue (compliant) or yellow (not compliant)
		colCompl = colorFormating("blue", "yellow", compliant, compliant)
	case "full-red-noncmpl":
		// full-red-noncmpl - only the whole line of the not compliant
		// parameter is colored red
		colFormat = colorFormating("", "red", format, compliant)
	case "red-noncmpl":
		// red-noncmpl - only the content in the Compliant column of the
		// not compliant parameter is colored red
		colCompl = colorFormating("", "red", compliant, compliant)
	case "full-yellow-noncmpl":
		// full-yellow-noncmpl - only the whole line of the not
		// compliant parameter is colored yellow
		colFormat = colorFormating("", "yellow", format, compliant)
	case "yellow-noncmpl":
		// yellow-noncmpl - only the content in the Compliant column of
		// the not compliant parameter is colored yellow
		colCompl = colorFormating("", "yellow", compliant, compliant)
	default:
		system.InfoLog("unknown color scheme definition - %s", colorScheme)
	}
	if colFormat == "" {
		colFormat = format
	}
	if colCompl == "" {
		colCompl = compliant
	}
	if system.IsFlagSet("fun") {
		colCompl = funPrint(colCompl)
	}
	return colFormat, colCompl
}

// colorFormating sets the colors for compliant and non compliant lines in
// 'saptune verify' output
func colorFormating(colCmpl, colNonCmpl, txt, compliant string) string {
	colFormat := ""
	if strings.Contains(compliant, "yes") {
		if colCmpl == "green" {
			colFormat = setGreenText + txt + resetTextColor
		}
		if colCmpl == "blue" {
			colFormat = setBlueText + txt + resetTextColor
		}
	}
	if strings.Contains(compliant, "no") {
		if colNonCmpl == "red" {
			colFormat = setRedText + txt + resetTextColor
		}
		if colNonCmpl == "yellow" {
			colFormat = setYellowText + txt + resetTextColor
		}
	}
	return colFormat
}

// funPrint prints emojis instead of text in the 'compliant' column of 'verify'
func funPrint(txt string) string {
	if strings.Contains(txt, "yes") {
		txt = strings.Replace(txt, "yes", " "+string(smileYes), 1)
	}
	if strings.Contains(txt, "no") {
		txt = strings.Replace(txt, "no ", " "+string(smileNo), 1)
	}
	return txt
}

// printTableRow prints one row of the table
// If needed the lines of the override column, the expected column and the
// actual column will be wrapped after 'fmtmax' characters
// if override exists, expected == override, so compare of width of expected and
// actual column is sufficient
func printTableRow(writer io.Writer, rowElements map[string]string) {
	lineCnt = lineCnt + 1
	wrappedActual := system.WrapTxt(rowElements["actual"], fmtmax)
	wrappedExpected := system.WrapTxt(rowElements["expected"], fmtmax)
	wrappedOverride := system.WrapTxt(rowElements["override"], fmtmax)
	linesAct := len(wrappedActual)
	linesExp := len(wrappedExpected)
	linesOver := len(wrappedOverride)
	if linesAct == 1 && linesExp == 1 && linesOver == 1 {
		if rowElements["type"] == "verify" {
			fmt.Fprintf(writer, rowElements["colFormat"], rowElements["note"], rowElements["parameter"], rowElements["expected"], rowElements["override"], rowElements["actual"], rowElements["compliant"])
		} else {
			// simulate
			fmt.Fprintf(writer, rowElements["colFormat"], rowElements["parameter"], rowElements["actual"], rowElements["expected"], rowElements["override"], rowElements["comment"])
		}
		return
	}
	wrappedElements := map[string][]string{"wrappedActual": wrappedActual, "wrappedExpected": wrappedExpected, "wrappedOverride": wrappedOverride}
	printWrappedRow(writer, wrappedElements, rowElements)
}

// printWrappedRow prints the wrapped columns of one verify output row
// twist - true - compare order ACT, EXP
// twist - false - compare order EXP, ACT
// if override exists, expected == override
func printWrappedRow(writer io.Writer, wrappedElem map[string][]string, rowElements map[string]string) {
	var wrappedA, wrappedB, wrappedC []string
	firstLine := true
	twist := false

	if len(wrappedElem["wrappedActual"]) >= len(wrappedElem["wrappedExpected"]) {
		twist = true
	}
	if twist {
		wrappedA = wrappedElem["wrappedActual"]
		wrappedB = wrappedElem["wrappedExpected"]
	} else {
		wrappedA = wrappedElem["wrappedExpected"]
		wrappedB = wrappedElem["wrappedActual"]
	}
	wrappedC = wrappedElem["wrappedOverride"]
	noLinesB := len(wrappedB)
	noLinesC := len(wrappedC)
	colB := ""
	colC := ""

	for c, colA := range wrappedA {
		// ANGI todo <=
		if c < noLinesB {
			colB = wrappedB[c]
		} else {
			colB = ""
		}
		if c < noLinesC {
			colC = wrappedC[c]
		} else {
			colC = ""
		}
		printRow(writer, twist, []string{colA, colB, colC}, rowElements)
		if firstLine {
			firstLine = false
			rowElements["note"] = ""
			rowElements["parameter"] = ""
			rowElements["compliant"] = ""
		}
	}
}

// printRow prints now the row of the table
func printRow(writer io.Writer, twist bool, cols []string, rowElements map[string]string) {
	if rowElements["type"] == "verify" {
		printVerifyRow(twist, writer, rowElements, cols)
	} else {
		printSimulateRow(twist, writer, rowElements, cols)
	}
}

// printSimulateRow prints the 'simulate' row
func printSimulateRow(twist bool, writer io.Writer, rowElements map[string]string, cols []string) {
	if twist {
		fmt.Fprintf(writer, rowElements["colFormat"], rowElements["parameter"], cols[0], cols[1], cols[2], rowElements["comment"])
	} else {
		fmt.Fprintf(writer, rowElements["colFormat"], rowElements["parameter"], cols[1], cols[0], cols[2], rowElements["comment"])
	}
}

// printVerifyRow prints the 'verify' row
func printVerifyRow(twist bool, writer io.Writer, rowElements map[string]string, cols []string) {
	if twist {
		fmt.Fprintf(writer, rowElements["colFormat"], rowElements["note"], rowElements["parameter"], cols[1], cols[2], cols[0], rowElements["compliant"])
	} else {
		fmt.Fprintf(writer, rowElements["colFormat"], rowElements["note"], rowElements["parameter"], cols[0], cols[2], cols[1], rowElements["compliant"])
	}
}
