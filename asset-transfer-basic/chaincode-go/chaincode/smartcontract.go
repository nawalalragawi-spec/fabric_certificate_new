package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// Certificate تعريف هيكل الشهادة
type Certificate struct {
	CertHash    string `json:"CertHash"`    // بصمة ملف الشهادة لمنع التزوير
	Degree      string `json:"Degree"`      // التخصص أو الدرجة
	ID          string `json:"ID"`          // الرقم التسلسلي للشهادة
	IsRevoked   bool   `json:"IsRevoked"`   // حالة الشهادة (ملغية أم لا)
	IssueDate   string `json:"IssueDate"`   // تاريخ الصدور
	Issuer      string `json:"Issuer"`      // الجهة المانحة للشهادة
	StudentName string `json:"StudentName"` // اسم الطالب
}

// 1. IssueCertificate: إصدار شهادة جديدة
func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, studentName string, degree string, issuer string, certHash string, issueDate string) error {
	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("الشهادة ذات الرقم %s موجودة مسبقاً", id)
	}

	cert := Certificate{
		ID:          id,
		StudentName: studentName,
		Degree:      degree,
		Issuer:      issuer,
		CertHash:    certHash,
		IssueDate:   issueDate,
		IsRevoked:   false, // الشهادة فعالة عند الإصدار
	}
	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certJSON)
}

// 2. QueryAllCertificates: استعلام عن جميع الشهادات المخزنة
func (s *SmartContract) QueryAllCertificates(ctx contractapi.TransactionContextInterface) ([]*Certificate, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var certificates []*Certificate
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var cert Certificate
		err = json.Unmarshal(queryResponse.Value, &cert)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, &cert)
	}

	return certificates, nil
}

// 3. RevokeCertificate: إلغاء شهادة (في حال التزوير أو الخطأ)
func (s *SmartContract) RevokeCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	cert, err := s.ReadCertificate(ctx, id)
	if err != nil {
		return err
	}

	cert.IsRevoked = true // تغيير الحالة إلى ملغية
	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certJSON)
}

// 4. VerifyCertificate: التحقق من صحة الشهادة وصلاحيتها
func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface, id string, certHash string) (bool, error) {
	cert, err := s.ReadCertificate(ctx, id)
	if err != nil {
		return false, fmt.Errorf("الشهادة غير موجودة في السجلات")
	}

	// التأكد من أن البصمة مطابقة وأن الشهادة ليست ملغية
	if cert.CertHash == certHash && !cert.IsRevoked {
		return true, nil
	}

	return false, nil
}

// --- وظائف مساعدة ---

func (s *SmartContract) ReadCertificate(ctx contractapi.TransactionContextInterface, id string) (*Certificate, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, err
	}
	if certJSON == nil {
		return nil, fmt.Errorf("الشهادة %s غير موجودة", id)
	}

	var cert Certificate
	err = json.Unmarshal(certJSON, &cert)
	return &cert, err
}

func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, err
	}
	return certJSON != nil, nil
}
