package txtparser

import (
	"fmt"
	"github.com/SUSE/saptune/system"
	"io/ioutil"
	"regexp"
	"strings"
)

// Operator definitions
const (
	OperatorLessThan      = "<"
	OperatorLessThanEqual = "<="
	OperatorMoreThan      = ">"
	OperatorMoreThanEqual = ">="
	OperatorEqual         = "="
)

// Operator is the comparison or assignment operator used in an INI file entry
type Operator string

// RegexKeyOperatorValue breaks up a line into key, operator, value.
var RegexKeyOperatorValue = regexp.MustCompile(`([\w.+_-]+)\s*([<=>]+)\s*["']*(.*?)["']*$`)

// counter to control the [block] section detected warning
var blckCnt = 0

// INIEntry contains a single key-value pair in INI file.
type INIEntry struct {
	Section  string
	Key      string
	Operator Operator
	Value    string
}

// INIFile contains all key-value pairs of an INI file.
type INIFile struct {
	AllValues []INIEntry
	KeyValue  map[string]map[string]INIEntry
}

// GetINIFileDescriptiveName return the descriptive name of the Note
func GetINIFileDescriptiveName(fileName string) string {
	var re = regexp.MustCompile(`# .*NOTE=.*VERSION=(\d*)\s*DATE=(.*)\s*NAME="([^"]*)"`)
	rval := ""
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return ""
	}
	matches := re.FindStringSubmatch(string(content))
	if len(matches) != 0 {
		rval = fmt.Sprintf("%s\n\t\t\t%sVersion %s from %s", matches[3], "", matches[1], matches[2])
	}
	return rval
}

// GetINIFileVersionSectionEntry returns the field 'entryName' from the version
// section of the Note configuration file
func GetINIFileVersionSectionEntry(fileName, entryName string) string {
	var re = regexp.MustCompile(`# .*NOTE=.*TEST=(\d*)\s*DATE=.*"`)
	switch entryName {
	case "version":
		re = regexp.MustCompile(`# .*NOTE=.*VERSION=(\d*)\s*DATE=.*"`)
	case "category":
		re = regexp.MustCompile(`# .*NOTE=.*CATEGORY=(\w*)\s*VERSION=.*"`)
	default:
		return ""
	}
	rval := ""
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return ""
	}
	matches := re.FindStringSubmatch(string(content))
	if len(matches) != 0 {
		rval = fmt.Sprintf("%s", matches[1])
	}
	return rval
}

// ParseINIFile read the content of the configuration file
func ParseINIFile(fileName string, autoCreate bool) (*INIFile, error) {
	content, err := system.ReadConfigFile(fileName, autoCreate)
	if err != nil {
		return nil, err
	}
	return ParseINI(string(content)), nil
}

