// Copyright 2020 The PipeCD Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package executor

import (
	"context"

	"go.uber.org/zap"

	"github.com/pipe-cd/pipe/pkg/cache"
	"github.com/pipe-cd/pipe/pkg/config"
	"github.com/pipe-cd/pipe/pkg/model"
)

type Executor interface {
	// Execute starts running executor until completion
	// or the StopSignal has emitted.
	Execute(sig StopSignal) model.StageStatus
}

type Factory func(in Input) Executor

type LogPersister interface {
	Append(log string, s model.LogSeverity)
	AppendInfo(log string)
	AppendSuccess(log string)
	AppendError(log string)
}

type MetadataStore interface {
	Get(key string) (string, bool)
	Set(ctx context.Context, key, value string) error

	GetStageMetadata(stageID string) (map[string]string, bool)
	SetStageMetadata(ctx context.Context, stageID string, metadata map[string]string) error
}

type CommandLister interface {
	ListCommands() []model.ReportableCommand
}

type AppLiveResourceLister interface {
	ListKubernetesResources() ([]*model.KubernetesResourceState, bool)
}

type Input struct {
	Stage       *model.PipelineStage
	StageConfig config.PipelineStage
	// Readonly deployment model.
	Deployment       *model.Deployment
	DeploymentConfig *config.Config
	PipedConfig      *config.PipedSpec
	Application      *model.Application
	WorkingDir       string
	// The path to the directory containing repository's source code at target commit.
	RepoDir string
	// The path to the directory containing repository's source code at running commit.
	// This directory is valid only when the Deployment.RunningCommitHash is not empty.
	RunningRepoDir        string
	StageWorkingDir       string
	CommandLister         CommandLister
	LogPersister          LogPersister
	MetadataStore         MetadataStore
	AppManifestsCache     cache.Cache
	AppLiveResourceLister AppLiveResourceLister
	Logger                *zap.Logger
}

func DetermineStageStatus(sig StopSignalType, ori, got model.StageStatus) model.StageStatus {
	switch sig {
	case StopSignalNone:
		return got
	case StopSignalTerminate:
		return ori
	case StopSignalCancel:
		return model.StageStatus_STAGE_CANCELLED
	case StopSignalTimeout:
		return model.StageStatus_STAGE_FAILURE
	}
	return model.StageStatus_STAGE_FAILURE
}
