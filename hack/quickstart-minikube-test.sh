#!/usr/bin/env bash
set -euo pipefail

# Serving Test
echo "creating ksvc"
kn service create hello --image gcr.io/knative-samples/helloworld-go --port 8080 --env TARGET=World --revision-name=world
SERVICE=$(kn service describe hello -o url)

echo "curling ksvc 'hello'"
curl "$SERVICE"

# Eventing Tests
kn broker list
echo "creating cloudevents playuer"
kn service create cloudevents-player --image ruromero/cloudevents-player:latest --env BROKER_URL=http://broker-ingress.knative-eventing.svc.cluster.local/default/example-broker
PLAYER=$(kn service describe cloudevents-player -o url)

echo "posting event"
curl -v "$PLAYER"   -H "Content-Type: application/json"   -H "Ce-Id: foo-1"   -H "Ce-Specversion: 1.0"   -H "Ce-Type: dev.example.events"   -H "Ce-Source: curl-source"   -d '{"msg":"Hello team!"}'

echo "creating trigger"
kn trigger create cloudevents-trigger --sink cloudevents-player  --broker example-broker

echo "posting trigger event"
curl -v "$PLAYER"   -H "Content-Type: application/json"   -H "Ce-Id: foo-1"   -H "Ce-Specversion: 1.0"   -H "Ce-Type: dev.example.trigger"   -H "Ce-Source: curl-source"   -d '{"msg":"Hello team!"}'

echo "minikube test finished"
