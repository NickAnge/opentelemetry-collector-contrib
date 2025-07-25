// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package host

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/awsutil"
	ci "github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/containerinsight"
)

type mockNodeCapacity struct{}

func (*mockNodeCapacity) getMemoryCapacity() int64 {
	return 1024
}

func (*mockNodeCapacity) getNumCores() int64 {
	return 2
}

type mockEC2Metadata struct{}

func (*mockEC2Metadata) getInstanceID() string {
	return "instance-id"
}

func (*mockEC2Metadata) getInstanceIP() string {
	return "instance-ip"
}

func (*mockEC2Metadata) getInstanceType() string {
	return "instance-type"
}

func (*mockEC2Metadata) getRegion() string {
	return "region"
}

type mockEBSVolume struct{}

func (*mockEBSVolume) getEBSVolumeID(_ string) string {
	return "ebs-volume-id"
}

func (*mockEBSVolume) extractEbsIDsUsedByKubernetes() map[string]string {
	return map[string]string{}
}

type mockEC2Tags struct{}

func (*mockEC2Tags) getClusterName() string {
	return "cluster-name"
}

func (*mockEC2Tags) getAutoScalingGroupName() string {
	return "asg"
}

func TestInfo(t *testing.T) {
	// test the case when nodeCapacity fails to initialize
	nodeCapacityCreatorOpt := func(m *Info) {
		m.nodeCapacityCreator = func(*zap.Logger, ...nodeCapacityOption) (nodeCapacityProvider, error) {
			return nil, errors.New("error")
		}
	}
	m, err := NewInfo(ci.EKS, time.Minute, zap.NewNop(), nodeCapacityCreatorOpt)
	assert.Nil(t, m)
	assert.Error(t, err)

	// test the case when aws session fails to initialize
	nodeCapacityCreatorOpt = func(m *Info) {
		m.nodeCapacityCreator = func(*zap.Logger, ...nodeCapacityOption) (nodeCapacityProvider, error) {
			return &mockNodeCapacity{}, nil
		}
	}
	awsSessionCreatorOpt := func(m *Info) {
		m.awsConfigCreator = func(context.Context, *zap.Logger, *awsutil.AWSSessionSettings) (aws.Config, error) {
			return aws.Config{}, errors.New("error")
		}
	}
	m, err = NewInfo(ci.EKS, time.Minute, zap.NewNop(), nodeCapacityCreatorOpt, awsSessionCreatorOpt)
	assert.Nil(t, m)
	assert.Error(t, err)

	// test normal case where everything is working
	awsSessionCreatorOpt = func(m *Info) {
		m.awsConfigCreator = func(context.Context, *zap.Logger, *awsutil.AWSSessionSettings) (aws.Config, error) {
			return aws.Config{}, nil
		}
	}
	ec2MetadataCreatorOpt := func(m *Info) {
		m.ec2MetadataCreator = func(context.Context, aws.Config, time.Duration, chan bool, chan bool, *zap.Logger,
			...ec2MetadataOption,
		) ec2MetadataProvider {
			return &mockEC2Metadata{}
		}
	}
	ebsVolumeCreatorOpt := func(m *Info) {
		m.ebsVolumeCreator = func(context.Context, aws.Config, string, string, time.Duration, *zap.Logger,
			...ebsVolumeOption,
		) ebsVolumeProvider {
			return &mockEBSVolume{}
		}
	}
	ec2TagsCreatorOpt := func(m *Info) {
		m.ec2TagsCreator = func(context.Context, aws.Config, string, string, string, time.Duration, *zap.Logger,
			...ec2TagsOption,
		) ec2TagsProvider {
			return &mockEC2Tags{}
		}
	}
	m, err = NewInfo(ci.EKS, time.Minute, zap.NewNop(), awsSessionCreatorOpt,
		nodeCapacityCreatorOpt, ec2MetadataCreatorOpt, ebsVolumeCreatorOpt, ec2TagsCreatorOpt)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	// before ebsVolume and ec2Tags are initialized
	assert.Empty(t, m.GetEBSVolumeID("dev"))
	assert.Empty(t, m.GetClusterName())
	assert.Empty(t, m.GetAutoScalingGroupName())

	// close the channel so that ebsVolume and ec2Tags can be initialized
	close(m.instanceIDReadyC)
	<-m.ebsVolumeReadyC
	<-m.ec2TagsReadyC

	assert.Equal(t, "instance-id", m.GetInstanceID())
	assert.Equal(t, "instance-type", m.GetInstanceType())
	assert.Equal(t, int64(2), m.GetNumCores())
	assert.Equal(t, int64(1024), m.GetMemoryCapacity())
	assert.Equal(t, "ebs-volume-id", m.GetEBSVolumeID("dev"))
	assert.Equal(t, "cluster-name", m.GetClusterName())
	assert.Equal(t, "asg", m.GetAutoScalingGroupName())
}

