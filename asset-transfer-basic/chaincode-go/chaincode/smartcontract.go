package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract defines the structure for our certificate management
type SmartContract struct {
	contractapi.Contract
}

// Certificate describes the details of a digital certificate
// تم ترتيب الحقول أبجدياً لضمان التوافق (Determinism) عند التحويل لـ JSON
type Certificate struct {
	CertHash   string `json:"CertHash"`   // بصمة SHA-256 للملف الأصلي
	ID         string `json:"ID"`         // الرقم التسلسلي الفريد للشهادة
	IssueDate  string `json:"IssueDate"`  // تاريخ الإصدار
	Issuer     string `json:"Issuer"`     // الجهة المصدرة (مثلاً الجامعة)
	Owner      string `json:"Owner"`      // اسم صاحب الشهادة
}

// InitLedger adds a base set of certificates to the ledger (اختياري للتهيئة)
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	certificates := []Certificate{
		{ID: "cert1", Owner: "Ahmed", Issuer: "University A", IssueDate: "2023-01-01", CertHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{ID: "cert2", Owner: "Sara", Issuer: "University B", IssueDate: "2023-02-15", CertHash: "5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9"},
	}

	for _, cert := range certificates {
		certBytes, err := json.Marshal(cert)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(cert.ID, certBytes)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// IssueCertificate creates a new certificate and stores it in the ledger
// هذه الدالة هي المسؤولة عن "حماية" الشهادة عبر تخزين بصمتها الرقمية
func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, owner string, issuer string, issueDate string, certHash string) error {
	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the certificate %s already exists", id)
	}

	certificate := Certificate{
		ID:        id,
		Owner:     owner,
		Issuer:    issuer,
		IssueDate: issueDate,
		CertHash:  certHash, // البصمة الناتجة عن SHA-256
	}
	
	certBytes, err := json.Marshal(certificate)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certBytes)
}

// VerifyCertificate compares a provided hash with the stored hash for a certificate
// الدالة الأساسية لكشف التزوير
func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface, id string, currentHash string) (bool, error) {
	certificate, err := s.ReadCertificate(ctx, id)
	if err != nil {
		return false, err
	}

	// مقارنة البصمة المدخلة مع البصمة المخزنة في البلوكشين
	isValid := certificate.CertHash == currentHash
	return isValid, nil
}

// ReadCertificate returns the certificate stored in the world state with given id
func (s *SmartContract) ReadCertificate(ctx contractapi.TransactionContextInterface, id string) (*Certificate, error) {
	certBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if certBytes == nil {
		return nil, fmt.Errorf("the certificate %s does not exist", id)
	}

	var certificate Certificate
	err = json.Unmarshal(certBytes, &certificate)
	if err != nil {
		return nil, err
	}

	return &certificate, nil
}

// RevokeCertificate deletes a certificate from the ledger
func (s *SmartContract) RevokeCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the certificate %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// CertificateExists returns true when certificate with given ID exists in world state
func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return certBytes != nil, nil
}

// GetAllCertificates returns all certificates found in world state
func (s *SmartContract) GetAllCertificates(ctx contractapi.TransactionContextInterface) ([]*Certificate, error) {
	// range query with empty string for startKey and endKey does an open-ended query of all assets in the chaincode namespace.
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

		var certificate Certificate
		err = json.Unmarshal(queryResponse.Value, &certificate)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, &certificate)
	}

	return certificates, nil
}
