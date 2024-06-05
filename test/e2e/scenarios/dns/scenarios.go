// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package dns

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/microsoft/retina/test/e2e/common"
	"github.com/microsoft/retina/test/e2e/framework/kubernetes"
	"github.com/microsoft/retina/test/e2e/framework/types"
)

const (
	sleepDelay    = 5 * time.Second
	EmptyResponse = "emptyResponse"
)

type RequestValidationParams struct {
	NumResponse string
	Query       string
	QueryType   string

	Command     string
	ExpectError bool
}

type ResponseValidationParams struct {
	NumResponse string
	Query       string
	QueryType   string
	ReturnCode  string
	Response    string
}

// ValidateBasicDNSMetrics validates basic DNS metrics present in the metrics endpoint
func ValidateBasicDNSMetrics(scenarioName string, req *RequestValidationParams, resp *ResponseValidationParams) *types.Scenario {
	// generate a random ID using rand
	id := fmt.Sprintf("basic-dns-port-forward-%d", rand.Int()) // nolint:gosec // fine to use math/rand here
	agnhostName := "agnhost-" + id
	podName := agnhostName + "-0"
	steps := []*types.StepWrapper{
		{
			Step: &kubernetes.CreateAgnhostStatefulSet{
				AgnhostName:      agnhostName,
				AgnhostNamespace: "kube-system",
			},
		},
		{
			Step: &kubernetes.ExecInPod{
				PodName:      podName,
				PodNamespace: "kube-system",
				Command:      req.Command,
			},
			Opts: &types.StepOptions{
				ExpectError:               req.ExpectError,
				SkipSavingParamatersToJob: true,
			},
		},
		{
			Step: &types.Sleep{
				Duration: sleepDelay,
			},
		},
		// Ref: https://github.com/microsoft/retina/issues/415
		{
			Step: &kubernetes.ExecInPod{
				PodName:      podName,
				PodNamespace: "kube-system",
				Command:      req.Command,
			},
			Opts: &types.StepOptions{
				ExpectError:               req.ExpectError,
				SkipSavingParamatersToJob: true,
			},
		},
		{
			Step: &types.Sleep{
				Duration: sleepDelay,
			},
		},
		{
			Step: &kubernetes.PortForward{
				Namespace:             "kube-system",
				LabelSelector:         "k8s-app=retina",
				LocalPort:             strconv.Itoa(common.RetinaPort),
				RemotePort:            strconv.Itoa(common.RetinaPort),
				Endpoint:              "metrics",
				OptionalLabelAffinity: "app=" + agnhostName, // port forward to a pod on a node that also has this pod with this label, assuming same namespace
			},
			Opts: &types.StepOptions{
				SkipSavingParamatersToJob: true,
				RunInBackgroundWithID:     id,
			},
		},
		{
			Step: &validateBasicDNSRequestMetrics{
				NumResponse: req.NumResponse,
				Query:       req.Query,
				QueryType:   req.QueryType,
			},
			Opts: &types.StepOptions{
				SkipSavingParamatersToJob: true,
			},
		},
		{
			Step: &validateBasicDNSResponseMetrics{
				NumResponse: resp.NumResponse,
				Query:       resp.Query,
				QueryType:   resp.QueryType,
				ReturnCode:  resp.ReturnCode,
				Response:    resp.Response,
			},
			Opts: &types.StepOptions{
				SkipSavingParamatersToJob: true,
			},
		},
		{
			Step: &types.Stop{
				BackgroundID: id,
			},
		},
		{
			Step: &kubernetes.DeleteKubernetesResource{
				ResourceType:      kubernetes.TypeString(kubernetes.StatefulSet),
				ResourceName:      agnhostName,
				ResourceNamespace: "kube-system",
			}, Opts: &types.StepOptions{
				SkipSavingParamatersToJob: true,
			},
		},
		{
			Step: &types.Sleep{
				Duration: sleepDelay,
			},
		},
	}
	return types.NewScenario(scenarioName, steps...)
}

