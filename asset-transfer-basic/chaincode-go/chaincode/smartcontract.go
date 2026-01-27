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
	CertHash   string `json:"CertHash"`   // ستخزن هنا البيانات المشفرة (AES Ciphertext)
	ID         string `json:"ID"`         
	IssueDate  string `json:"IssueDate"`  
	Issuer     string `json:"Issuer"`     
	Owner      string `json:"Owner"`      
}

// مفتاح AES ثابت (يجب أن يتكون من 32 رمزاً لـ AES-256) ويطابق الموجود في ملفات Caliper
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
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// دمج ciphertext مع authTag لأن مكتبة Go تتوقعهما معاً
	fullCiphertext := append(ciphertext, authTag...)
	plaintext, err := aesgcm.Open(nil, iv, fullCiphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// IssueCertificate تخزن الشهادة (الهاش يصل مشفراً من طرف العميل)
func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, owner string, issuer string, issueDate string, encryptedHash string) error {
	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the certificate %s already exists", id)
	}

	certificate := Certificate{
		ID:         id,
		Owner:      owner,
		Issuer:     issuer,
		IssueDate:  issueDate,
		CertHash:   encryptedHash, // يتم التخزين وهو مشفر
	}

	certBytes, err := json.Marshal(certificate)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certBytes)
}

// VerifyCertificate تفك تشفير الهاش المخزن وتقارنه بالهاش المدخل
func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface, id string, currentHash string) (bool, error) {
	certificate, err := s.ReadCertificate(ctx, id)
	if err != nil {
		return false, err
	}

	// فك التشفير عن البيانات المخزنة
	decryptedStoredHash, err := decrypt(certificate.CertHash)
	if err != nil {
		return false, fmt.Errorf("failed to decrypt stored hash: %v", err)
	}

	// مقارنة الهاش الصريح (Plaintext) مع الهاش المقدم للتحقق
	isValid := decryptedStoredHash == currentHash
	return isValid, nil
}

// ReadCertificate لاسترجاع الشهادة كما هي (مشفرة)
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

// CertificateExists للتحقق من وجود المعرف
func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return certBytes != nil, nil
}

// RevokeCertificate لحذف الشهادة
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

// GetAllCertificates لاسترجاع قائمة بكافة الشهادات
func (s *SmartContract) GetAllCertificates(ctx contractapi.TransactionContextInterface) ([]*Certificate, error) {
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
