// Copyright (C) 2022 Satoshi Konno All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package redis

import (
	"errors"

	"github.com/cybergarage/go-redis/redis/proto"
)

func parseHashArg(cmd string, args Arguments) (string, error) {
	hash, err := args.NextString()
	if err != nil {
		return "", newMissingArgumentError(cmd, "hash", err)
	}
	return hash, nil
}

func parseKeyArg(cmd string, args Arguments) (string, error) {
	key, err := args.NextString()
	if err != nil {
		return "", newMissingArgumentError(cmd, "key", err)
	}
	return key, nil
}

func parseSetArgs(cmd string, args Arguments) (string, string, error) {
	key, err := args.NextString()
	if err != nil {
		return "", "", newMissingArgumentError(cmd, "key", err)
	}
	val, err := args.NextString()
	if err != nil {
		return "", "", newMissingArgumentError(cmd, "value", err)
	}
	return key, val, err
}

func parseMSetArgs(cmd string, args Arguments) (map[string]string, error) {
	var key, val string
	var err error
	dir := map[string]string{}
	key, err = args.NextString()
	for err == nil {
		val, err = args.NextString()
		if err != nil {
			newMissingArgumentError(cmd, key, err)
		}
		dir[key] = val
		key, err = args.NextString()
	}
	if !errors.Is(err, proto.ErrEOM) {
		return nil, err
	}
	return dir, nil
}

func parseMGetArgs(cmd string, args Arguments) ([]string, error) {
	var key string
	var err error
	keys := []string{}
	key, err = args.NextString()
	for err == nil {
		keys = append(keys, key)
		key, err = args.NextString()
	}
	if !errors.Is(err, proto.ErrEOM) {
		return nil, err
	}
	return keys, nil
}