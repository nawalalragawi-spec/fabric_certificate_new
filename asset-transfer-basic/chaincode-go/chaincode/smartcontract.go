package chaincode

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract defines the structure for our chaincode
type SmartContract struct {
	contractapi.Contract
}

// encryptionKey: مفتاح التشفير (يجب أن يكون 32 بايت بالضبط لـ AES-256)
// ملاحظة بحثية: في بيئة الإنتاج يتم تمرير المفتاح عبر الـ Transient field لزيادة الأمان
var encryptionKey = []byte("asupersecretkeythatis32byteslong")

// Certificate: هيكل الشهادة الأكاديمية
type Certificate struct {
	ID          string `json:"ID"`
	StudentName string `json:"StudentName"` // سيخزن كنص مشفر Base64
	Degree      string `json:"Degree"`
	IssueDate   string `json:"IssueDate"`
	Issuer      string `json:"Issuer"`
}

// =========================================================================================
// وظائف التشفير المساعدة (AES-256 GCM)
// =========================================================================================

// encrypt: دالة لتشفير البيانات وتحويلها إلى Base64
func encrypt(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt: دالة لفك التشفير واسترجاع النص الأصلي
func decrypt(cryptoText string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeToString(cryptoText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// =========================================================================================
// وظائف العقد الذكي الأساسية
// =========================================================================================

// IssueCertificate: إنشاء شهادة جديدة مع تشفير بيانات الطالب
func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, studentName string, degree string, issueDate string, issuer string) error {
	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the certificate %s already exists", id)
	}

	// تشفير اسم الطالب قبل حفظه في قاعدة البيانات (CouchDB)
	encryptedName, err := encrypt([]byte(studentName), encryptionKey)
	if err != nil {
		return fmt.Errorf("encryption failed: %v", err)
	}

	cert := Certificate{
		ID:          id,
		StudentName: encryptedName,
		Degree:      degree,
		IssueDate:   issueDate,
		Issuer:      issuer,
	}

	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certJSON)
}

// ReadCertificate: قراءة الشهادة وفك تشفير اسم الطالب للعرض
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

	// فك تشفير الاسم قبل إرجاع النتيجة للمستخدم
	decryptedName, err := decrypt(cert.StudentName, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %v", err)
	}
	cert.StudentName = decryptedName

	return &cert, nil
}

// CertificateExists: التحقق من وجود الشهادة
func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, err
	}
	return certJSON != nil, nil
} هل الكود صحيح
