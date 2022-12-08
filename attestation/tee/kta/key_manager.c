/*
kunpengsecl licensed under the Mulan PSL v2.
You can use this software according to the terms and conditions of
the Mulan PSL v2. You may obtain a copy of Mulan PSL v2 at:
    http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.

Author: leezhenxiang
Create: 2022-11-04
Description: key managing module in kta.
	1. 2022-11-04	leezhenxiang
		define the structures.
    2. 2022-12-01   waterh2o
        Implementation function sendrequest and getresponse
*/

#include <tee_defines.h>
#include <kta_common.h>
#include <tee_object_api.h>
#include <tee_crypto_api.h>
//#include <cJSON.h>

#define PARAM_COUNT 4

extern Cache cache;
extern CmdQueue cmdqueue;
extern ReplyQueue replyqueue;
// ===================Communication with kcm====================================

//--------------------------1、SendRequest---------------------------------------------
bool isQueueEmpty(CmdQueue cmdQueue){
    // 1=empty,0=not empty
    if (cmdQueue.head == cmdQueue.tail){
        tlogd("cmdQueue is empty,nothing should be sent.\n");
        return 1;
    }
    return 0;
}

//generate Data encryption symmetric key kcm-pub key
TEE_Result generaterCmdDataKey(CmdRequest *intermediateRequest){
};

void encryption(char *pubkeyname, CmdRequest *intermediateRequest,
                CmdRequest *finalrequest){
    /*
    todo:
    1 encrypt cmd_data by symmetric key
    2 encrypt symmetric key by kcm-pub key
    */
   
};

void generaterFinalRequest(){
    /* get request data from cmdqueue,and generate final request*/
    //申请IntermediateRequest临时内存
    TEE_AllocateTransientObject();
    generaterCmdDataKey();
    json();//cmdQueue.queue[cmdQueue.head]
    encryption();
    json();
    //释放IntermediateRequest
    return;
}

TEE_Result SendRequest(uint32_t param_type, TEE_Param params[PARAM_COUNT]) {
    //todo: send request to ka when ka polls, and answer ta trusted state which ka asks
    TEE_Result ret;
    void *buffer; //the buffer is to be specified
    int queue_empty;
    if (!check_param_type(param_type,
        TEE_PARAM_TYPE_MEMREF_OUTPUT,
        TEE_PARAM_TYPE_NONE,
        TEE_PARAM_TYPE_NONE,
        TEE_PARAM_TYPE_NONE)) {
        tloge("Bad expected parameter types, 0x%x.\n", param_type);
        return TEE_ERROR_BAD_PARAMETERS;
    }
    //Judge whether cmd queue is empty 
    queue_empty = isQueueEmpty(cmdqueue);
    if (!queue_empty){
        return TEE_ERROR_ITEM_NOT_FOUND;
    }
    //generater Request Return value for ka
    CmdRequest finalrequest;
    generaterFinalRequest(finalrequest);
    params[0].memref.buffer = buffer;
    return TEE_SUCCESS;
}

int dequeue(CmdQueue cmdQueue){
    //1=failed ;0=success
    int rtn;
    rtn = isQueueEmpty(cmdqueue);
    if (rtn){
        return rtn;
    }
    cmdQueue.head = (cmdQueue.head + 1) % MAX_QUEUE_SIZE;
    return 0;
};

//--------------------------2、GetResponse---------------------------------------------

TEE_Result GetResponse(uint32_t param_type, TEE_Param params[PARAM_COUNT]) {
    //todo: Get Response from ka when kta had sent request to kcm before
    TEE_Result ret;
    if (!check_param_type(param_type,
        TEE_PARAM_TYPE_MEMREF_INPUT,
        TEE_PARAM_TYPE_NONE,
        TEE_PARAM_TYPE_NONE,
        TEE_PARAM_TYPE_NONE)) {
        tloge("Bad expected parameter types, 0x%x.\n", param_type);
        return TEE_ERROR_BAD_PARAMETERS;
    }
    decryption();
    parsejson();
    switch(cmd);//saveinfo?savetakey?
    //put it to cmd
}

void decryption(){
    /*
    todo:
    1 decrypt symmetric key by kta-priv key
    2 decrypt cmd_data by symmetric key
    */
};


void saveTaInfo(TEE_UUID TA_uuid, char *account, char *password) {
    /*
    todo: options to save ta info in cache,insert the info to cache.ta[?]
    1、search for empty tainfo-node(in empty ta info,head and tail = -1)
    2、save the info in the empty node
    3、modify head next tail etc...
    */

}

void saveTaKey(TEE_UUID TA_uuid, uint32_t keyid, char *keyvalue) {
    //todo: options to save a certain key in cache, Same as the above example
}

// ===================Communication with kcm from ta====================================

