/*
Copyright 2017 The Kubernetes Authors.

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

package azure

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/stretchr/testify/mock"
)

const (
	fakeVirtualMachineScaleSetVMID = "/subscriptions/test-subscription-id/resourceGroups/test-asg/providers/Microsoft.Compute/virtualMachineScaleSets/agents/virtualMachines/%d"
)

// DeploymentsClientMock mocks for DeploymentsClient.
type DeploymentsClientMock struct {
	mock.Mock

	mutex     sync.Mutex
	FakeStore map[string]resources.DeploymentExtended
}

// Get gets the DeploymentExtended by deploymentName.
func (m *DeploymentsClientMock) Get(ctx context.Context, resourceGroupName string, deploymentName string) (result resources.DeploymentExtended, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	deploy, ok := m.FakeStore[deploymentName]
	if !ok {
		return result, fmt.Errorf("deployment not found")
	}

	return deploy, nil
}

// ExportTemplate exports the deployment's template.
func (m *DeploymentsClientMock) ExportTemplate(ctx context.Context, resourceGroupName string, deploymentName string) (result resources.DeploymentExportResult, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	deploy, ok := m.FakeStore[deploymentName]
	if !ok {
		return result, fmt.Errorf("deployment not found")
	}

	return resources.DeploymentExportResult{
		Template: deploy.Properties.Template,
	}, nil
}

// CreateOrUpdate creates or updates the Deployment.
func (m *DeploymentsClientMock) CreateOrUpdate(ctx context.Context, resourceGroupName string, deploymentName string, parameters resources.Deployment) (resp *http.Response, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	deploy, ok := m.FakeStore[deploymentName]
	if !ok {
		deploy = resources.DeploymentExtended{
			Properties: &resources.DeploymentPropertiesExtended{},
		}
		m.FakeStore[deploymentName] = deploy
	}

	deploy.Properties.Parameters = parameters.Properties.Parameters
	deploy.Properties.Template = parameters.Properties.Template
	return &http.Response{StatusCode: 200}, nil
}

// List gets all the deployments for a resource group.
func (m *DeploymentsClientMock) List(ctx context.Context, resourceGroupName, filter string, top *int32) (result []resources.DeploymentExtended, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	result = make([]resources.DeploymentExtended, 0)
	for i := range m.FakeStore {
		result = append(result, m.FakeStore[i])
	}

	return result, nil
}

// Delete deletes the given deployment
func (m *DeploymentsClientMock) Delete(ctx context.Context, resourceGroupName, deploymentName string) (resp *http.Response, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.FakeStore[deploymentName]; !ok {
		return nil, fmt.Errorf("there is no such a deployment with name %s", deploymentName)
	}

	delete(m.FakeStore, deploymentName)

	return
}

func fakeVMSSWithTags(vmssName string, tags map[string]*string) compute.VirtualMachineScaleSet {
	skuName := "Standard_D4_v2"
	var vmssCapacity int64 = 3

	return compute.VirtualMachineScaleSet{
		Name: &vmssName,
		Sku: &compute.Sku{
			Capacity: &vmssCapacity,
			Name:     &skuName,
		},
		Tags: tags,
	}

}
