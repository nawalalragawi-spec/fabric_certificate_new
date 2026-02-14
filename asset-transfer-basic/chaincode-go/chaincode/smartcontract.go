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

// SmartContract defines the structure for our certificate management
type SmartContract struct {
	contractapi.Contract
}

// Certificate describes the details of a digital certificate
type Certificate struct {
	ID        string `json:"ID"`
	CertHash  string `json:"CertHash"`  // البيانات المشفرة (AES Ciphertext)
	IssueDate string `json:"IssueDate"`
	Issuer    string `json:"Issuer"`
	Owner     string `json:"Owner"`
	IsLocked  bool   `json:"IsLocked"`  // إضافة حقل القفل لمطابقة بروتوكول عمر سعد
}

const AES_KEY = "12345678901234567890123456789012"

// دالة مساعدة لفك التشفير باستخدام AES-GCM
func decrypt(encryptedData string) (string, error) {
	parts := strings.Split(encryptedData, ":")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid encrypted data format")
	}

	iv, _ := hex.DecodeString(parts[0])
	ciphertext, _ := hex.DecodeString(parts[1])
	authTag, _ := hex.DecodeString(parts[2])

	block, err := aes.NewCipher([]byte(AES_KEY))
	if err != nil { return "", err }

	aesgcm, err := cipher.NewGCM(block)
	if err != nil { return "", err }

	fullCiphertext := append(ciphertext, authTag...)
	plaintext, err := aesgcm.Open(nil, iv, fullCiphertext, nil)
	if err != nil { return "", err }

	return string(plaintext), nil
}

// 1. IssueCertificate: التحقق من دور admin وضبط الحالة على Locked تلقائياً
func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, owner string, issuer string, issueDate string, encryptedHash string) error {
	
	// التحقق من سمة الدور (ABAC) لمطابقة ورقة عمر سعد
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "admin")
	if err != nil {
		return fmt.Errorf("unauthorized: only admins can issue certificates")
	}

	exists, err := s.CertificateExists(ctx, id)
	if err != nil { return err }
	if exists { return fmt.Errorf("the certificate %s already exists", id) }

	certificate := Certificate{
		ID:        id,
		Owner:     owner,
		Issuer:    issuer,
		IssueDate: issueDate,
		CertHash:  encryptedHash,
		IsLocked:  true, // الحالة الافتراضية مقفلة لتعزيز الخصوصية
	}

	certBytes, err := json.Marshal(certificate)
	if err != nil { return err }

	return ctx.GetStub().PutState(id, certBytes)
}

// 2. LockCertificate: السماح للمالك فقط بإعادة قفل الشهادة
func (s *SmartContract) LockCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	certificate, err := s.ReadCertificate(ctx, id)
	if err != nil { return err }

	// التحقق من أن المستدعي هو المالك (Self-Sovereignty)
	clientID, _ := ctx.GetClientIdentity().GetID()
	if certificate.Owner != clientID {
		return fmt.Errorf("unauthorized: only the owner can lock the certificate")
	}

	certificate.IsLocked = true
	certBytes, _ := json.Marshal(certificate)
	return ctx.GetStub().PutState(id, certBytes)
}

// 3. UnlockCertificate: السماح للمالك فقط بفتح الشهادة للتحقق
func (s *SmartContract) UnlockCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	certificate, err := s.ReadCertificate(ctx, id)
	if err != nil { return err }

	clientID, _ := ctx.GetClientIdentity().GetID()
	if certificate.Owner != clientID {
		return fmt.Errorf("unauthorized: only the owner can unlock the certificate")
	}

	certificate.IsLocked = false
	certBytes, _ := json.Marshal(certificate)
	return ctx.GetStub().PutState(id, certBytes)
}

// 4. VerifyCertificate: منع التحقق إذا كانت الشهادة مقفلة
func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface, id string, currentHash string) (bool, error) {
	certificate, err := s.ReadCertificate(ctx, id)
	if err != nil { return false, err }

	// إذا كانت الشهادة مقفلة، نرفض المعاملة (جوهر التحكم في الوصول)
	if certificate.IsLocked {
		return false, fmt.Errorf("access denied: certificate is locked by the student")
	}

	decryptedStoredHash, err := decrypt(certificate.CertHash)
	if err != nil { return false, fmt.Errorf("failed to decrypt: %v", err) }

	return decryptedStoredHash == currentHash, nil
}

// الدوال المساعدة (Read, Exists, Revoke, GetAll)
func (s *SmartContract) ReadCertificate(ctx contractapi.TransactionContextInterface, id string) (*Certificate, error) {
	certBytes, err := ctx.GetStub().GetState(id)
	if err != nil || certBytes == nil {
		return nil, fmt.Errorf("certificate %s does not exist", id)
	}
	var certificate Certificate
	err = json.Unmarshal(certBytes, &certificate)
	return &certificate, err
}

func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certBytes, err := ctx.GetStub().GetState(id)
	return certBytes != nil, nil
}

func (s *SmartContract) RevokeCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	// التحقق من دور الأدمن عند الإلغاء أيضاً
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "admin")
	if err != nil { return fmt.Errorf("unauthorized") }

	exists, err := s.CertificateExists(ctx, id)
	if err != nil || !exists { return fmt.Errorf("not found") }
	return ctx.GetStub().DelState(id)
}

func (s *SmartContract) GetAllCertificates(ctx contractapi.TransactionContextInterface) ([]*Certificate, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil { return nil, err }
	defer resultsIterator.Close()

	var certificates []*Certificate
	for resultsIterator.HasNext() {
		queryResponse, _ := resultsIterator.Next()
		var certificate Certificate
		json.Unmarshal(queryResponse.Value, &certificate)
		certificates = append(certificates, &certificate)
	}
	return certificates, nil
}
