#!/bin/bash

NABU='go run ../../cmd/nabu/main.go'
CFG='../../xxx/iow/iow_dev_v4.yaml'
BLAZE='http://coreos.lan:9090/blazegraph/namespace/iow/sparql'
GRAPHDB='http://coreos.lan:7200/repositories/testing'

echo "-------  BLAZEGRAPH"
echo "----------  clear, bulk, prune "
${NABU} clear --cfg ${CFG} --dangerous --endpoint ec_blazegraph
${NABU} bulk --cfg ${CFG}  --prefix summoned/test --endpoint ec_blazegraph
${NABU} prune --cfg ${CFG}  --prefix summoned/test --endpoint ec_blazegraph

curl -H 'Accept: application/sparql-results+json' ${BLAZE} --data-urlencode 'query=SELECT (COUNT(DISTINCT ?graph) AS ?namedGraphsCount)(COUNT(*) AS ?triplesCount)WHERE {GRAPH ?graph {?subject ?predicate ?object}}' | jq '.results.bindings[0] | { namedGraphsCount: .namedGraphsCount.value, triplesCount: .triplesCount.value }' input.json -


echo "----------  clear, release, object, prune "
${NABU} clear --cfg ${CFG} --dangerous --endpoint ec_blazegraph
${NABU} release --cfg ${CFG}  --prefix summoned/test --endpoint ec_blazegraph
${NABU} object --cfg ${CFG}  graphs/latest/test_release.nq --endpoint ec_blazegraph
${NABU} prune --cfg ${CFG}  --prefix summoned/test --endpoint ec_blazegraph

curl -H 'Accept: application/sparql-results+json' ${BLAZE} --data-urlencode 'query=SELECT (COUNT(DISTINCT ?graph) AS ?namedGraphsCount)(COUNT(*) AS ?triplesCount)WHERE {GRAPH ?graph {?subject ?predicate ?object}}' | jq '.results.bindings[0] | { namedGraphsCount: .namedGraphsCount.value, triplesCount: .triplesCount.value }' input.json -


echo "----------  clear, prefix, prune "
${NABU} clear --cfg ${CFG} --dangerous --endpoint ec_blazegraph
${NABU} prefix --cfg ${CFG}  --prefix summoned/test --endpoint ec_blazegraph
${NABU} prune --cfg ${CFG}  --prefix summoned/test --endpoint ec_blazegraph

curl -H 'Accept: application/sparql-results+json' ${BLAZE} --data-urlencode 'query=SELECT (COUNT(DISTINCT ?graph) AS ?namedGraphsCount)(COUNT(*) AS ?triplesCount)WHERE {GRAPH ?graph {?subject ?predicate ?object}}' | jq '.results.bindings[0] | { namedGraphsCount: .namedGraphsCount.value, triplesCount: .triplesCount.value }' input.json -


echo "-------  GraphDB"
echo "----------  clear, bulk, prune "
${NABU} clear --cfg ${CFG} --dangerous --endpoint iow_graphdb
${NABU} bulk --cfg ${CFG}  --prefix summoned/test --endpoint iow_graphdb
${NABU} prune --cfg ${CFG}  --prefix summoned/test --endpoint iow_graphdb

curl -H 'Accept: application/sparql-results+json' ${GRAPHDB} --data-urlencode 'query=SELECT (COUNT(DISTINCT ?graph) AS ?namedGraphsCount)(COUNT(*) AS ?triplesCount)WHERE {GRAPH ?graph {?subject ?predicate ?object}}' | jq '.results.bindings[0] | { namedGraphsCount: .namedGraphsCount.value, triplesCount: .triplesCount.value }' input.json -


echo "----------  clear, release, object, prune "
${NABU} clear --cfg ${CFG} --dangerous --endpoint iow_graphdb
${NABU} release --cfg ${CFG}  --prefix summoned/test --endpoint iow_graphdb
${NABU} object --cfg ${CFG}  graphs/latest/test_release.nq --endpoint iow_graphdb
${NABU} prune --cfg ${CFG}  --prefix summoned/test --endpoint iow_graphdb

