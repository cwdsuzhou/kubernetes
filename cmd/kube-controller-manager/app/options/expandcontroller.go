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

package options

import (
	"time"

	"github.com/spf13/pflag"

	expandconfig "k8s.io/kubernetes/pkg/controller/volume/expand/config"
)

// ExpandControllerOptions holds the ExpandController options.
type ExpandControllerOptions struct {
	*expandconfig.ExpandControllerConfiguration
}

// AddFlags adds flags related to ExpandController for controller manager to the specified FlagSet.
func (o *ExpandControllerOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.DurationVar(&o.VolumeOperationMaxBackoff.Duration, "expand-max-backoff-time", 2*time.Minute+2*time.Second, "<Warning: Alpha feature> The maximum backoff time of expand. If it is not specified, it will not be applied.")
}

// ApplyTo fills up ExpandController config with options.
func (o *ExpandControllerOptions) ApplyTo(cfg *expandconfig.ExpandControllerConfiguration) error {
	if o == nil {
		return nil
	}

	cfg.VolumeOperationMaxBackoff = o.VolumeOperationMaxBackoff

	return nil
}

// Validate checks validation of ExpandControllerOptions.
func (o *ExpandControllerOptions) Validate() []error {
	if o == nil {
		return nil
	}

	errs := []error{}
	return errs
}
