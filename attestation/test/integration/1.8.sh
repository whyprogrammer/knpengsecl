#!/bin/bash
# this scripts should be run under the root folder of kunpengsecl project
#set -eux
PROJROOT=.
# run number of rac clients to test
NUM=1
# include common part
. ${PROJROOT}/attestation/test/integration/common.sh
. ${PROJROOT}/attestation/test/integration/1.x.common.sh

# above are common preparation steps, below are specific preparation step, scope includs:
# configure files, input files, environment variables, cmdline paramenters, flow control paramenters, etc.
### Start Preparation
echo "start test preparation..." | tee -a ${DST}/control.txt
pushd $(pwd)
cd ${PROJROOT}/attestation/quick-scripts
echo "clean database" | tee -a ${DST}/control.txt
sh clear-database.sh | tee -a ${DST}/control.txt
popd
### End Preparation

### start launching binaries for testing
### Here we launch in https mode
echo "start ras..." | tee -a ${DST}/control.txt
( cd ${DST}/ras ; ./ras -T &>${DST}/ras/echo.txt ; ./ras -v &>>${DST}/ras/echo.txt ;)&
echo "wait for 5s..." | tee -a ${DST}/control.txt
sleep 5
AUTHTOKEN=$(grep "Bearer " ${DST}/ras/echo.txt)

# start number of rac clients
echo "start ${NUM} rac clients..." | tee -a ${DST}/control.txt
(( count=0 ))
for (( i=1; i<=${NUM}; i++ ))
do
    ( cd ${DST}/rac-${i} ; ${DST}/rac/raagent -t -v &>${DST}/rac-${i}/echo.txt ; )&
    (( count++ ))
    if (( count >= 1 ))
    then
        (( count=0 ))
        echo "start ${i} rac clients at $(date)..." | tee -a ${DST}/control.txt
    fi
done
echo "wait for 5s" | tee -a ${DST}/control.txt
sleep 5

