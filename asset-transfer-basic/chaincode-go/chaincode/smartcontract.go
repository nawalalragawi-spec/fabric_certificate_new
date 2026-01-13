package chaincode

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json" // تم إضافتها لأنك قد تحتاجها لاحقاً رغم عدم استخدامها في الدوال المساعدة حالياً
	"fmt"
	"io"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract defines the structure for our chaincode
type SmartContract struct {
	contractapi.Contract
}

// encryptionKey: مفتاح التشفير (يجب أن يكون 32 بايت بالضبط لـ AES-256)
// ملاحظة: في الإنتاج يفضل استخدام Transient data لتمرير المفاتيح بدلاً من كتابتها في الكود
var encryptionKey = []byte("asupersecretkeythatis32byteslong")

// Certificate: هيكل الشهادة الأكاديمية
// تم تصحيح 'ype' إلى 'type' وإضافة الـ backticks للـ json tags
type Certificate struct {
	ID          string `json:"ID"`
	StudentName string `json:"StudentName"`
	Degree      string `json:"Degree"`
	IssueDate   string `json:"IssueDate"`
	Issuer      string `json:"Issuer"`
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

	// يتم دمج الـ nonce مع النص المشفر لكي نستخدمه لاحقاً في فك التشفير
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt: دالة لفك التشفير واسترجاع النص الأصلي
// تم إكمال الكود وتصحيح دالة فك ترميز Base64
func decrypt(cryptoText string, key []byte) (string, error) {
	// تصحيح: الدالة هي DecodeString وليست DecodeToString
	data, err := base64.StdEncoding.DecodeString(cryptoText)
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

	// فصل الـ nonce عن النص المشفر الفعلي
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	
	// فك التشفير
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