func TestInfoForECS(t *testing.T) {
	// test the case when nodeCapacity fails to initialize
	nodeCapacityCreatorOpt := func(m *Info) {
		m.nodeCapacityCreator = func(*zap.Logger, ...nodeCapacityOption) (nodeCapacityProvider, error) {
			return nil, errors.New("error")
		}
	}
	m, err := NewInfo(ci.ECS, time.Minute, zap.NewNop(), nodeCapacityCreatorOpt)
	assert.Nil(t, m)
	assert.Error(t, err)

	// test the case when aws session fails to initialize
	nodeCapacityCreatorOpt = func(m *Info) {
		m.nodeCapacityCreator = func(*zap.Logger, ...nodeCapacityOption) (nodeCapacityProvider, error) {
			return &mockNodeCapacity{}, nil
		}
	}
	awsSessionCreatorOpt := func(m *Info) {
		m.awsConfigCreator = func(context.Context, *zap.Logger, *awsutil.AWSSessionSettings) (aws.Config, error) {
			return aws.Config{}, errors.New("error")
		}
	}
	m, err = NewInfo(ci.ECS, time.Minute, zap.NewNop(), nodeCapacityCreatorOpt, awsSessionCreatorOpt)
	assert.Nil(t, m)
	assert.Error(t, err)

	// test normal case where everything is working
	awsSessionCreatorOpt = func(m *Info) {
		m.awsConfigCreator = func(context.Context, *zap.Logger, *awsutil.AWSSessionSettings) (aws.Config, error) {
			return aws.Config{}, nil
		}
	}
	ec2MetadataCreatorOpt := func(m *Info) {
		m.ec2MetadataCreator = func(context.Context, aws.Config, time.Duration, chan bool, chan bool, *zap.Logger,
			...ec2MetadataOption,
		) ec2MetadataProvider {
			return &mockEC2Metadata{}
		}
	}
	ebsVolumeCreatorOpt := func(m *Info) {
		m.ebsVolumeCreator = func(context.Context, aws.Config, string, string, time.Duration, *zap.Logger,
			...ebsVolumeOption,
		) ebsVolumeProvider {
			return &mockEBSVolume{}
		}
	}
	ec2TagsCreatorOpt := func(m *Info) {
		m.ec2TagsCreator = func(context.Context, aws.Config, string, string, string, time.Duration, *zap.Logger,
			...ec2TagsOption,
		) ec2TagsProvider {
			return &mockEC2Tags{}
		}
	}
	m, err = NewInfo(ci.ECS, time.Minute, zap.NewNop(), awsSessionCreatorOpt,
		nodeCapacityCreatorOpt, ec2MetadataCreatorOpt, ebsVolumeCreatorOpt, ec2TagsCreatorOpt)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	// before ebsVolume and ec2Tags are initialized
	assert.Empty(t, m.GetEBSVolumeID("dev"))
	assert.Empty(t, m.GetAutoScalingGroupName())

	// close the channel so that ebsVolume and ec2Tags can be initialized
	close(m.instanceIDReadyC)
	<-m.ebsVolumeReadyC
	<-m.ec2TagsReadyC

	assert.Equal(t, "instance-id", m.GetInstanceID())
	assert.Equal(t, "instance-type", m.GetInstanceType())
	assert.Equal(t, "instance-ip", m.GetInstanceIP())
	assert.Equal(t, int64(2), m.GetNumCores())
	assert.Equal(t, int64(1024), m.GetMemoryCapacity())
	assert.Equal(t, "ebs-volume-id", m.GetEBSVolumeID("dev"))
	assert.Equal(t, "asg", m.GetAutoScalingGroupName())
}