bool generateKcmRequest(TEE_Param params[PARAM_COUNT]){
    /* when kta can't complete ta-operation in local kta,
    generate a request and insert it in cmdqueue*/
    // 若队列已满，则无法添加新命令
    if (cmdqueue.head == cmdqueue.tail + 1) {
        tloge("cmd queue is already full");
        return false;
    }
    CmdNode *n = params[0].memref.buffer;
    cmdqueue.queue[cmdqueue.tail] = *n;
    cmdqueue.tail = (cmdqueue.tail + 1) % MAX_TA_NUM;
    return true;
}


// ===================Communication with ta=============================================

//---------------------------InitTAKey--------------------------------------------------

TEE_Result GenerateTAKey(uint32_t param_type, TEE_Param params[PARAM_COUNT]) {
    //TEE_UUID TA_uuid, TEE_UUID masterkey, char *account, char *password
    //todo: new a key for ta
    TEE_Result ret;
    if (!check_param_type(param_type,
        TEE_PARAM_TYPE_MEMREF_INPUT,  
        TEE_PARAM_TYPE_NONE,
        TEE_PARAM_TYPE_VALUE_OUTPUT,  
        TEE_PARAM_TYPE_VALUE_OUTPUT )) {
        tloge("Bad expected parameter types, 0x%x.\n", param_type);
        return TEE_ERROR_BAD_PARAMETERS;
    }
    //params[0].memref.buffer内为输入的cmd结构体
    //params[2]值固定为1
    bool res = generateKcmRequest(params); //生成请求成功或失败的结果存放到params[3]的值中
    if (res) {
        params[3].value.b = 1;
        return TEE_SUCCESS;
    }
    params[3].value.b = 0;
    return TEE_ERROR_OVERFLOW;
}
//---------------------------SearchTAKey------------------------------------------------

void flushCache(TEE_UUID taid, TEE_UUID keyid) {
    /*
    flush the cache according to the LRU algorithm
    support two types of element refresh:
    1.ta sequence;
    2.key sequence;
    */
    int32_t head = cache.head;
    if (!CheckUUID(cache.ta[head].id, taid)) {
        int32_t cur = head;
        int32_t nxt = cache.ta[cur].next;
        while (nxt != -1) {
            if (CheckUUID(cache.ta[nxt].id, taid)) {
                cache.ta[cur].next = cache.ta[nxt].next;
                cache.ta[nxt].next = head;
                cache.head = nxt;
                break;
            }
            cur = nxt;
            nxt = cache.ta[nxt].next;
        }
    }
    TaInfo ta = cache.ta[head];
    head = ta.head;
    if (!CheckUUID(ta.key[head].id, keyid)) {
        int32_t cur = head;
        int32_t nxt = ta.key[cur].next;
        while (nxt != -1) {
            if (CheckUUID(ta.key[nxt].id, keyid)) {
                ta.key[cur].next = ta.key[nxt].next;
                ta.key[nxt].next = head;
                ta.head = nxt;
                break;
            }
            cur = nxt;
            nxt = ta.key[nxt].next;
        }
    }
}

TEE_Result SearchTAKey(uint32_t param_type, TEE_Param params[PARAM_COUNT]) {
    //todo: search a certain ta key, if not exist, call generateKcmRequest(）to generate SearchTAKey request
    TEE_Result ret;
    if (!check_param_type(param_type,
        TEE_PARAM_TYPE_MEMREF_INPUT,  
        TEE_PARAM_TYPE_MEMREF_OUTPUT,
        TEE_PARAM_TYPE_VALUE_OUTPUT,  
        TEE_PARAM_TYPE_VALUE_OUTPUT )) {
        tloge("Bad expected parameter types, 0x%x.\n", param_type);
        return TEE_ERROR_BAD_PARAMETERS;
    }
    //params[0].memref.buffer内为输入的cmd结构体
    CmdNode *n = params[0].memref.buffer;
    int32_t cur = cache.head;
    while (cur != -1) {
        if (CheckUUID(cache.ta[cur].id, n->taId)) {
            TaInfo ta = cache.ta[cur];
            int32_t idx = ta.head;
            while (idx != -1) {
                if (CheckUUID(ta.key[idx].id, n->keyId)) {
                    params[1].memref.size = sizeof(ta.key[idx].value);
                    params[1].memref.buffer = ta.key[idx].value;
                    params[2].value.a = 0;
                    // 更新cache
                    flushCache(n->taId, n->keyId);
                    return TEE_SUCCESS;
                }
                idx = ta.key[idx].next;
            }
        }
        cur = cache.ta[cur].next;
    }
    params[2].value.a = 1;
    bool res = generateKcmRequest(params);
    if (res) {
        params[3].value.b = 1;
        return TEE_SUCCESS;
    }
    params[3].value.b = 0;
    return TEE_ERROR_OVERFLOW;
}

//---------------------------DeleteTAKey------------------------------------------------

