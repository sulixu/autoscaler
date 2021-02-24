/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

// GetNodeGroupSizeMap return a map of node group id and its target size
func GetNodeGroupSizeMap(cloudProvider cloudprovider.CloudProvider) map[string]int {
	nodeGroupSize := make(map[string]int)
	for _, nodeGroup := range cloudProvider.NodeGroups() {
		size, err := nodeGroup.TargetSize()
		if err != nil {
			klog.Errorf("Error while checking node group size %s: %v", nodeGroup.Id(), err)
			continue
		}
		nodeGroupSize[nodeGroup.Id()] = size
	}
	return nodeGroupSize
}

// FilterOutNodes filters out nodesToFilterOut from nodes
func FilterOutNodes(nodes []*apiv1.Node, nodesToFilterOut []*apiv1.Node) []*apiv1.Node {
	var filtered []*apiv1.Node
	for _, node := range nodes {
		found := false
		for _, nodeToFilter := range nodesToFilterOut {
			if nodeToFilter.Name == node.Name {
				found = true
			}
		}
		if !found {
			filtered = append(filtered, node)
		}
	}

	return filtered
}

const (
	// ResourceENI is a resource name for ENI, which represents a dedicated eni
	ResourceENI apiv1.ResourceName = "pinterest.com/eni"

	// ResourceIP is a resource name for IP, which represents a routable ip on shared eni
	ResourceIP apiv1.ResourceName = "pinterest.com/ip"

	// ResourceBridgePort is a resource name for bridge port, which represents a slot in docker bridge
	ResourceBridgePort apiv1.ResourceName = "pinterest.com/bridge-port"
)

var (
	// QuantityNodeBridgePortCount is a Quantity representation of NodeBridgePortCount
	QuantityNodeBridgePortCount = resource.MustParse("253")

	maxRoutableIPByNodeType = map[string]resource.Quantity{
		"c5.9xlarge":    resource.MustParse("29"),
		"c5.24xlarge":   resource.MustParse("49"),
		"c5.metal":      resource.MustParse("49"),
		"c5d.9xlarge":   resource.MustParse("29"),
		"c5d.24xlarge":  resource.MustParse("49"),
		"c5d.metal":     resource.MustParse("49"),
		"m5.24xlarge":   resource.MustParse("49"),
		"m5.metal":      resource.MustParse("49"),
		"r5.12xlarge":   resource.MustParse("29"),
		"r5.24xlarge":   resource.MustParse("49"),
		"r5.metal":      resource.MustParse("49"),
		"p3.16xlarge":   resource.MustParse("29"),
		"p3dn.24xlarge": resource.MustParse("49"),
		"x1.32xlarge":   resource.MustParse("29"),
		"default":       resource.MustParse("29"),
	}

	maxDedicatedENIByNodeType = map[string]resource.Quantity{
		"c5.9xlarge":    resource.MustParse("6"),
		"c5.24xlarge":   resource.MustParse("13"),
		"c5.metal":      resource.MustParse("13"),
		"c5d.9xlarge":   resource.MustParse("6"),
		"c5d.24xlarge":  resource.MustParse("13"),
		"c5d.metal":     resource.MustParse("13"),
		"m5.24xlarge":   resource.MustParse("13"),
		"m5.metal":      resource.MustParse("13"),
		"r5.12xlarge":   resource.MustParse("5"),
		"r5.24xlarge":   resource.MustParse("13"),
		"r5.metal":      resource.MustParse("13"),
		"p3.16xlarge":   resource.MustParse("5"),
		"p3dn.24xlarge": resource.MustParse("13"),
		"x1.32xlarge":   resource.MustParse("5"),
		"default":       resource.MustParse("5"),
	}

)

func MaxNetworkResourceFromNode(node *apiv1.Node) {
	// we need to maximize the numbers as network resources can change dynamically
	instanceType := NodeInstanceType(node)

	node.Status.Allocatable[ResourceIP] = MaxRoutableIPForNodeType(instanceType)
	node.Status.Allocatable[ResourceENI] = MaxDedicatedENIForNodeType(instanceType)
	node.Status.Allocatable[ResourceBridgePort] = QuantityNodeBridgePortCount

	node.Status.Capacity[ResourceIP] = MaxRoutableIPForNodeType(instanceType)
	node.Status.Capacity[ResourceENI] = MaxDedicatedENIForNodeType(instanceType)
	node.Status.Capacity[ResourceBridgePort] = QuantityNodeBridgePortCount
}

func NodeInstanceType(node *apiv1.Node) string {
	instanceType, ok := node.Labels[apiv1.LabelInstanceType]
	if !ok {
		return ""
	}
	return instanceType
}

// MaxRoutableIPForNodeType returns maximum routable ip count for the given node type, if the node type is
// unknown, it returns a default value for estimation
func MaxRoutableIPForNodeType(nodeType string) resource.Quantity {
	if num, ok := maxRoutableIPByNodeType[nodeType]; ok {
		return num
	}
	return maxRoutableIPByNodeType["default"]
}

// MaxDedicatedENIForNodeType returns maximum dedicated eni count for the given node type, if the node type is
// unknown, it returns a default value for estimation
func MaxDedicatedENIForNodeType(nodeType string) resource.Quantity {
	if num, ok := maxDedicatedENIByNodeType[nodeType]; ok {
		return num
	}
	return maxDedicatedENIByNodeType["default"]
}
