// Copyright 2019-2021 The Liqo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package outgoing

import (
	apimgmt "github.com/liqotech/liqo/pkg/virtualKubelet/apiReflection"
	ri "github.com/liqotech/liqo/pkg/virtualKubelet/apiReflection/reflectors/reflectorsInterfaces"
	"github.com/liqotech/liqo/pkg/virtualKubelet/options"
)

var ReflectorBuilders = map[apimgmt.ApiType]func(reflector ri.APIReflector, opts map[options.OptionKey]options.Option) ri.OutgoingAPIReflector{
	apimgmt.Configmaps: configmapsReflectorBuilder,
	apimgmt.Secrets:    secretsReflectorBuilder,
}

func configmapsReflectorBuilder(reflector ri.APIReflector, _ map[options.OptionKey]options.Option) ri.OutgoingAPIReflector {
	return &ConfigmapsReflector{APIReflector: reflector}
}

func secretsReflectorBuilder(reflector ri.APIReflector, _ map[options.OptionKey]options.Option) ri.OutgoingAPIReflector {
	return &SecretsReflector{APIReflector: reflector}
}