// ValidateAdvancedDNSMetrics validates the advanced DNS metrics present in the metrics endpoint
func ValidateAdvancedDNSMetrics(scenarioName string, req *RequestValidationParams, resp *ResponseValidationParams, kubeConfigFilePath string) *types.Scenario {
	// random ID
	id := fmt.Sprintf("adv-dns-port-forward-%d", rand.Int()) // nolint:gosec // fine to use math/rand here
	agnhostName := "agnhost-" + id
	podName := agnhostName + "-0"
	steps := []*types.StepWrapper{
		{
			Step: &kubernetes.CreateAgnhostStatefulSet{
				AgnhostName:      agnhostName,
				AgnhostNamespace: "kube-system",
			},
		},
		{
			Step: &kubernetes.ExecInPod{
				PodName:      podName,
				PodNamespace: "kube-system",
				Command:      req.Command,
			},
			Opts: &types.StepOptions{
				ExpectError:               req.ExpectError,
				SkipSavingParamatersToJob: true,
			},
		},
		{
			Step: &types.Sleep{
				Duration: sleepDelay,
			},
		},
		// Ref: https://github.com/microsoft/retina/issues/415
		{
			Step: &kubernetes.ExecInPod{
				PodName:      podName,
				PodNamespace: "kube-system",
				Command:      req.Command,
			},
			Opts: &types.StepOptions{
				ExpectError:               req.ExpectError,
				SkipSavingParamatersToJob: true,
			},
		},
		{
			Step: &types.Sleep{
				Duration: sleepDelay,
			},
		},
		{
			Step: &kubernetes.PortForward{
				Namespace:             "kube-system",
				LabelSelector:         "k8s-app=retina",
				LocalPort:             strconv.Itoa(common.RetinaPort),
				RemotePort:            strconv.Itoa(common.RetinaPort),
				Endpoint:              "metrics",
				OptionalLabelAffinity: "app=" + agnhostName, // port forward to a pod on a node that also has this pod with this label, assuming same namespace
			},
			Opts: &types.StepOptions{
				SkipSavingParamatersToJob: true,
				RunInBackgroundWithID:     id,
			},
		},
		{
			Step: &ValidateAdvancedDNSRequestMetrics{
				Namespace:          "kube-system",
				NumResponse:        req.NumResponse,
				PodName:            podName,
				Query:              req.Query,
				QueryType:          req.QueryType,
				WorkloadKind:       "StatefulSet",
				WorkloadName:       agnhostName,
				KubeConfigFilePath: kubeConfigFilePath,
			},
			Opts: &types.StepOptions{
				SkipSavingParamatersToJob: true,
			},
		},
		{
			Step: &ValidateAdvanceDNSResponseMetrics{
				Namespace:          "kube-system",
				NumResponse:        resp.NumResponse,
				PodName:            podName,
				Query:              resp.Query,
				QueryType:          resp.QueryType,
				Response:           resp.Response,
				ReturnCode:         resp.ReturnCode,
				WorkloadKind:       "StatefulSet",
				WorkloadName:       agnhostName,
				KubeConfigFilePath: kubeConfigFilePath,
			},
			Opts: &types.StepOptions{
				SkipSavingParamatersToJob: true,
			},
		},
		{
			Step: &types.Stop{
				BackgroundID: id,
			},
		},
		{
			Step: &kubernetes.DeleteKubernetesResource{
				ResourceType:      kubernetes.TypeString(kubernetes.StatefulSet),
				ResourceName:      agnhostName,
				ResourceNamespace: "kube-system",
			}, Opts: &types.StepOptions{
				SkipSavingParamatersToJob: true,
			},
		},
		{
			Step: &types.Sleep{
				Duration: sleepDelay,
			},
		},
	}
	return types.NewScenario(scenarioName, steps...)
}