TEE_Result DeleteTAKey(uint32_t param_type, TEE_Param params[PARAM_COUNT]) {
    //todo: delete a certain key in the cache
    TEE_Result ret;
    if (!check_param_type(param_type,
        TEE_PARAM_TYPE_MEMREF_INPUT,  
        TEE_PARAM_TYPE_MEMREF_OUTPUT,
        TEE_PARAM_TYPE_VALUE_OUTPUT,  
        TEE_PARAM_TYPE_NONE )) {
        tloge("Bad expected parameter types, 0x%x.\n", param_type);
        return TEE_ERROR_BAD_PARAMETERS;
    }
    //params[0].memref.buffer内为输入的cmd结构体
}

//----------------------------DestoryKey------------------------------------------------

TEE_Result DestoryKey(uint32_t param_type, TEE_Param params[PARAM_COUNT]) {
    //todo: delete a certain key by calling DeleteTAKey(), then generate a delete key request in TaCache
    TEE_Result ret;
    if (!check_param_type(param_type,
        TEE_PARAM_TYPE_MEMREF_INPUT,  
        TEE_PARAM_TYPE_MEMREF_OUTPUT,
        TEE_PARAM_TYPE_VALUE_OUTPUT,  
        TEE_PARAM_TYPE_VALUE_OUTPUT)) {
        tloge("Bad expected parameter types, 0x%x.\n", param_type);
        return TEE_ERROR_BAD_PARAMETERS;
    }
    //params[0].memref.buffer内为输入的cmd结构体
}

//----------------------------GetKcmReply------------------------------------------------

TEE_Result GetKcmReply(uint32_t param_type, TEE_Param params[PARAM_COUNT]){
    TEE_Result ret;
    if (!check_param_type(param_type,
        TEE_PARAM_TYPE_MEMREF_INPUT,  
        TEE_PARAM_TYPE_MEMREF_OUTPUT,
        TEE_PARAM_TYPE_NONE,  
        TEE_PARAM_TYPE_NONE)) {
        tloge("Bad expected parameter types, 0x%x.\n", param_type);
        return TEE_ERROR_BAD_PARAMETERS;
    }
    //params[0].memref.buffer内为输入的cmd结构体
    if (replyqueue.head == replyqueue.tail) {
        tloge("get kcm reply error: reply queue is empty\n");
        return TEE_ERROR_ITEM_NOT_FOUND;
    }
    int32_t first = replyqueue.head;
    params[1].memref.size = sizeof(replyqueue.queue[first]);
    params[1].memref.buffer = (void*)malloc(params[1].memref.size);
    params[1].memref.buffer = &replyqueue.queue[first];
    replyqueue.head = (replyqueue.head + 1) % MAX_QUEUE_SIZE;
    return TEE_SUCCESS;
}

//----------------------------ClearCache------------------------------------------------

TEE_Result ClearCache(uint32_t param_type, TEE_Param params[PARAM_COUNT]) {
    //todo: clear all ta cache
    TEE_Result ret;
    if (!check_param_type(param_type,
        TEE_PARAM_TYPE_MEMREF_INPUT,  
        TEE_PARAM_TYPE_VALUE_OUTPUT,
        TEE_PARAM_TYPE_NONE,  
        TEE_PARAM_TYPE_NONE)) {
        tloge("Bad expected parameter types, 0x%x.\n", param_type);
        return TEE_ERROR_BAD_PARAMETERS;
    }
    //params[0].memref.buffer内为输入的cmd结构体
    CmdNode *n = params[0].memref.buffer;
    // 验证帐号密码
    if (!verifyTApasswd(n->taId, n->account, n->password)) {
        params[1].value.b = 0;
        return TEE_ERROR_ACCESS_DENIED;
    }

    // cache仅1个元素且命中
    if (CheckUUID(cache.ta[cache.head].id, n->taId) && cache.head == cache.tail) {
        cache.head = END_NULL;
        cache.tail = END_NULL;
        tloge("clear ta cache succeeded.\n");
        params[1].value.b = 1;
        return TEE_SUCCESS;
    }

    // cache仅1个元素且未命中
    if (!CheckUUID(cache.ta[cache.head].id, n->taId) && cache.head == cache.tail) {
        tloge("ta cache not fount.\n");
        params[1].value.b = 0;
        return TEE_ERROR_ITEM_NOT_FOUND;
    }

    // cache有2个或以上元素
    int32_t cur = cache.head;
    if (CheckUUID(cache.ta[cur].id, n->taId)) {
        cache.head = cache.ta[cur].next;
        tloge("clear ta cache succeeded.\n");
        params[1].value.b = 1;
        return TEE_SUCCESS;
    }
    int32_t nxt = cache.ta[cur].next;
    while (nxt != END_NULL) {
        TEE_UUID tmp = cache.ta[nxt].id;
        if (CheckUUID(tmp, n->taId)) {
            cache.ta[cur].next = cache.ta[nxt].next;
            if (nxt == cache.tail) {
                cache.tail = cur;
            }
            tloge("clear ta cache succeeded.\n");
            params[1].value.b = 1;
            return TEE_SUCCESS;
        }
        cur = nxt;
        nxt = cache.ta[nxt].next;
    }
    tloge("ta cache not found.\n");
    params[1].value.b = 0;
    return TEE_ERROR_ITEM_NOT_FOUND;
}