### start monitoring and control the testing
echo "start to perform test ..." | tee -a ${DST}/control.txt
echo "check config items via restapi request"
RESPONSE1=$(curl -k -H "Content-Type: application/json" https://localhost:40003/config)

# modify the ras config items
echo "start modifying ras config..." | tee -a ${DST}/control.txt
curl -X POST -k -H "Authorization: $AUTHTOKEN" -H "Content-Type: application/json" https://localhost:40003/config --data "{\"isallupdate\":${newALLUPDATE},\"logtestmode\":${newLOGTESTMODE},\"mgrstrategy\":\"${newMGRSTRATEGY}\",\"extractrules\":\"${newEXTRACTRULES}\"}"
RESPONSE3=$(curl -k -H "Content-Type: application/json" https://localhost:40003/config)
### analyse the ras testing data
MGRSTRATEGY1=$(echo $RESPONSE1 | jq -r '.' | awk '/mgrstrategy/ {gsub(",","",$2);print $2}')
MGRSTRATEGY2=$(echo $RESPONSE3 | jq -r '.' | awk '/mgrstrategy/ {gsub(",","",$2);print $2}')
EXTRACTRULES1=$(echo $RESPONSE1 | jq -r '.' | awk '/extractrules/ {print $2}')
EXTRACTRULES2=$(echo $RESPONSE3 | jq -r '.' | awk '/extractrules/ {print $2}')
AUTOUPDATE1=$(echo $RESPONSE1 | jq -r '.' | awk '/isallupdate/ {gsub(",","",$2);print $2}')
AUTOUPDATE2=$(echo $RESPONSE3 | jq -r '.' | awk '/isallupdate/ {gsub(",","",$2);print $2}')
LOGTESTMODE1=$(echo $RESPONSE1 | jq -r '.' | awk '/logtestmode/ {gsub(",","",$2);print $2}')
LOGTESTMODE2=$(echo $RESPONSE3 | jq -r '.' | awk '/logtestmode/ {gsub(",","",$2);print $2}')
echo "First time: mgrStrategy:${MGRSTRATEGY1}, extractRules:${EXTRACTRULES1}, autoUpdateConfig:${AUTOUPDATE1}, logTestMode:${LOGTESTMODE1}" | tee -a ${DST}/control.txt
echo "Second time: mgrStrategy:${MGRSTRATEGY2}, extractRules:${EXTRACTRULES2}, autoUpdateConfig:${AUTOUPDATE2}, logTestMode:${LOGTESTMODE2}" | tee -a ${DST}/control.txt
if [[ "${MGRSTRATEGY2}" == "\"${newMGRSTRATEGY}\"" && "${EXTRACTRULES2}" == "\"${newEXTRACTRULES}\"" &&  "${AUTOUPDATE2}" == "${newALLUPDATE}" && "${LOGTESTMODE2}" == "${newLOGTESTMODE}" ]]
then
    echo "ras test succeeded!" | tee -a ${DST}/control.txt
    echo "start recovering ras config..." | tee -a ${DST}/control.txt
    curl -X POST -k -H "Authorization: $AUTHTOKEN" -H "Content-Type: application/json" https://localhost:40003/config --data "{\"isallupdate\":${oriALLUPDATE},\"logtestmode\":${oriLOGTESTMODE},\"mgrstrategy\":\"${oriMGRSTRATEGY}\",\"extractrules\":\"${oriEXTRACTRULES}\"}"
    sleep 5
    echo "ras config recover succeeded!" | tee -a ${DST}/control.txt
else
    echo "ras test failed!" | tee -a ${DST}/control.txt
    echo "kill all test processes..." | tee -a ${DST}/control.txt
    pkill -u ${USER} ras
    pkill -u ${USER} raagent
    echo "test DONE!!!" | tee -a ${DST}/control.txt
    exit 1
fi

# modify the rac config items
echo "start modifying rac config..." | tee -a ${DST}/control.txt
curl -X POST -k -H "Authorization: $AUTHTOKEN" -H "Content-Type: application/json" https://localhost:40003/config --data "{\"hbduration\":\"${newHBDURATION}\",\"trustduration\":\"${newTRUSTDURATION}\",\"DigestAlgorithm\":\"${newDIGESTALG}\"}"
RESPONSE4=$(curl -k -H "Content-Type: application/json" https://localhost:40003/config)
### analyse the rac testing data
HBDURATION1=$(echo $RESPONSE1 | jq -r '.' | awk '/hbduration/ {gsub(",","",$2);print $2}')
HBDURATION2=$(echo $RESPONSE4 | jq -r '.' | awk '/hbduration/ {gsub(",","",$2);print $2}')
TRUSTDURATION1=$(echo $RESPONSE1 | jq -r '.' | awk '/trustduration/ {gsub(",","",$2);print $2}')
TRUSTDURATION2=$(echo $RESPONSE4 | jq -r '.' | awk '/trustduration/ {gsub(",","",$2);print $2}')
DIGESTALG1=$(echo $RESPONSE1 | jq -r '.' | awk '/digestalgorithm/ {gsub(",","",$2);print $2}')
DIGESTALG2=$(echo $RESPONSE4 | jq -r '.' | awk '/digestalgorithm/ {gsub(",","",$2);print $2}')
echo "First time: hbDuration:${HBDURATION1}, trustDuration:${TRUSTDURATION1}, digestAlgorithm:${DIGESTALG1}" | tee -a ${DST}/control.txt
echo "Second time: hbDuration:${HBDURATION2}, trustDuration:${TRUSTDURATION2}, digestAlgorithm:${DIGESTALG2}" | tee -a ${DST}/control.txt
if [[ "${HBDURATION2}" == "\"${newHBDURATION}\"" && "${TRUSTDURATION2}" == "\"${newTRUSTDURATION}\"" && "${DIGESTALG2}" == "\"${newDIGESTALG}\"" ]]
then
    echo "rac test succeeded!" | tee -a ${DST}/control.txt
    echo "start recovering rac config..." | tee -a ${DST}/control.txt
    curl -X POST -k -H "Authorization: $AUTHTOKEN" -H "Content-Type: application/json" https://localhost:40003/config --data "{\"hbduration\":\"${oriHBDURATION}\",\"trustduration\":\"${oriTRUSTDURATION}\",\"DigestAlgorithm\":\"${oriDIGESTALG}\"}"
    sleep 5
    echo "rac config recover succeeded!" | tee -a ${DST}/control.txt
else
    echo "rac test failed!" | tee -a ${DST}/control.txt
    echo "kill all test processes..." | tee -a ${DST}/control.txt
    pkill -u ${USER} ras
    pkill -u ${USER} raagent
    echo "test DONE!!!" | tee -a ${DST}/control.txt
    exit 1
fi

### stop test
echo "all test succeeded!" | tee -a ${DST}/control.txt
echo "kill all test processes..." | tee -a ${DST}/control.txt
pkill -u ${USER} ras
pkill -u ${USER} raagent
echo "test DONE!!!" | tee -a ${DST}/control.txt
exit 0
