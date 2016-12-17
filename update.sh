#!/bin/bash

stackID=$1
serviceID=$2
#check the current state
state=$(curl -u "${CATTLE_ACCESS_KEY}:${CATTLE_SECRET_KEY}" \
        -X GET \
        -H 'Accept: application/json' \
        -H 'Content-Type: application/json' \
        "https://rancherdev.io.nuxeo.com/v2-beta/projects/$stackID/services/$serviceID/" | jq '.state')
echo "----------------------------------------"
echo "$state"
echo "----------------------------------------  "
if [ "$state" != '"active"' ] && [ "$state" != '"inactive"' ] 
    then
      echo "Unkonw state: $state, can not upgrade"; 
      exit 1;
fi      

#needs this parameter to be send back for the upgrade
inServiceStrategy=$(curl -u "${CATTLE_ACCESS_KEY}:${CATTLE_SECRET_KEY}" \
        -X GET \
        -H 'Accept: application/json' \
        -H 'Content-Type: application/json' \
        "https://rancherdev.io.nuxeo.com/v2-beta/projects/$stackID/services/$serviceID/" | jq '.upgrade.inServiceStrategy')

#upgrade

curl -u "${CATTLE_ACCESS_KEY}:${CATTLE_SECRET_KEY}" \
        -X POST \
        -H 'Accept: application/json' \
        -H 'Content-Type: application/json' \
        -d "{
          \"inServiceStrategy\": ${inServiceStrategy}
          }
        }" \
        "https://rancherdev.io.nuxeo.com/v2-beta/projects/$stackID/services/$serviceID/?action=upgrade"

# pooling the state while waiting for the service to be upgrade
state="upgrading"
while [ "$state" != '"upgraded"' ]
  do
   state=$(curl -u "${CATTLE_ACCESS_KEY}:${CATTLE_SECRET_KEY}" \
        -X GET \
        -H 'Accept: application/json' \
        -H 'Content-Type: application/json' \
        "https://rancherdev.io.nuxeo.com/v2-beta/projects/$stackID/services/$serviceID/" | jq '.state')
    echo "$state"    
    sleep 2;
done

curl -u "${CATTLE_ACCESS_KEY}:${CATTLE_SECRET_KEY}" \
    -X POST \
    -H 'Accept: application/json' \
    -H 'Content-Type: application/json' \
    -d '{}' \
    "https://rancherdev.io.nuxeo.com/v2-beta/projects/$stackID/services/$serviceID/?action=finishupgrade"
 