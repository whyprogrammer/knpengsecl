package trustmgr

import (
	"testing"

	"gitee.com/openeuler/kunpengsecl/attestation/ras/entity"
)

type testValidator struct {
}

func (tv *testValidator) Validate(report *entity.Report) error {
	return nil
}

func TestRecordReport(t *testing.T) {
	vm := new(testValidator)
	SetValidator(vm)

	pcrInfo := entity.PcrInfo{
		AlgName: "sha1",
		Values: map[int]string{
			1: "pcr value 1",
			2: "pcr value 2",
		},
		Quote: []byte("test quote"),
	}

	biosItem1 := entity.ManifestItem{
		Name:   "test bios name1",
		Value:  "test bios value1",
		Detail: "test bios detail1",
	}

	biosItem2 := entity.ManifestItem{
		Name:   "test bios name2",
		Value:  "test bios value2",
		Detail: "test bios detail2",
	}

	imaItem1 := entity.ManifestItem{
		Name:   "test ima name1",
		Value:  "test ima value1",
		Detail: "test ima detail1",
	}

	biosManifest := entity.Manifest{
		Type: "bios",
		Items: []entity.ManifestItem{
			0: biosItem1,
			1: biosItem2,
		},
	}

	imaManifest := entity.Manifest{
		Type: "ima",
		Items: []entity.ManifestItem{
			0: imaItem1,
		},
	}

	testReport := &entity.Report{
		PcrInfo: pcrInfo,
		Manifest: []entity.Manifest{
			0: biosManifest,
			1: imaManifest,
		},
		ClientID: 1,
		ClientInfo: entity.ClientInfo{
			Info: map[string]string{
				"client_name":        "test_client",
				"client_type":        "test_type",
				"client_description": "test description",
			},
		},
	}

	RecordReport(testReport)
}
