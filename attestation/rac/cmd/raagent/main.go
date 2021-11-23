package main

import (
	"encoding/json"
	"log"
	"time"

	"gitee.com/openeuler/kunpengsecl/attestation/rac/ractools"
	"gitee.com/openeuler/kunpengsecl/attestation/ras/cache"
	"gitee.com/openeuler/kunpengsecl/attestation/ras/clientapi"
	"gitee.com/openeuler/kunpengsecl/attestation/ras/config"
)

func main() {
	const addr string = "127.0.0.1:40001"
	// step 1. get configuration from local file, clientId, hbDuration, Cert, etc.
	cid := config.GetDefault().GetClientId()
	tpm, err := ractools.OpenTPM(true)
	if err != nil {
		log.Printf("OpenTPM failed, error: %s \n", err)
	}
	//the input is not be used now
	//TODO: add tpm config file
	tpm.Prepare(&ractools.TPMConfig{})

	// step 2. if rac doesn't have clientId, it uses Cert to do the register process.
	if cid < 0 {
		_, err := tpm.GetEKCert()
		if err != nil {
			log.Printf("GetEkCert failed, error: %s \n", err)
		}
		req := clientapi.CreateIKCertRequest{
			// EkCert: ekCert,
			// IkPub:  tpm.GetIKPub(),
			EkCert: ractools.CertPEM,
			IkPub:  ractools.PubPEM,
			IkName: tpm.GetIKName(),
		}
		bkk, err := clientapi.DoCreateIKCert(addr, &req)
		if err != nil {
			log.Fatal("Client:can't Create IkCert")
		}
		eic := bkk.GetIcEncrypted()
		log.Println("Client:get encryptedIC=", eic)

		ci, err := json.Marshal(map[string]string{"test name": "test value"})
		bk, err := clientapi.DoRegisterClient(addr, &clientapi.RegisterClientRequest{
			Ic:         &clientapi.Cert{Cert: []byte{1, 2}},
			ClientInfo: &clientapi.ClientInfo{ClientInfo: string(ci)},
		})
		if err != nil {
			log.Fatal("Client: can't register rac!")
		}
		cid = bk.GetClientId()
		config.GetDefault().SetClientId(cid)
		config.GetDefault().SetHBDuration(time.Duration((int64)(time.Second) * bk.GetClientConfig().HbDurationSeconds))
		log.Printf("Client: get clientId=%d", cid)
		config.Save()
	}

	// step 3. if rac has clientId, it uses clientId to send heart beat.
	for {
		rpy, err := clientapi.DoSendHeartbeat(addr, &clientapi.SendHeartbeatRequest{ClientId: cid})
		if err != nil {
			log.Fatalf("Client: send heart beat error %v", err)
		}
		log.Printf("Client: get heart beat back %v", rpy.GetNextAction())

		// step 4. do what ras tells to do by NextAction...
		DoNextAction(tpm, addr, cid, rpy)

		// step 5. what else??

		// step n. sleep and wait.
		time.Sleep(config.GetDefault().GetHBDuration())
	}
}

// DoNextAction checks the nextAction field and invoke the corresponding handler function.
func DoNextAction(tpm *ractools.TPM, srv string, id int64, rpy *clientapi.SendHeartbeatReply) {
	action := rpy.GetNextAction()
	if (action & cache.CMDSENDCONF) == cache.CMDSENDCONF {
		SetNewConf(srv, id, rpy)
	}
	if (action & cache.CMDGETREPORT) == cache.CMDGETREPORT {
		SendTrustReport(tpm, srv, id, rpy)
	}
	// add new command handler functions here.
}

// SetNewConf sets the new configuration values from RAS.
func SetNewConf(srv string, id int64, rpy *clientapi.SendHeartbeatReply) {
	log.Printf("Client: get new configuration from RAS.")
	config.GetDefault().SetHBDuration(time.Duration(rpy.GetActionParameters().GetClientConfig().HbDurationSeconds))
	config.GetDefault().SetTrustDuration(time.Duration(rpy.GetActionParameters().GetClientConfig().TrustDurationSeconds))
}

// SendTrustReport sneds a new trust report to RAS.
func SendTrustReport(tpm *ractools.TPM, srv string, id int64, rpy *clientapi.SendHeartbeatReply) {
	tRep, err := tpm.GetTrustReport(rpy.GetActionParameters().GetNonce(), id)
	if err != nil {
		log.Printf("Client: create a new trust report failed :%v", err)
	} else {
		log.Printf("Client: create a new trust report success")
	}
	var manifest []*clientapi.Manifest
	for _, m := range tRep.Manifest {
		manifest = append(manifest, &clientapi.Manifest{Type: m.Type, Item: m.Content})
	}
	/*
		srr, _ := clientapi.DoSendReport(srv, &clientapi.SendReportRequest{
			ClientId: id,
			TrustReport: &clientapi.TrustReport{
				PcrInfo: &clientapi.PcrInfo{
					Algorithm: tRep.PcrInfo.AlgName,
					PcrValues: (map[int32]string)(tRep.PcrInfo.Values),
					PcrQuote: &clientapi.PcrQuote{
						Quoted:    tRep.PcrInfo.Quote.Quoted,
						Signature: tRep.PcrInfo.Quote.Signature,
					},
				},
				ClientId: tRep.ClientID,
				ClientInfo: &clientapi.ClientInfo{
					ClientInfo: tRep.ClientInfo,
				},
				Manifest: manifest,
			},
		})
		if srr.Result {
			log.Printf("Client: send a new trust report to RAS ok.")
		} else {
			log.Printf("Client: send a new trust report to RAS failed.")
		}
	*/
}
