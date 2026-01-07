package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract لإدارة الشهادات الرقمية
type SmartContract struct {
	contractapi.Contract
}

// Certificate يمثل هيكل الشهادة الرقمية
// تم ترتيب الحقول أبجدياً لضمان التوافق (Determinism)
type Certificate struct {
	CertHash      string `json:"CertHash"`      // البصمة الرقمية للشهادة (SHA-256)
	Degree        string `json:"Degree"`        // الدرجة العلمية (مثلاً: بكالوريوس هندسة)
	ID            string `json:"ID"`            // الرقم التسلسلي الفريد للشهادة
	Issuer        string `json:"Issuer"`        // الجهة المصدرة (الجامعة)
	IssueDate     string `json:"IssueDate"`     // تاريخ الإصدار
	StudentName   string `json:"StudentName"`   // اسم الطالب
	Verified      bool   `json:"Verified"`      // حالة التحقق
}

// InitLedger إدخال بيانات أولية (شهادات تجريبية)
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	certificates := []Certificate{
		{ID: "CERT001", StudentName: "Ahmed Ali", Degree: "Computer Science", Issuer: "University A", CertHash: "a591a...f321", IssueDate: "2023-01-01", Verified: true},
		{ID: "CERT002", StudentName: "Sara Omar", Degree: "Cybersecurity", Issuer: "University B", CertHash: "b212b...e110", IssueDate: "2023-05-15", Verified: true},
	}

	for _, cert := range certificates {
		certJSON, err := json.Marshal(cert)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(cert.ID, certJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// IssueCertificate إصدار شهادة جديدة وإضافتها للبلوكشين
func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, studentName string, degree string, issuer string, certHash string, issueDate string) error {
	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the certificate %s already exists", id)
	}

	cert := Certificate{
		ID:          id,
		StudentName: studentName,
		Degree:      degree,
		Issuer:      issuer,
		CertHash:    certHash,
		IssueDate:   issueDate,
		Verified:    true,
	}
	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certJSON)
}

// ReadCertificate قراءة بيانات شهادة من البلوكشين للتحقق منها
func (s *SmartContract) ReadCertificate(ctx contractapi.TransactionContextInterface, id string) (*Certificate, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if certJSON == nil {
		return nil, fmt.Errorf("the certificate %s does not exist", id)
	}

	var cert Certificate
	err = json.Unmarshal(certJSON, &cert)
	if err != nil {
		return nil, err
	}

	return &cert, nil
}

// CertificateExists وظيفة مساعدة للتأكد من وجود الشهادة
func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return certJSON != nil, nil
}
