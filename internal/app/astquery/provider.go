package astquery

import (
	"encoding/xml"
	"sort"
	"strconv"

	"github.com/checkmarxDev/ast-sast-export/internal/app/common"
	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"
	"github.com/checkmarxDev/ast-sast-export/internal/app/querymapping"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
	"github.com/rs/zerolog/log"
)

const (
	notCustomPackageType = "Cx"
)

type (
	Provider struct {
		queryProvider interfaces.QueriesRepo
		mapping       []querymapping.QueryMap
	}
)

func NewProvider(queryProvider interfaces.QueriesRepo, queryMappingProvider interfaces.QueryMappingRepo) (*Provider, error) {
	return &Provider{
		queryProvider: queryProvider,
		mapping:       queryMappingProvider.GetMapping(),
	}, nil
}

func (e *Provider) GetQueryID(language, name, group, sastQueryID string) (string, error) {
	mappedAstID := e.getMappedID(sastQueryID)
	if mappedAstID != "" {
		return mappedAstID, nil
	}
	return common.GetAstQueryID(language, name, group)
}

func (e *Provider) GetCustomQueriesList() (*soap.GetQueryCollectionResponse, error) {
	var output soap.GetQueryCollectionResponse
	queryResponse, err := e.queryProvider.GetQueriesList()
	if err != nil {
		return nil, err
	}

	output.XMLName = xml.Name{Local: "GetQueryCollectionResponse"}
	output.GetQueryCollectionResult.IsSuccessful = true
	output.GetQueryCollectionResult.XMLName = xml.Name{Local: "GetQueryCollectionResult"}
	output.GetQueryCollectionResult.QueryGroups.XMLName = xml.Name{Local: "QueryGroups"}
	output.GetQueryCollectionResult.QueryGroups.CxWSQueryGroup = []soap.CxWSQueryGroup{}

	//nolint:gocritic
	for _, v := range queryResponse.GetQueryCollectionResult.QueryGroups.CxWSQueryGroup {
		if v.PackageType != notCustomPackageType {
			output.GetQueryCollectionResult.QueryGroups.CxWSQueryGroup =
				append(output.GetQueryCollectionResult.QueryGroups.CxWSQueryGroup, v)
		}
	}

	return &output, nil
}

// defaultStates returns the list of default states used for comparison in both GetCustomStatesList and GetStateMapping.
func defaultStates() []soap.ResultState {
	return []soap.ResultState{
		{
			ResultName:       "To Verify",
			ResultID:         0,
			ResultPermission: "set-result-state-toverify",
		},
		{
			ResultName:       "Not Exploitable",
			ResultID:         1,
			ResultPermission: "set-result-state-notexploitable",
		},
		{
			ResultName:       "Confirmed",
			ResultID:         2,
			ResultPermission: "set-result-state-confirmed",
		},
		{
			ResultName:       "Urgent",
			ResultID:         3,
			ResultPermission: "set-result-state-urgent",
		},
		{
			ResultName:       "Proposed Not Exploitable",
			ResultID:         4,
			ResultPermission: "set-result-state-proposednotexploitable",
		},
	}
}

