// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extvirtualservice

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extconversion"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"net/http"
)

const actionBasePath = basePath + "/actions/log"

func RegisterActionHandlers() {
	exthttp.RegisterHttpHandler(actionBasePath, exthttp.GetterAsHandler(getRobotLogActionDescription))
	exthttp.RegisterHttpHandler(actionBasePath+"/prepare", prepareLog)
	exthttp.RegisterHttpHandler(actionBasePath+"/start", startLog)
	exthttp.RegisterHttpHandler(actionBasePath+"/status", statusLog)
	exthttp.RegisterHttpHandler(actionBasePath+"/stop", stopLog)
}

func GetActionList() action_kit_api.ActionList {
	return action_kit_api.ActionList{
		Actions: []action_kit_api.DescribingEndpointReference{
			{
				Method: "GET",
				Path:   actionBasePath,
			},
		},
	}
}

func getRobotLogActionDescription() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          fmt.Sprintf("%s.log", virtualServiceTargetId),
		Label:       "log",
		Description: "collects information about the monitor status and optionally verifies that the monitor has an expected status.",
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Icon:        extutil.Ptr(targetIcon),
		TargetType:  extutil.Ptr(virtualServiceTargetId),
		TargetSelectionTemplates: extutil.Ptr([]action_kit_api.TargetSelectionTemplate{
			{
				Label: "by robot name",
				Query: "steadybit.label=\"\"",
			},
		}),
		Category:    extutil.Ptr("other"),
		Kind:        action_kit_api.Other,
		TimeControl: action_kit_api.Internal,
		Parameters: []action_kit_api.ActionParameter{
			{
				Name:         "message",
				Label:        "Message",
				Description:  extutil.Ptr("What should we log to the console? Use %s to insert the robot name."),
				Type:         action_kit_api.String,
				DefaultValue: extutil.Ptr("Hello from %s"),
				Required:     extutil.Ptr(true),
				Order:        extutil.Ptr(0),
			},
		},
		Prepare: action_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   actionBasePath + "/prepare",
		},
		Start: action_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   actionBasePath + "/start",
		},
		Status: extutil.Ptr(action_kit_api.MutatingEndpointReferenceWithCallInterval{
			Method:       "POST",
			Path:         actionBasePath + "/status",
			CallInterval: extutil.Ptr("1s"),
		}),
		Stop: extutil.Ptr(action_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   actionBasePath + "/stop",
		}),
	}
}

type LogActionState struct {
	FormattedMessage string
}

func prepareLog(w http.ResponseWriter, _ *http.Request, body []byte) {
	var request action_kit_api.PrepareActionRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse request body", err))
		return
	}

	state := LogActionState{
		FormattedMessage: fmt.Sprintf(request.Config["message"].(string), request.Target.Name),
	}

	var convertedState action_kit_api.ActionState
	err = extconversion.Convert(state, &convertedState)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to encode action state", err))
		return
	}

	formattedMessage := fmt.Sprintf("Logging in log action **prepare**: %s", state.FormattedMessage)
	log.Info().Msg(formattedMessage)

	exthttp.WriteBody(w, action_kit_api.PrepareResult{
		State: convertedState,
		Messages: extutil.Ptr([]action_kit_api.Message{
			{
				Level:   extutil.Ptr(action_kit_api.Info),
				Message: formattedMessage,
			},
		})})
}

func startLog(w http.ResponseWriter, _ *http.Request, body []byte) {
	var request action_kit_api.StartActionRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse request body", err))
		return
	}

	var state LogActionState
	err = extconversion.Convert(request.State, &state)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to convert log action state", err))
		return
	}

	formattedMessage := fmt.Sprintf("Logging in log action **start**: %s", state.FormattedMessage)
	log.Info().Msg(formattedMessage)

	exthttp.WriteBody(w, action_kit_api.StartResult{
		Messages: extutil.Ptr([]action_kit_api.Message{
			{
				Level:   extutil.Ptr(action_kit_api.Info),
				Message: formattedMessage,
			},
		}),
	})
}

func statusLog(w http.ResponseWriter, _ *http.Request, body []byte) {
	var request action_kit_api.ActionStatusRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse request body", err))
		return
	}

	var state LogActionState
	err = extconversion.Convert(request.State, &state)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to convert log action state", err))
		return
	}

	formattedMessage := fmt.Sprintf("Logging in log action **status**: %s", state.FormattedMessage)
	log.Info().Msg(formattedMessage)

	exthttp.WriteBody(w, action_kit_api.StatusResult{
		Completed: true,
		Messages: extutil.Ptr([]action_kit_api.Message{
			{
				Level:   extutil.Ptr(action_kit_api.Info),
				Message: formattedMessage,
			},
		}),
	})
}

func stopLog(w http.ResponseWriter, _ *http.Request, body []byte) {
	var request action_kit_api.StopActionRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse request body", err))
		return
	}

	var state LogActionState
	err = extconversion.Convert(request.State, &state)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to convert log action state", err))
		return
	}

	formattedMessage := fmt.Sprintf("Logging in log action **stop**: %s", state.FormattedMessage)
	log.Info().Msg(formattedMessage)

	exthttp.WriteBody(w, action_kit_api.StopResult{
		Messages: extutil.Ptr([]action_kit_api.Message{
			{
				Level:   extutil.Ptr(action_kit_api.Info),
				Message: formattedMessage,
			},
		}),
	})
}
