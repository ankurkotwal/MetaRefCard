package sws

import (
	"bufio"
	"bytes"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/ankurkotwal/MetaRefCard/metarefcard/common"
)

var initiliased bool = false
var regexes swsRegexes
var gameData *common.GameData

// HandleRequest services the request to load files
func HandleRequest(files [][]byte,
	config *common.Config) (common.OverlaysByImage, map[string]string) {
	if !initiliased {
		gameData = common.LoadGameModel("config/sws.yaml",
			"StarWarsSquadrons Data", config.DebugOutput)
		regexes.Bind = regexp.MustCompile(gameData.Regexes["Bind"])
		regexes.Joystick = regexp.MustCompile(gameData.Regexes["Joystick"])
		initiliased = true
	}

	gameBinds, devices, contexts := loadInputFiles(files, gameData.DeviceNameMap,
		config.DebugOutput, config.VerboseOutput)
	common.GenerateContextColours(contexts, config)
	deviceMap := common.FilterDevices(devices, config)
	return populateImageOverlays(deviceMap, config.ImageMap, gameBinds, gameData), contexts
}

// Load the game config files (provided by user)
func loadInputFiles(files [][]byte, deviceNameMap common.DeviceNameFullToShort,
	debugOutput bool, verboseOutput bool) (swsBindsByDevice, common.MockSet, common.MockSet) {
	gameBinds := make(swsBindsByDevice)
	deviceNames := make(common.MockSet)
	contexts := make(common.MockSet)

	// deviceIndex: deviceId -> full name
	deviceIndex := make(map[string]string)
	contextActionIndex := make(swsContextActionIndex)

	// Load all the device and inputs
	var matches [][]string
	for idx, file := range files {
		scanner := bufio.NewScanner(bytes.NewReader(file))
		for scanner.Scan() {
			line := scanner.Text()

			matches = regexes.Bind.FindAllStringSubmatch(line, -1)
			if matches != nil {
				addAction(contextActionIndex, matches[0][1], contexts, matches[0][2],
					matches[0][3], matches[0][4], matches[0][5])
			}
			matches = regexes.Joystick.FindAllStringSubmatch(line, -1)
			if matches != nil && len(matches[0][2]) > 0 {
				if shortName, found := deviceNameMap[matches[0][2]]; !found {
					log.Printf("Error: SWS Unknown device found %s\n", matches[0][2])
					continue
				} else {
					num, err := strconv.Atoi(matches[0][1])
					// Subtract 1 from the Joystick index to match deviceIds in the file
					num--
					if err == nil && num >= 0 {
						deviceIndex[strconv.Itoa(num)] = shortName
						deviceNames[shortName] = ""
					} else {
						log.Printf("Error: SWS unexpected device number %s\n", matches[0][1])
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error: SWS scan file %d. %s\n", idx, err)
		}
	}

	// Now iterate through the object to build our internal index.
	// We do it in multiple passes to avoid having to make assumptions around
	// the order of fields in the game's config files.
	for context, actionMap := range contextActionIndex {
		for action, actionSubMap := range actionMap {
			// Get the device first
			deviceID, found := actionSubMap["deviceid"]
			if !found {
				log.Printf("Error: SWS couldn't find deviceId in %s->%s->%v\n", context, action, actionSubMap)
				continue
			}
			shortName, found := deviceIndex[deviceID]
			if !found {
				continue // We only care about devices in deviceIndex
			}
			// Add this button to index
			contextActions, found := gameBinds[shortName]
			if !found {
				contextActions = make(swsContextActions)
				gameBinds[shortName] = contextActions
			}
			actions, found := contextActions[context]
			if !found {
				actions = make(swsActions)
				contextActions[context] = actions
			}

			// Build the action details
			actionDetails := swsActionDetails{}
			for actionSub, value := range actionSubMap {
				field := getInputTypeAsField(actionSub, &actionDetails)
				if field == nil {
					log.Printf("Error: SWS unknown inputType %s value %s\n",
						actionSub, value)
				} else {
					*field = value
				}
			}

			// Assign action details accordingly
			input := interpretInput(&actionDetails)
			if len(input) > 0 {
				actions[action] = input
			} else {
				delete(actions, action)
			}
		}
	}

	return gameBinds, deviceNames, contexts
}

func addAction(contextActionIndex swsContextActionIndex,
	context string, contexts common.MockSet, action string, deviceNum string,
	actionSub string, value string) {
	contexts[context] = ""

	var found bool
	var actionMap map[string]map[string]string
	if actionMap, found = contextActionIndex[context]; !found {
		// First time for this context
		actionMap = make(map[string]map[string]string)
		contextActionIndex[context] = actionMap
	}
	var actionSubMap map[string]string
	if actionSubMap, found = actionMap[action]; !found {
		// First time for this device number
		actionSubMap = make(map[string]string)
		actionMap[action] = actionSubMap
	}
	actionSubMap[actionSub] = value
}

func getInputTypeAsField(actionSub string, currAction *swsActionDetails) *string {
	actionSub = strings.ToLower(actionSub)
	switch actionSub {
	case "altbutton":
		return &currAction.AltButton
	case "axis":
		return &currAction.Axis
	case "button":
		return &currAction.Button
	case "deviceid":
		return &currAction.DeviceID
	case "identifier":
		return &currAction.Identifier
	case "modifier":
		return &currAction.Modifier
	case "negate":
		return &currAction.Negate
	case "type":
		return &currAction.Type
	}
	return nil
}

func interpretInput(details *swsActionDetails) string {
	switch details.Axis {
	case "8":
		return "XAxis" // Throttle
	case "9":
		return "YAxis" // Stick
	case "10":
		return "XAxis" // Stick
	case "26":
		switch details.Button {
		case "46":
			fallthrough
		case "47":
			return "RZAxis" // Stick
		case "48":
			return "POV1Up" // Stick
		case "49":
			return "POV1Down" // Stick
		case "50":
			return "POV1Left" // Stick
		case "51":
			return "POV1Right" // Stick
		case "73":
			return "28" // Throttle Pinky Rocker Up
		case "74":
			return "29" // Throttle Pinky Rocker Down
		case "80":
			fallthrough // Stick TODO
		case "86":
			return ""
		}
		button, err := strconv.Atoi(details.Button)
		if err == nil {
			button -= 21 // Seems like a hardcoded number?
		}
		return strconv.Itoa(button)
	}
	log.Printf("Error SWS unknown input %v\n", details)
	return ""
}

func populateImageOverlays(deviceMap common.DeviceMap, imageMap common.ImageMap,
	gameBinds swsBindsByDevice, data *common.GameData) common.OverlaysByImage {
	// Iterate through our game binds
	overlaysByImage := make(common.OverlaysByImage)
	for shortName, gameDevice := range gameBinds {
		inputs := deviceMap[shortName]
		image := imageMap[shortName]
		for context, actions := range gameDevice {
			for actionName, input := range actions {
				inputData, found := inputs[input]
				if !found {
					log.Printf("Error: SWS unknown input to lookup %s for device %s\n",
						input, shortName)
				}
				if inputData.X == 0 && inputData.Y == 0 {
					log.Printf("Error: SWS location 0,0 for %s device %s %v\n",
						actionName, shortName, inputData)
					continue
				}
				common.GenerateImageOverlays(overlaysByImage, input, &inputData,
					gameData, actionName, context, shortName, image)
			}
		}
	}

	return overlaysByImage
}

// swsContextActionIndex: context -> action name -> action sub -> value
type swsContextActionIndex map[string]map[string]map[string]string

type swsRegexes struct {
	Bind     *regexp.Regexp
	Joystick *regexp.Regexp
}

// Device short name -> ContextAction
type swsBindsByDevice map[string]swsContextActions

// Context -> Actions
type swsContextActions map[string]swsActions

// Action -> Input
type swsActions map[string]string

// Parsed fields from sws config
type swsActionDetails struct {
	AltButton  string
	Axis       string
	Button     string
	DeviceID   string
	Identifier string
	Modifier   string
	Negate     string
	Type       string
}