func (e *Provider) GetCustomStatesList() (*soap.GetResultStateListResponse, error) {
	var output soap.GetResultStateListResponse

	log.Info().Msgf("Fetching custom states list from SOAP")
	statesResponse, err := e.queryProvider.GetCustomStatesList()
	if err != nil {
		log.Error().Msgf("Failed to fetch custom states list: %v", err)
		return nil, err
	}

	// Populate the SOAP response structure
	output.XMLName = xml.Name{Local: "GetCustomStatesResponse"}
	output.GetResultStateListResult.XMLName = xml.Name{Local: "GetCustomStatesResult"}
	output.GetResultStateListResult.ResultStateList.XMLName = xml.Name{Local: "GetResultStateListResult"}
	output.GetResultStateListResult.ResultStateList.ResultState = []soap.ResultState{}

	defaultStates := defaultStates()

	// Create maps for quick lookup: by ResultName and by ResultID
	defaultStateByName := make(map[string]soap.ResultState)
	defaultStateByID := make(map[int]soap.ResultState)
	for _, state := range defaultStates {
		defaultStateByName[state.ResultName] = state
		defaultStateByID[state.ResultID] = state
	}

	// Get the states from the SOAP response
	responseStates := statesResponse.GetResultStateListResult.ResultStateList.ResultState

	// Find the highest ResultID to assign new IDs for overwritten states
	maxID := 4 // Start with the highest default ID
	for _, state := range responseStates {
		if state.ResultID > maxID {
			maxID = state.ResultID
		}
	}

	// Process the SOAP response states
	newStates := []soap.ResultState{}
	for _, state := range responseStates {
		// Check if this state's ResultID matches a default state
		if defaultStateByID, existsByID := defaultStateByID[state.ResultID]; existsByID {
			// Check if the ResultName matches the default state for this ID
			if state.ResultName != defaultStateByID.ResultName {
				// Overwrite detected: same ID, different name
				log.Warn().Msgf("Detected overwrite for ID %d: default name '%s', SOAP name '%s'",
					state.ResultID, defaultStateByID.ResultName, state.ResultName)
				maxID++
				newState := soap.ResultState{
					ResultName:       state.ResultName,
					ResultID:         maxID,
					ResultPermission: state.ResultPermission,
				}
				newStates = append(newStates, newState)
				// Ensure the default state is included
				if !containsState(newStates, defaultStateByID.ResultID) {
					newStates = append(newStates, defaultStateByID)
				}
				continue
			}
		}

		// Check if this state's ResultName matches a default state
		if defaultStateByName, existsByName := defaultStateByName[state.ResultName]; existsByName {
			if state.ResultID != defaultStateByName.ResultID {
				// Overwrite detected: same name, different ID
				maxID++
				newState := soap.ResultState{
					ResultName:       state.ResultName,
					ResultID:         maxID,
					ResultPermission: state.ResultPermission,
				}
				log.Info().Msgf("Assigned new ID %d to overwritten state '%s'", maxID, state.ResultName)
				newStates = append(newStates, newState)
				// Ensure the default state is included
				if !containsState(newStates, defaultStateByName.ResultID) {
					log.Info().Msgf("Adding original default state: %s (ID: %d)", defaultStateByName.ResultName, defaultStateByName.ResultID)
					newStates = append(newStates, defaultStateByName)
				}
			} else {
				// Not overwritten (same name and ID), keep as is
				newStates = append(newStates, state)
			}
		} else {
			// Not a default state name, treat as a custom state and keep as is
			newStates = append(newStates, state)
		}
	}

	// Add any missing default states that weren't in the SOAP response
	for _, defaultState := range defaultStates {
		if !containsState(newStates, defaultState.ResultID) {
			newStates = append(newStates, defaultState)
		}
	}

	// Sort states by ResultID for consistency
	sort.Slice(newStates, func(i, j int) bool {
		return newStates[i].ResultID < newStates[j].ResultID
	})

	// Add the processed states to the output
	output.GetResultStateListResult.ResultStateList.ResultState = newStates

	return &output, nil
}

func (e *Provider) GetRawCustomStatesList() (*soap.GetResultStateListResponse, error) {
	statesResponse, err := e.queryProvider.GetCustomStatesList()
	if err != nil {
		log.Error().Msgf("Failed to fetch raw custom states list: %v", err)
		return nil, err
	}

	return statesResponse, nil
}

func (e *Provider) GetStateMapping() (map[string]string, error) {
	stateMapping := make(map[string]string)

	// Fetch the raw custom states list
	statesResponse, err := e.GetRawCustomStatesList()
	if err != nil {
		return nil, err
	}

	// Get the default states
	defaultStates := defaultStates()

	// Create maps for quick lookup: by ResultName and by ResultID
	defaultStateByName := make(map[string]soap.ResultState)
	defaultStateByID := make(map[int]soap.ResultState)
	for _, state := range defaultStates {
		defaultStateByName[state.ResultName] = state
		defaultStateByID[state.ResultID] = state
	}

	// Get the states from the raw SOAP response
	responseStates := statesResponse.GetResultStateListResult.ResultStateList.ResultState

	// Find the highest ResultID to assign new IDs for overwritten states
	maxID := 4 // Start with the highest default ID
	for _, state := range responseStates {
		if state.ResultID > maxID {
			maxID = state.ResultID
		}
	}
	// Process the SOAP response states to build the mapping
	for _, state := range responseStates {
		// Check if this state's ResultID matches a default state
		if defaultStateByID, existsByID := defaultStateByID[state.ResultID]; existsByID {
			// Check if the ResultName matches the default state for this ID
			if state.ResultName != defaultStateByID.ResultName {
				// Overwrite detected: same ID, different name
				maxID++
				// Add to mapping: old ID -> new ID
				stateMapping[strconv.Itoa(defaultStateByID.ResultID)] = strconv.Itoa(maxID)
				continue
			}
		}

		// Check if this state's ResultName matches a default state
		if defaultStateByName, existsByName := defaultStateByName[state.ResultName]; existsByName {
			if state.ResultID != defaultStateByName.ResultID {
				// Overwrite detected: same name, different ID
				log.Warn().Msgf("Detected overwrite for state '%s': default ID %d, SOAP ID %d",
					state.ResultName, defaultStateByName.ResultID, state.ResultID)
				maxID++
				// Add to mapping: old ID -> new ID
				stateMapping[strconv.Itoa(defaultStateByName.ResultID)] = strconv.Itoa(maxID)
			}
		}
	}
	return stateMapping, nil
}

// Helper function to check if a state with a specific ResultID exists in the list
func containsState(states []soap.ResultState, resultID int) bool {
	for _, state := range states {
		if state.ResultID == resultID {
			return true
		}
	}
	return false
}

func (e *Provider) getMappedID(sastID string) string {
	for _, queryMap := range e.mapping {
		if queryMap.SastID == sastID {
			return queryMap.AstID
		}
	}
	return ""
}
