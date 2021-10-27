#!/usr/bin/env bash
set -euo pipefail

# Serving Test
echo "creating ksvc"
kn service create hello --image gcr.io/knative-samples/helloworld-go --port 8080 --env TARGET=World --revision-name=world

echo "curling ksvc 'hello'"
curl http://hello.default.127.0.0.1.nip.io

# Eventing Tests
kn broker list
echo "creating cloudevents playuer"
kn service create cloudevents-player --image ruromero/cloudevents-player:latest --env BROKER_URL=http://broker-ingress.knative-eventing.svc.cluster.local/default/example-broker

echo "posting event"
curl -v http://cloudevents-player.default.127.0.0.1.nip.io   -H "Content-Type: application/json"   -H "Ce-Id: foo-1"   -H "Ce-Specversion: 1.0"   -H "Ce-Type: dev.example.events"   -H "Ce-Source: curl-source"   -d '{"msg":"Hello team!"}'

echo "creating trigger"
kn trigger create cloudevents-trigger --sink cloudevents-player  --broker example-broker

echo "posting trigger event"
curl -v http://cloudevents-player.default.127.0.0.1.nip.io   -H "Content-Type: application/json"   -H "Ce-Id: foo-1"   -H "Ce-Specversion: 1.0"   -H "Ce-Type: dev.example.events-trigger"   -H "Ce-Source: curl-source"   -d '{"msg":"Hello team!"}'

echo "kind test finished"
