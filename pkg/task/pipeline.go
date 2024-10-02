/*
 * Copyright 2024 CoreLayer BV
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package task

import "sync"

type Pipeline struct {
	Data map[string]interface{}
	mux  sync.RWMutex
}

func (p *Pipeline) Set(key string, value interface{}) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.Data[key] = value
}

func (p *Pipeline) Get(key string) interface{} {
	p.mux.RLock()
	defer p.mux.RUnlock()
	return p.Data[key]
}