// ParseINI parse the content of the configuration file
func ParseINI(input string) *INIFile {
	ret := &INIFile{
		AllValues: make([]INIEntry, 0, 64),
		KeyValue:  make(map[string]map[string]INIEntry),
	}

	reminder := ""
	currentSection := ""
	currentEntriesArray := make([]INIEntry, 0, 8)
	currentEntriesMap := make(map[string]INIEntry)
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if line[0] == '[' {
			// Save previous section
			if currentSection != "" {
				ret.KeyValue[currentSection] = currentEntriesMap
				ret.AllValues = append(ret.AllValues, currentEntriesArray...)
			}
			// Start a new section
			currentSection = line[1 : len(line)-1]
			currentEntriesArray = make([]INIEntry, 0, 8)
			currentEntriesMap = make(map[string]INIEntry)
			continue
		}
		if strings.HasPrefix(line, "#") {
			// Skip comments. Need to be done before
			// 'break apart the line into key, operator, value'
			// to support comments like # something (default = 60)
			// without side effects
			if currentSection == "reminder" {
				reminder = reminder + line + "\n"
			}
			continue
		}
		// Break apart a line into key, operator, value.
		kov := make([]string, 0)
		if currentSection == "rpm" {
			fields := strings.Fields(line)
			if fields[1] == "all" || fields[1] == system.GetOsVers() {
				kov = []string{"rpm", "rpm:" + fields[0], fields[1], fields[2]}
			} else {
				kov = nil
			}
		} else {
			kov = RegexKeyOperatorValue.FindStringSubmatch(line)
			if currentSection == "grub" {
				if len(kov) == 0 {
					// seams to be a single option and not
					// a key=value pair
					kov = []string{line, "grub:" + line, "=", line}
				} else {
					kov[1] = "grub:" + kov[1]
				}
			}
		}
		if kov == nil {
			// Skip comments, empty, and irregular lines.
			continue
		}
		if currentSection == "limits" {
			for _, limits := range strings.Split(kov[3], ",") {
				limits = strings.TrimSpace(limits)
				lim := strings.Fields(limits)
				key := ""
				if len(lim) == 0 {
					// empty LIMITS parameter means
					// override file is setting all limits to 'untouched'
					// or a wrong limits entry in an 'extra' file
					key = fmt.Sprintf("%s_NA", kov[1])
					limits = "NA"
				} else {
					key = fmt.Sprintf("LIMIT_%s_%s_%s", lim[0], lim[1], lim[2])
				}
				entry := INIEntry{
					Section:  currentSection,
					Key:      key,
					Operator: Operator(kov[2]),
					Value:    limits,
				}
				currentEntriesArray = append(currentEntriesArray, entry)
				currentEntriesMap[entry.Key] = entry
			}
		} else if currentSection == "block" {
			if blckCnt == 0 {
				system.WarningLog("[block] section detected: Traversing all block devices can take a considerable amount of time.")
				blckCnt = blckCnt + 1
			}
			// identify virtio block devices
			isVD := regexp.MustCompile(`^vd\w+$`)
			_, sysDevs := system.ListDir("/sys/block", "the available block devices of the system")
			for _, bdev := range sysDevs {
				// /sys/block/*/device/type (TYPE_DISK / 0x00)
				// does not work for virtio block devices
				fname := fmt.Sprintf("/sys/block/%s/device/type", bdev)
				dtype, err := ioutil.ReadFile(fname)
				if err != nil || strings.TrimSpace(string(dtype)) != "0" {
					if strings.Join(isVD.FindStringSubmatch(bdev), "") == "" {
						// skip unsupported devices
						continue
					}
				}
				//if strings.Contains(bdev, "dm-") {
				// skip unsupported devices
				//	continue
				//}
				entry := INIEntry{
					Section:  currentSection,
					Key:      fmt.Sprintf("%s_%s", kov[1], bdev),
					Operator: Operator(kov[2]),
					Value:    kov[3],
				}
				currentEntriesArray = append(currentEntriesArray, entry)
				currentEntriesMap[entry.Key] = entry
			}
		} else {
			// handle tunables with more than one value
			value := strings.Replace(kov[3], " ", "\t", -1)
			entry := INIEntry{
				Section:  currentSection,
				Key:      kov[1],
				Operator: Operator(kov[2]),
				Value:    value,
			}
			currentEntriesArray = append(currentEntriesArray, entry)
			currentEntriesMap[entry.Key] = entry
		}
	}
	if reminder != "" {
		// save reminder section
		// Save previous section
		if currentSection != "" {
			ret.KeyValue[currentSection] = currentEntriesMap
			ret.AllValues = append(ret.AllValues, currentEntriesArray...)
		}
		// Start the reminder section
		currentEntriesArray = make([]INIEntry, 0, 8)
		currentEntriesMap = make(map[string]INIEntry)
		currentSection = "reminder"

		entry := INIEntry{
			Section:  "reminder",
			Key:      "reminder",
			Operator: "",
			Value:    reminder,
		}
		currentEntriesArray = append(currentEntriesArray, entry)
		currentEntriesMap[entry.Key] = entry
	}

	// Save last section
	if currentSection != "" {
		ret.KeyValue[currentSection] = currentEntriesMap
		ret.AllValues = append(ret.AllValues, currentEntriesArray...)
	}
	return ret
}