curl -H 'Accept: application/sparql-results+json' ${GRAPHDB} --data-urlencode 'query=SELECT (COUNT(DISTINCT ?graph) AS ?namedGraphsCount)(COUNT(*) AS ?triplesCount)WHERE {GRAPH ?graph {?subject ?predicate ?object}}' | jq '.results.bindings[0] | { namedGraphsCount: .namedGraphsCount.value, triplesCount: .triplesCount.value }' input.json -


echo "----------  clear, prefix, prune "
${NABU} clear --cfg ${CFG} --dangerous --endpoint iow_graphdb
${NABU} prefix --cfg ${CFG}  --prefix summoned/test --endpoint iow_graphdb
${NABU} prune --cfg ${CFG}  --prefix summoned/test --endpoint iow_graphdb

curl -H 'Accept: application/sparql-results+json' ${GRAPHDB} --data-urlencode 'query=SELECT (COUNT(DISTINCT ?graph) AS ?namedGraphsCount)(COUNT(*) AS ?triplesCount)WHERE {GRAPH ?graph {?subject ?predicate ?object}}' | jq '.results.bindings[0] | { namedGraphsCount: .namedGraphsCount.value, triplesCount: .triplesCount.value }' input.json -




echo "-------  BLAZEGRAPH PROV"
echo "----------  clear, bulk, prune "
${NABU} clear --cfg ${CFG} --dangerous --endpoint ec_blazegraph
${NABU} bulk --cfg ${CFG}  --prefix prov/counties0 --endpoint ec_blazegraph
${NABU} prune --cfg ${CFG}  --prefix prov/counties0 --endpoint ec_blazegraph

curl -H 'Accept: application/sparql-results+json' ${BLAZE} --data-urlencode 'query=SELECT (COUNT(DISTINCT ?graph) AS ?namedGraphsCount)(COUNT(*) AS ?triplesCount)WHERE {GRAPH ?graph {?subject ?predicate ?object}}' | jq '.results.bindings[0] | { namedGraphsCount: .namedGraphsCount.value, triplesCount: .triplesCount.value }' input.json -


echo "----------  clear, release, object, prune "
${NABU} clear --cfg ${CFG} --dangerous --endpoint ec_blazegraph
${NABU} release --cfg ${CFG}  --prefix prov/counties0 --endpoint ec_blazegraph
${NABU} object --cfg ${CFG}  graphs/latest/test_release.nq --endpoint ec_blazegraph
${NABU} prune --cfg ${CFG}  --prefix prov/counties0 --endpoint ec_blazegraph

curl -H 'Accept: application/sparql-results+json' ${BLAZE} --data-urlencode 'query=SELECT (COUNT(DISTINCT ?graph) AS ?namedGraphsCount)(COUNT(*) AS ?triplesCount)WHERE {GRAPH ?graph {?subject ?predicate ?object}}' | jq '.results.bindings[0] | { namedGraphsCount: .namedGraphsCount.value, triplesCount: .triplesCount.value }' input.json -


echo "----------  clear, prefix, prune "
${NABU} clear --cfg ${CFG} --dangerous --endpoint ec_blazegraph
${NABU} prefix --cfg ${CFG}  --prefix prov/counties0 --endpoint ec_blazegraph
${NABU} prune --cfg ${CFG}  --prefix prov/counties0 --endpoint ec_blazegraph

curl -H 'Accept: application/sparql-results+json' ${BLAZE} --data-urlencode 'query=SELECT (COUNT(DISTINCT ?graph) AS ?namedGraphsCount)(COUNT(*) AS ?triplesCount)WHERE {GRAPH ?graph {?subject ?predicate ?object}}' | jq '.results.bindings[0] | { namedGraphsCount: .namedGraphsCount.value, triplesCount: .triplesCount.value }' input.json -



