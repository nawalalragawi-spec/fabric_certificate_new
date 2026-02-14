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
	CertHash  string `json:"CertHash"`  // الهاش المشفر بـ AES
	IssueDate string `json:"IssueDate"`
	Issuer    string `json:"Issuer"`
	Owner     string `json:"Owner"`     // سيتم تخزين المعرف الرقمي للمالك هنا
	IsLocked  bool   `json:"IsLocked"`  // حالة القفل (قلب البروتوكول)
}

const AES_KEY = "12345678901234567890123456789012"

// دالة مساعدة لفك التشفير باستخدام AES-GCM
func decrypt(encryptedData string) (string, error) {
	parts := strings.Split(encryptedData, ":")
	if len(parts) != 3 {
		return "", fmt.Errorf("صيغة التشفير غير صحيحة")
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

// 1. IssueCertificate: التحقق من دور admin وضبط الحالة Locked تلقائياً
// هذا هو التعديل المطلوب لضمان تخزين المالك والتحقق من الصلاحيات
func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, owner string, issuer string, issueDate string, encryptedHash string) error {
	
	// أ) التحقق من سمة "admin" (ABAC) لضمان أن الجامعة فقط من تصدر الشهادات
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "admin")
	if err != nil {
		return fmt.Errorf("غير مصرح: يجب أن تملك دور admin لتتمكن من إصدار شهادة")
	}

	// ب) التأكد من عدم تكرار المعرف
	exists, err := s.CertificateExists(ctx, id)
	if err != nil { return err }
	if exists { return fmt.Errorf("الشهادة بالمعرف %s موجودة مسبقاً", id) }

	// ج) بناء الشهادة: ضبط المالك (Owner) والحالة Locked (مقفولة) افتراضياً
	certificate := Certificate{
		ID:        id,
		Owner:     owner,         // تخزين معرف المالك المار من Caliper
		Issuer:    issuer,
		IssueDate: issueDate,
		CertHash:  encryptedHash, // الهاش القادم مشفراً من العميل
		IsLocked:  true,          // "بروتوكول عمر": الخصوصية تبدأ بالقفل التلقائي
	}

	certBytes, err := json.Marshal(certificate)
	if err != nil { return err }

	return ctx.GetStub().PutState(id, certBytes)
}

// 2. LockCertificate: المالك فقط يغلق شهادته
func (s *SmartContract) LockCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	certificate, err := s.ReadCertificate(ctx, id)
	if err != nil { return err }

	clientID, _ := ctx.GetClientIdentity().GetID()
	if certificate.Owner != clientID {
		return fmt.Errorf("غير مصرح: المالك الفعلي فقط من يمكنه قفل الشهادة")
	}

	certificate.IsLocked = true
	certBytes, _ := json.Marshal(certificate)
	return ctx.GetStub().PutState(id, certBytes)
}

// 3. UnlockCertificate: المالك فقط يفتح الشهادة
func (s *SmartContract) UnlockCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	certificate, err := s.ReadCertificate(ctx, id)
	if err != nil { return err }

	clientID, _ := ctx.GetClientIdentity().GetID()
	if certificate.Owner != clientID {
		return fmt.Errorf("غير مصرح: المالك الفعلي فقط من يمكنه فتح الشهادة")
	}

	certificate.IsLocked = false
	certBytes, _ := json.Marshal(certificate)
	return ctx.GetStub().PutState(id, certBytes)
}

// 4. VerifyCertificate: يمنع التحقق إذا كانت الشهادة مقفلة
func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface, id string, currentHash string) (bool, error) {
	certificate, err := s.ReadCertificate(ctx, id)
	if err != nil { return false, err }

	// منطق الأمان: إذا لم يفتح المالك الشهادة، نرفض محاولة التحقق تماماً
	if certificate.IsLocked {
		return false, fmt.Errorf("مرفوض: الشهادة مقفلة من قبل الطالب ولا يمكن التحقق منها حالياً")
	}

	decryptedStoredHash, err := decrypt(certificate.CertHash)
	if err != nil { return false, fmt.Errorf("فشل فك التشفير: %v", err) }

	return decryptedStoredHash == currentHash, nil
}

// الدوال المساعدة الأساسية
func (s *SmartContract) ReadCertificate(ctx contractapi.TransactionContextInterface, id string) (*Certificate, error) {
	certBytes, err := ctx.GetStub().GetState(id)
	if err != nil || certBytes == nil {
		return nil, fmt.Errorf("الشهادة %s غير موجودة", id)
	}
	var certificate Certificate
	json.Unmarshal(certBytes, &certificate)
	return &certificate, nil
}

func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certBytes, err := ctx.GetStub().GetState(id)
	return certBytes != nil, nil
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
