package chaincode

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Certificate struct {
	ID        string `json:"ID"`
	CertHash  string `json:"CertHash"`
	IssueDate string `json:"IssueDate"`
	Issuer    string `json:"Issuer"`
	Owner     string `json:"Owner"`
	IsLocked  bool   `json:"IsLocked"`
}

// مفتاح AES ثابت للاختبار (يجب أن يكون 32 حرفاً لـ AES-256)
const AES_KEY = "12345678901234567890123456789012"

// تحسين دالة فك التشفير لمعالجة الأخطاء بشكل آمن
func decrypt(encryptedData string) (string, error) {
	parts := strings.Split(encryptedData, ":")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid encrypted data format")
	}

	iv, err := hex.DecodeString(parts[0])
	if err != nil { return "", fmt.Errorf("error decoding IV") }
	
	ciphertext, err := hex.DecodeString(parts[1])
	if err != nil { return "", fmt.Errorf("error decoding ciphertext") }
	
	authTag, err := hex.DecodeString(parts[2])
	if err != nil { return "", fmt.Errorf("error decoding auth tag") }

	block, err := aes.NewCipher([]byte(AES_KEY))
	if err != nil { return "", err }

	aesgcm, err := cipher.NewGCM(block)
	if err != nil { return "", err }

	fullCiphertext := append(ciphertext, authTag...)
	plaintext, err := aesgcm.Open(nil, iv, fullCiphertext, nil)
	if err != nil { return "", err }

	return string(plaintext), nil
}

// --- الدوال المعدلة للاختبار ---

func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, owner string, issuer string, issueDate string, encryptedHash string) error {
	// تم تعطيل التحقق من دور Admin مؤقتاً لتسهيل عملية النشر والاختبار
	// err := ctx.GetClientIdentity().AssertAttributeValue("role", "admin")
	// if err != nil { return fmt.Errorf("unauthorized: missing admin role") }

	exists, err := s.CertificateExists(ctx, id)
	if err != nil { return err }
	if exists { return fmt.Errorf("certificate with ID %s already exists", id) }

	certificate := Certificate{
		ID: id, Owner: owner, Issuer: issuer, IssueDate: issueDate, CertHash: encryptedHash, IsLocked: true,
	}

	certBytes, err := json.Marshal(certificate)
	if err != nil { return err }

	return ctx.GetStub().PutState(id, certBytes)
}

func (s *SmartContract) LockCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	certificate, err := s.ReadCertificate(ctx, id)
	if err != nil { return err }

	// تحقق من المالك (يمكن تعطيله أيضاً إذا كنت تواجه مشاكل في التواقيع حالياً)
	clientID, _ := ctx.GetClientIdentity().GetID()
	if certificate.Owner != clientID { 
		return fmt.Errorf("unauthorized: only owner can lock (Current: %s, Owner: %s)", clientID, certificate.Owner) 
	}

	certificate.IsLocked = true
	certBytes, err := json.Marshal(certificate)
	if err != nil { return err }

	return ctx.GetStub().PutState(id, certBytes)
}

func (s *SmartContract) UnlockCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	certificate, err := s.ReadCertificate(ctx, id)
	if err != nil { return err }

	clientID, _ := ctx.GetClientIdentity().GetID()
	if certificate.Owner != clientID { return fmt.Errorf("unauthorized") }

	certificate.IsLocked = false
	certBytes, err := json.Marshal(certificate)
	if err != nil { return err }

	return ctx.GetStub().PutState(id, certBytes)
}

func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface, id string, currentHash string) (bool, error) {
	certificate, err := s.ReadCertificate(ctx, id)
	if err != nil { return false, err }

	if certificate.IsLocked { return false, fmt.Errorf("certificate is locked and cannot be verified") }

	decryptedStoredHash, err := decrypt(certificate.CertHash)
	if err != nil { return false, fmt.Errorf("decryption failed: %v", err) }

	return decryptedStoredHash == currentHash, nil
}

// --- دوال القراءة المساعدة ---

func (s *SmartContract) ReadCertificate(ctx contractapi.TransactionContextInterface, id string) (*Certificate, error) {
	certBytes, err := ctx.GetStub().GetState(id)
	if err != nil { return nil, err }
	if certBytes == nil { return nil, fmt.Errorf("certificate %s not found", id) }

	var certificate Certificate
	err = json.Unmarshal(certBytes, &certificate)
	if err != nil { return nil, err }

	return &certificate, nil
}

func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certBytes, err := ctx.GetStub().GetState(id)
	if err != nil { return false, err }
	return certBytes != nil, nil
}

func (s *SmartContract) GetAllCertificates(ctx contractapi.TransactionContextInterface) ([]*Certificate, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil { return nil, err }
	defer resultsIterator.Close()

	var certificates []*Certificate
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil { return nil, err }

		var certificate Certificate
		err = json.Unmarshal(queryResponse.Value, &certificate)
		if err != nil { return nil, err }
		certificates = append(certificates, &certificate)
	}
	return certificates, nil
}
