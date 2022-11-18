#!/bin.bash
### this scripts should be run under the root folder of kunpengsecl project
PROJROOT=.

# include common part
. ${PROJROOT}/attestation/tas/test/common.sh
echo "=========="
echo "start ${CASENAME} at: $(date)" | tee -a ${DST}/control.txt
echo "prepare the test environments..." | tee -a ${DST}/control.txt

echo "start akservice..." | tee -a ${DST}/control.txt
( cd ${DST}/tas ; ./akservice -T &>${DST}/tas/echo.txt ; ./akservice &>>${DST}/tas/echo.txt ; )&
echo "wait for 3s..." | tee -a ${DST}/control.txt
sleep 3

if [ ! -f "${DAACERT}" ]; then
    echo "daa-ac.crt is not exist" | tee -a ${DST}/control.txt
    echo "test continue..." | tee -a ${DST}/control.txt
else
    echo "daa-ac.crt already exists" | tee -a ${DST}/control.txt
    echo "remove daa-ac.crt" | tee -a ${DST}/control.txt
    rm -rf ${DAACERT}
    echo "test continue..." | tee -a ${DST}/control.txt
fi

echo "start qca demo..." | tee -a ${DST}/control.txt
( cd ${DST}/qca ; ./qcaserver -C 2 &>${DST}/qca/echo.txt ; )&
echo "wait for 3s..." | tee -a ${DST}/control.txt
sleep 3

echo "kill all processes..." | tee -a ${DST}/control.txt
pkill -u ${USER} akservice
pkill -u ${USER} qcaserver
echo "wait for 3s..." | tee -a ${DST}/control.txt
sleep 3

S=$(cat ${DST}/tas/echo.txt | grep "Verify ak signature & QCA ok")
if [ ! -z "${S}" ]; then
    echo "QTA measurement have been compared" | tee -a ${DST}/control.txt
else
    echo "QTA measurement have not been compared" | tee -a ${DST}/control.txt
    echo "test failed!" | tee -a ${DST}/control.txt
    exit 1
fi
echo "wait for 3s..." | tee -a ${DST}/control.txt
sleep 3

if [ -f "${DAACERT}" ]; then
    echo "daa-ac.crt generated successed!" | tee -a ${DST}/control.txt
    echo "test succeeded!" | tee -a ${DST}/control.txt
    exit 0
else
    echo "daa-ac.crt generated failed!" | tee -a ${DST}/control.txt
    echo "test failed!" | tee -a ${DST}/control.txt
    exit 1
fi
