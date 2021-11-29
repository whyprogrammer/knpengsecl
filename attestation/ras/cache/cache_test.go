package cache

import (
	"strconv"
	"testing"
	"time"

	"gitee.com/openeuler/kunpengsecl/attestation/ras/config"
)

func TestCacheCommand(t *testing.T) {
	c := &Cache{command: 0}

	c.SetCMDSendConfigure()
	if (c.command & CMDSENDCONF) != CMDSENDCONF {
		t.Errorf("test cache command error at send configure\n")
	}

	c.ClearCommands()

	c.SetCMDGetTrustReport()
	if (c.command & CMDGETREPORT) != CMDGETREPORT {
		t.Errorf("test cache command error at get trust report\n")
	}
}

func TestCacheUpdate(t *testing.T) {
	cfg := config.GetDefault(config.ConfServer)
	c := &Cache{
		command:         0,
		hbExpiration:    time.Now(),
		trustExpiration: time.Now(),
	}
	testCases1 := []struct {
		inputHB time.Duration
		inputT  time.Duration
		result  bool
		cmd     uint64
	}{
		{10 * time.Second, 20 * time.Second, false, 0},
		{2000000, 20 * time.Second, false, 0},
		{10, 20, true, CMDGETREPORT}, // default nano-second unit.
	}
	for i := 0; i < len(testCases1); i++ {
		c.ClearCommands()
		cfg.SetTrustDuration(testCases1[i].inputT)
		c.UpdateTrustReport()
		cfg.SetHBDuration(testCases1[i].inputHB)
		c.UpdateHeartBeat()
		if c.IsHeartBeatExpired() != testCases1[i].result {
			t.Errorf("test cache update error at case %d, expiration wrong\n", i)
		}
		if c.command != testCases1[i].cmd {
			t.Errorf("test cache update error at case %d, command wrong\n", i)
		}
	}
}

func TestCacheTrust(t *testing.T) {
	cfg := config.GetDefault(config.ConfServer)
	c := &Cache{
		command:         0,
		hbExpiration:    time.Now(),
		trustExpiration: time.Now(),
	}
	testCases1 := []struct {
		inputT     time.Duration
		inputDelay time.Duration
		result     bool
		cmd        uint64
	}{
		{time.Second, time.Second, false, CMDGETREPORT},
		{2 * time.Second, time.Second, true, CMDGETREPORT},
		{3 * time.Second, time.Second, true, 0},
	}
	for i := 0; i < len(testCases1); i++ {
		c.ClearCommands()
		cfg.SetTrustDuration(testCases1[i].inputT)
		c.UpdateTrustReport()
		time.Sleep(testCases1[i].inputDelay)
		if c.IsReportValid() != testCases1[i].result {
			t.Errorf("test cache trust error at case %d, expiration wrong\n", i)
		}
		if c.command != testCases1[i].cmd {
			t.Errorf("test cache trust error at case %d, command wrong\n", i)
		}
	}
}

func TestNonce(t *testing.T) {
	c := &Cache{}
	for i := 0; i < 10; i++ {
		nonce, err := c.CreateNonce()
		if err != nil {
			t.FailNow()
		}
		t.Log(len(strconv.FormatUint(nonce, 2)))
	}

}
