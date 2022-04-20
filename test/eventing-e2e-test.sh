#!/usr/bin/env bash

# Copyright 2021 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

# Eventing Tests
kn broker list
echo "creating cloudevents player"
kn service create cloudevents-player --image ruromero/cloudevents-player:latest --env BROKER_URL=http://broker-ingress.knative-eventing.svc.cluster.local/default/example-broker
PLAYER=$(kn service describe cloudevents-player -o url)

echo "posting event"
curl -v "$PLAYER"   -H "Content-Type: application/json"   -H "Ce-Id: foo-1"   -H "Ce-Specversion: 1.0"   -H "Ce-Type: dev.example.events"   -H "Ce-Source: curl-source"   -d '{"msg":"Hello team!"}'

curl -v "$PLAYER"/messages

echo "creating trigger"
kn trigger create cloudevents-trigger --sink cloudevents-player  --broker example-broker

echo "posting trigger event"
curl -v "$PLAYER"   -H "Content-Type: application/json"   -H "Ce-Id: foo-1"   -H "Ce-Specversion: 1.0"   -H "Ce-Type: dev.example.trigger"   -H "Ce-Source: curl-source"   -d '{"msg":"Hello team!"}'

echo "test finished!"
